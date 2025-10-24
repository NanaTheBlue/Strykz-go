package auth

import (
	"context"
	"log"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionToken, err := r.Cookie("auth_token")

		csrfToken, errr := r.Cookie("csrf_token")
		csrfTokenHeader := r.Header.Get("X-CSRF-TOKEN")
		type contextKey string
		const userContextKey = contextKey("user")

		if err != nil || errr != nil {
			log.Println(err)
			http.Error(w, "cookie not found", http.StatusBadRequest)
			return
		}

		err = validateCSRF(csrfToken.Value, csrfTokenHeader)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		user, err := validateJWT(sessionToken.Value)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)

		log.Println(user.Username)

		next(w, r.WithContext(ctx))
	}
}
