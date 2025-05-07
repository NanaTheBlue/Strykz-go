package social

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strykz/auth"
	"strykz/db"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/*
	type party struct {
		party    string
		senderId string
	}
*/

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	UserID  string `json:"userID"`
	Message string `json:"message"`
}

type Client struct {
	UserID string
	Conn   *websocket.Conn
}

//plan is to send a message to all the clients on join events and leave events

// also plan to just send the whole list of users in the map to the users

//todo setup ping pong

var onlineUsers sync.Map

func SetOnlineStatus(s db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// very important that i change this line later
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		user, ok := r.Context().Value(auth.UserKey).(auth.User)

		if !ok {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Client Connected to Websocket")
		value, err := json.Marshal(user.UserName)
		if err != nil {
			return
		}
		s.Add(context.Background(), user.UserID, value, 60)

		onlineUsers.Store(user.UserID, &Client{
			UserID: user.UserID,
			Conn:   ws,
		})
		//"notifications"
		go SubscribeToChannel(r.Context(), s)

		checkNotifications(r.Context())

		reader(s, user.UserID, ws)

	}

}

func PartyInvite(s db.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}

		u, ok := r.Context().Value(auth.UserKey).(auth.User)
		if !ok {
			http.Error(w, "Username not found", http.StatusInternalServerError)
			return
		}

		username := r.FormValue("username")

		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		var recipientID string
		// Should think about having the users ids accesible when doing a party invite so i dont have to query for a username
		err := db.Pool.QueryRow(r.Context(),
			"SELECT id FROM users WHERE username = $1", username).Scan(&recipientID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "User lookup failed: %v\n", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		party := "PartyInvite"
		id := uuid.New().String()

		//Instead of inserting the party invite here can prob just put it in a redis set

		fmt.Println(id)
		//Here we need to put notifications into a struct
		var notifications []Notification

		notifications = append(notifications, Notification{
			Notification_id:   id,
			Sender_id:         u.UserID,
			Recipient_id:      recipientID,
			Notification_type: party,
		})

		msgJson, err := json.Marshal(notifications)
		if err != nil {
			return
		}
		s.Add(r.Context(), id, msgJson, 300)
		s.Publish(r.Context(), u.UserID, msgJson)

		//to DO need to get Notification IDS so we can Query it Fast No Cap

	}

}

func acceptNotification(s db.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}
		var notification Notification
		err := json.NewDecoder(r.Body).Decode(&notification)
		if err != nil {
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}
		// prob gonna change this up here a bit later but atm we just deal with party invites in the API
		if notification.Notification_id == "PartyInvite" {
			handlePartyInvite(notification)
			// do thing
		} else {
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}

		return
	}

}
