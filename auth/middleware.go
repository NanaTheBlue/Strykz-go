package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strykz/db"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

/*
	var middleware = []func(http.HandlerFunc) http.HandlerFunc{
		//authMiddleware,
	}
*/
func getIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Error parsing IP: %v", err)
		return ""
	}
	return host
}

var ipLimiterMap sync.Map

func Rate(next http.HandlerFunc, limit rate.Limit, burst int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		limiterAny, _ := ipLimiterMap.LoadOrStore(ip, rate.NewLimiter(limit, burst))
		limiter := limiterAny.(*rate.Limiter)

		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "To many requests"})
			return
		}
		next(w, r)
	}
}

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
		const useridKey = contextKey("userID")

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

		var userID string
		var expires time.Time

		// i should prob increase the tokens expiration date if its within a 24 hours of expiration date.
		er := db.Pool.QueryRow(context.Background(), "SELECT id,expires_at FROM users WHERE session_token = $1;", sessionToken).Scan(&userID, &expires)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Select Failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if time.Now().After(expires) {
			// wanna return something to the front end here so it can redirect the user to the login page
			return
		}
		// putting the username in context so that we we can use it with the next request avoiding another database query
		ctx := context.WithValue(r.Context(), "userID", userID)

		next(w, r.WithContext(ctx))
	}
}
