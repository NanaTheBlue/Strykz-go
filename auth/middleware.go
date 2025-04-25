package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strykz/db"
	"sync"
	"time"
)

/*
	var middleware = []func(http.HandlerFunc) http.HandlerFunc{
		//authMiddleware,
	}
*/

type User struct {
	UserID         string
	UserName       string
	ProfilePicture string
	Expires        time.Time
	Elo            int
}

var ipLimiterMap sync.Map

/*
	func recoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					msg := "Caught panic: %v, Stack trace: %s"
					log.Printf(msg, err, string(debug.Stack()))

					er := http.StatusInternalServerError
					http.Error(w, "Internal Server Error", er)
				}
			}()
			next(w, r)
		}
	}
*/

type contextKey string

const UserKey = contextKey("user")

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionToken, errr := r.Cookie("session_token")
		csrf_token, err := r.Cookie("CSRF_Token")
		if err != nil || errr != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csrf := r.Header.Get("X-CSRF-Token")
		if csrf != csrf_token.Value || csrf == "" {
			println("bingo bongo no csrfo")
			return
		}

		var u User

		// i should prob increase the tokens expiration date if its within a 24 hours of expiration date.
		er := db.Pool.QueryRow(context.Background(), "SELECT id,expires_at,username FROM users WHERE session_token = $1;", sessionToken).Scan(&u.UserID, &u.Expires, &u.UserName)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Select Failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if time.Now().After(u.Expires) {
			// wanna return something to the front end here so it can redirect the user to the login page
			return
		}
		// putting the user struct  in context so that we we can use it with the next request avoiding another database query
		ctx := context.WithValue(r.Context(), UserKey, u)

		next(w, r.WithContext(ctx))
	}
}

func Cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func WSAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// pass csrf token as string query paramater to reduce risk of cross site websocket hijacking
		sessionToken, err := r.Cookie("session_token")
		fmt.Println("Session Token:", sessionToken.Value)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		//fmt.Println("We In The Route")
		var u User

		// i should prob increase the tokens expiration date if its within a 24 hours of expiration date.
		er := db.Pool.QueryRow(context.Background(), "SELECT id, expires_at, username FROM users WHERE session_token = $1;", sessionToken.Value).Scan(&u.UserID, &u.Expires, &u.UserName)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Select Failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if time.Now().After(u.Expires) {
			// wanna return something to the front end here so it can redirect the user to the login page
			return
		}
		// putting the user struct  in context so that we we can use it with the next request avoiding another database query
		ctx := context.WithValue(r.Context(), UserKey, u)

		next(w, r.WithContext(ctx))
	}
}
