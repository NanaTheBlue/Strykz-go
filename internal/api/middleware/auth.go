package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/nanagoboiler/internal/services/auth"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionToken, err := r.Cookie("auth_token")

		csrfToken, errr := r.Cookie("csrf_token")

		csrfTokenHeader := r.Header.Get("X-CSRF-TOKEN")

		if err != nil || errr != nil {
			log.Println(err)
			http.Error(w, "cookie not found", http.StatusBadRequest)
			return
		}

		err = auth.ValidateCSRF(csrfToken.Value, csrfTokenHeader)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		user, err := auth.ValidateJWT(sessionToken.Value)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserContextKey, user)

		log.Println(user.Username)

		next(w, r.WithContext(ctx))
	}
}
