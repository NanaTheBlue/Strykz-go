package auth

import (
	"context"
	"errors"
	"net/http"
	"strykz/db"
)

var AuthError = errors.New("Unauthorized")

func Authorize(r *http.Request) error {
	//username := r.FormValue("username")
	var session_token string
	var csrf_token string
	st, err := r.Cookie("session_token")
	// check refresh token aswell should also make sure they arent expired

	user := db.Pool.QueryRow(context.Background(), "SELECT session_token FROM users WHERE session_token = $1;", st.Value).Scan(&session_token)
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
