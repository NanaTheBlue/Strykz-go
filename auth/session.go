package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

var AuthError = errors.New("Unauthorized")

func Authorize(conn *pgxpool.Pool, r *http.Request) error {
	//username := r.FormValue("username")
	var session_token string
	var csrf_token string
	st, err := r.Cookie("session_token")

	user := conn.QueryRow(context.Background(), "SELECT session_token, csrf_token FROM users WHERE session_token = $1;", st.Value).Scan(&session_token, &csrf_token)
	if user != nil {
		return AuthError
	}
	if err != nil || st.Value == "" || st.Value != session_token {
		return AuthError
	}

	//Get the Session Token from the cookie

	csrf := r.Header.Get("X-CSRF-Token")
	if csrf != csrf_token || csrf == "" {
		println("bingo bongo no csrfo")
		return AuthError
	}

	return nil

}
