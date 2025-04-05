package auth

import (
	"context"
	"errors"
	"net/http"
	"strykz/db"
)

var AuthError = errors.New("Unauthorized")

func Authorize(r *http.Request) error {

	var session_token string
	csrf_token, erro := r.Cookie("csrf_token")
	st, err := r.Cookie("session_token")

	user := db.Pool.QueryRow(context.Background(), "SELECT session_token FROM users WHERE session_token = $1;", st.Value).Scan(&session_token)
	if user != nil {
		return AuthError
	}
	if err != nil || erro != nil || st.Value == "" || st.Value != session_token {
		return AuthError
	}

	csrf := r.Header.Get("X-CSRF-Token")
	if csrf != csrf_token.Value || csrf == "" {
		println("bingo bongo no csrfo")
		return AuthError
	}

	return nil

}
