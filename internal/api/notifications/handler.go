package notificationsapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/internal/services/notifications"
	"github.com/nanagoboiler/internal/services/social"
	"github.com/nanagoboiler/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Notifications(s notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
			return
		}
		s.AddConnection(user.ID, conn)
		defer s.RemoveConnection(user.ID)

		notifications, err := s.GetNotifications(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "Failed to Retrieve Notifications", http.StatusBadRequest)
			return
		}
		marshalled, err := json.Marshal(notifications)
		if err != nil {
			http.Error(w, "Failed to Marshal json", http.StatusBadRequest)
			return
		}
		conn.WriteJSON(marshalled)

	}

}

func AcceptNotification(s social.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req models.Notification
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		/*
			err := s.
			if err != nil {
				http.Error(w, "failed to accept notification", http.StatusInternalServerError)
				return
			}
		*/

		w.WriteHeader(http.StatusOK)

	}
}

func RejectNotification(s social.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

	}
}
func BlockUser(s notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.BlockRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid Request Json", http.StatusBadRequest)
			return
		}

	}
}
