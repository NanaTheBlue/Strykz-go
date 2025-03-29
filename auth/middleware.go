package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strykz/db"
	"time"
)

var middleware = []func(http.HandlerFunc) http.HandlerFunc{
	authMiddleware,
	recoveryMiddleware,
}

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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionToken, err := r.Cookie("session_token")
		csrf_token := r.Header.Get("CSRF-Token")
		csrf := r.Header.Get("X-CSRF-Token")
		if csrf != csrf_token || csrf == "" {
			println("bingo bongo no csrfo")
			return
		}

		var username string
		var expires time.Time

		// i should prob increase the tokens expiration date if its within a 24 hours of expiration date.
		er := db.Pool.QueryRow(context.Background(), "SELECT username,expires_at FROM users WHERE session_token = $1;", sessionToken).Scan(&username, &expires)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Select Failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if time.Now().After(expires) {
			// wanna return something to the front end here so it can redirect the user to the login page
			return
		}

		next(w, r)
	}
}

func welcomeHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintln(w, "hello, welcome to my website!")
		next(w, r)
	}

}
