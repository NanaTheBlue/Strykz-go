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
		origin := r.Header.Get("Origin")
		switch origin {
		case "https://strykz.net":
			return true
		default:
			return false
		}
	},
}

func Notifications(s notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		s.AddConnection(user.ID, conn)
		defer s.RemoveConnection(user.ID)

		notifications, err := s.GetNotifications(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "Failed to Retrieve Notifications", http.StatusBadRequest)
			return
		}

		if err := conn.WriteJSON(notifications); err != nil {
			return
		}

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}

	}

}

func RejectNotification(s social.Service) http.HandlerFunc {
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

	}
}
