package socialapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/internal/services/social"
	"github.com/nanagoboiler/models"
)

func BlockUser(s social.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req models.BlockRequest

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if decoder.More() {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		err := s.BlockUser(r.Context(), user.ID, req.BlockedID)
		if err != nil {
			http.Error(w, "Failed to BlockUser", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
func AcceptFriendRequest(s social.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}

		idStr := parts[3]
		reqID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "invalid request id", http.StatusBadRequest)
			return
		}

		if err := s.AcceptFriendRequest(r.Context(), user.ID, reqID.String()); err != nil {
			http.Error(w, "failed to accept friend request", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
