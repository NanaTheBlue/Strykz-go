package social

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strykz/auth"
	"strykz/db"
	"sync"

	//"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/config"
	//"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/websocket"
	//"github.com/joho/godotenv"
)

/*
	type party struct {
		party    string
		senderId string
	}
*/

var (
	bucketName string
	s3Region   string
	uploadDir  string
	s3Client   *s3.Client
)

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
		s.Add(context.Background(), user.UserID, user.UserName, 60)

		onlineUsers.Store(user.UserID, &Client{
			UserID: user.UserID,
			Conn:   ws,
		})
		CheckNotifications(r.Context())

		go reader(s, user.UserID, ws)

	}

}

func PartyInvite() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			er := http.StatusMethodNotAllowed
			http.Error(w, "Invalid method", er)
			return
		}
		//reFactor this dont work since i migrated to using a struct
		senderId, ok := r.Context().Value("user").(string)
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

		err := db.Pool.QueryRow(context.Background(),
			"SELECT id FROM users WHERE username = $1", username).Scan(&recipientID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "User lookup failed: %v\n", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		party := "PartyInvite"

		_, error := db.Pool.Exec(context.Background(), "INSERT INTO notifications (recipient_id, sender_id, type ) VALUES ($1, $2, $3);", recipientID, senderId, party)
		if error != nil {
			fmt.Fprintf(os.Stderr, "Insert failed: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

}

func ChangeProfilePicture() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var imgBytes []byte
		var err error
		imgBytes, err = validateImage(w, r)
		if err != nil {
			er := http.StatusNotAcceptable
			http.Error(w, "File Too Big Gang", er)
			return

		}
		err = os.WriteFile("profile.jpg", imgBytes, 0644)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(os.Stderr, "Bing Bong ")

	}
}
