package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strykz/db"
	"time"
)

func Register() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")
		if len(username) < 4 || len(password) < 8 || len(email) < 5 {
			er := http.StatusNotAcceptable
			http.Error(w, "Invalid username/password or email", er)
			return
		}

		hashedPassword, _ := hashPassword(password)
		_, err := db.Pool.Exec(context.Background(), "INSERT INTO users (username, hashed_password, email ) VALUES ($1, $2, $3);", username, hashedPassword, email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Insert failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "User registered Successfully")
	}

}

func Login() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		var hashed_password string

		err := db.Pool.QueryRow(context.Background(), "SELECT hashed_password FROM users WHERE username = $1;", username).Scan(&hashed_password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Select Failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !checkPasswordHash(password, hashed_password) {
			er := http.StatusUnauthorized
			http.Error(w, "invalid username or password", er)
			return
		}
		sessionToken := generateToken(32)
		csrfToken := generateToken(32)

		_, insertError := db.Pool.Exec(context.Background(), "UPDATE users SET session_token = $1 WHERE username = $2;", sessionToken, username)
		if insertError != nil {
			fmt.Fprintf(os.Stderr, "Token Update failed: %v\n", insertError)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "Login Success")

		// set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  time.Now().Add(time.Hour * 24 * 14),
			HttpOnly: true,
		})

		// set CSRF token in a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Expires:  time.Now().Add(time.Hour * 24 * 14),
			HttpOnly: false,
		})

	}

}

func Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := Authorize(r); err != nil {
			er := http.StatusUnauthorized
			http.Error(w, "Unauthorized", er)
			return
		}

		// Clear cookie
		sessionToken, err := r.Cookie("session_token")
		if err != nil || sessionToken.Value == "" {
			return
		}

		_, updateError := db.Pool.Exec(context.Background(), "UPDATE users SET session_token = null, expires_at = null, WHERE session_token = $1;", sessionToken)
		if updateError != nil {
			fmt.Fprintf(os.Stderr, "Token Update failed: %v\n", updateError)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HttpOnly: true,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HttpOnly: false,
		})

		// clear token from database

		// should also set refresh token to "" and clear it from database

		fmt.Fprintln(w, "Logged out Success")
	}

}

/*
func protected(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}

		if err := Authorize(r); err != nil {
			er := http.StatusUnauthorized
			http.Error(w, "Unauthorized", er)
			return
		}

		fmt.Fprintf(w, "CSRF validation succesful! Welcome, ")
	}


func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}

		if err := Authorize(r); err != nil {
			er := http.StatusUnauthorized
			http.Error(w, "Unauthorized", er)
			return
		}

		fmt.Fprintf(w, "CSRF validation succesful! Welcome, ")
		next(w, r)
	}
}
*/
