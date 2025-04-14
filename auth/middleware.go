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

type user struct {
	userID         string
	userName       string
	profilePicture string
	expires        time.Time
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

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type contextKey string
		const userKey = contextKey("user")

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

		var u user

		// i should prob increase the tokens expiration date if its within a 24 hours of expiration date.
		er := db.Pool.QueryRow(context.Background(), "SELECT id,expires_at,username FROM users WHERE session_token = $1;", sessionToken).Scan(&u.userID, &u.expires, &u.userName)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Select Failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if time.Now().After(u.expires) {
			// wanna return something to the front end here so it can redirect the user to the login page
			return
		}
		// putting the user struct  in context so that we we can use it with the next request avoiding another database query
		ctx := context.WithValue(r.Context(), userKey, u)

		next(w, r.WithContext(ctx))
	}
}
