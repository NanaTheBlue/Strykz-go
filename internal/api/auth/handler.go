package authapi

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nanagoboiler/internal/auth"
	"github.com/nanagoboiler/models"
)

func Register(s auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		csrf, err := uuid.NewRandom()
		if err != nil {
			return
		}
		now := time.Now()

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid Request Json", http.StatusBadRequest)
			return
		}

		err = validateRegistration(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tokens, err := s.RegisterUser(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		setCookie(w, "auth_token", tokens.Auth_token, now.Add(10*time.Minute), http.SameSiteNoneMode, true, true)
		setCookie(w, "refresh_token", tokens.Refresh_token, now.Add(24*30*time.Hour), http.SameSiteNoneMode, true, true)
		setCookie(w, "csrf_token", csrf.String(), now.Add(10*time.Minute), http.SameSiteNoneMode, true, false)

		w.WriteHeader(http.StatusCreated)

	}

}

func Renew(s auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := r.Cookie("refresh_token")
		now := time.Now()

		if err != nil {
			log.Println(err)
			http.Error(w, "cookie not found", http.StatusBadRequest)
			return
		}
		csrf, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
			return
		}

		tokens, err := s.RenewToken(r.Context(), refreshToken.String())

		if err != nil {
			http.Error(w, "Failed to Renew Tokens", http.StatusInternalServerError)
		}

		setCookie(w, "auth_token", tokens.Auth_token, now.Add(10*time.Minute), http.SameSiteNoneMode, true, true)
		setCookie(w, "csrf_token", csrf.String(), now.Add(10*time.Minute), http.SameSiteNoneMode, true, false)

		w.WriteHeader(http.StatusCreated)

	}

}

func Login(s auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		csrf, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
			return
		}
		now := time.Now()

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid Request JSON", http.StatusBadRequest)
			return
		}

		tokens, err := s.LoginUser(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		setCookie(w, "auth_token", tokens.Auth_token, now.Add(10*time.Minute), http.SameSiteNoneMode, true, true)
		setCookie(w, "refresh_token", tokens.Refresh_token, now.Add(24*30*time.Hour), http.SameSiteNoneMode, true, true)
		setCookie(w, "csrf_token", csrf.String(), now.Add(10*time.Minute), http.SameSiteNoneMode, true, false)

		w.WriteHeader(http.StatusCreated)

	}

}

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	}
}
