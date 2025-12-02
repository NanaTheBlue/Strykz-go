package notificationsapi

// Would like to eventually rip this apart as its own micro service

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/internal/services/notifications"
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
		defer conn.Close()

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

func AcceptNotification(s notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func RejectNotification(s notifications.Service) http.HandlerFunc {
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
