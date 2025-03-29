package social

import (
	"context"
	"fmt"
	"os"

	"net/http"
	"strykz/db"
)

/*
type party struct {
	party    string
	senderId string
}
*/

func PartyInvite() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}
		username := r.FormValue("username")

		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		var recipientID string

		err := db.Pool.QueryRow(context.Background(),
			"SELECT id FROM users WHERE username = $1", username).Scan(&recipientID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "User lookup failed: %v\n", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		party := "PartyInvite"

		var senderId = "c1f4232b-a3df-4a1a-a3c5-0515ff90aaf5"

		_, error := db.Pool.Exec(context.Background(), "INSERT INTO notifications (recipient_id, sender_id, type ) VALUES ($1, $2, $3);", recipientID, senderId, party)
		if error != nil {
			fmt.Fprintf(os.Stderr, "Insert failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

}
