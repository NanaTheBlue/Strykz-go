package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	_ "image/jpeg"
	_ "image/png"

	"log"

	"net/http"
	"strykz/auth"
	"strykz/db"
	"time"

	"github.com/gorilla/websocket"
)

type Notification struct {
	Notification_id   string `json:"notification_id"`
	Sender_id         string `json:"sender"`
	Recipient_id      string `json:"recipient"`
	Notification_type string `json:"notification_type"`
}

func heartBeat(p []byte) bool {
	if string(p) == "Pong" {
		return true
	}
	return false
}

func SubscribeToChannel(ctx context.Context, s db.Store) {
	u, ok := ctx.Value(auth.UserKey).(auth.User)
	if !ok {
		fmt.Println("user not found in context")
		return
	}

	s.Subscribe(context.Background(), "onlineUsers", func(message string) {
		msgJSON, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshalling message to JSON: %v", err)
			return
		}

		sendNotifications(u.UserID, msgJSON)
	})

}

func checkNotifications(ctx context.Context) error {
	u, ok := ctx.Value(auth.UserKey).(auth.User)
	if !ok {
		fmt.Println("user not found in context")
		return nil
	}

	var notifications []Notification

	// should prob get the notification ID and put that in the struct can use it to map over the mf aswell
	fmt.Println(u.UserID)
	rows, err := db.Pool.Query(context.Background(), "SELECT sender_id, type, notification_id FROM notifications WHERE recipient_id = $1", u.UserID)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var sender_id string
		var notification_type string
		var notification_id string
		err = rows.Scan(&sender_id, &notification_type, &notification_id)
		if err != nil {
			return err
		}
		notifications = append(notifications, Notification{
			Notification_id:   notification_id,
			Sender_id:         sender_id,
			Notification_type: notification_type,
		})

	}

	if rows.Err() != nil {
		return err
	}

	if len(notifications) == 0 {
		return nil
	}
	//fmt.Printf("%+v\n", notifications)

	msgJSON, err := json.Marshal(notifications)
	if err != nil {
		log.Printf("Error marshalling message to JSON: %v", err)
		return err
	}

	if sendNotifications(u.UserID, msgJSON) != nil {
		failed := errors.New("Error Sending to Client")
		return failed
	}

	return nil

}

func sendNotifications(userID string, message []byte) error {

	value, ok := onlineUsers.Load(userID)
	if !ok {
		failed := errors.New("User Likely not in the map")
		return failed
	}

	client := value.(*Client)
	err := client.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println("Failed to send message:", err)
		return err
	}

	return nil

}

func reader(s db.Store, userID string, conn *websocket.Conn) {
	defer func() {
		onlineUsers.Delete(userID)
		s.Delete(context.Background(), userID)
		//s.Publish(context.Background(), "onlineUsers", fmt.Sprintf("%s disconnected", userID))
		broadcast(fmt.Sprintf("%s disconnected", userID))
		conn.Close()
	}()
	messageTime := time.Now()
	//s.Publish(context.Background(), "onlineUsers", userID)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if time.Since(messageTime) > 29*time.Second {
			messageTime = time.Now()
		} else {
			log.Println("Bingus")
			continue
		}

		if heartBeat(p) {
			s.Expire(context.Background(), userID, 60)
			continue
		}

		log.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

		broadcast(string(p))

	}

}

// Change this to iterate over the Redis Cluster  and send a message if a user joins or Leaves IT
func broadcast(message string) {

	// wait isnt this redundant now that im using Redis, can just have all users subscribe to like a public channel or something
	// and publish messages there that need to be sent to all users  this approach should be more efficent aswell Thats BUSSIN!!
	onlineUsers.Range(func(key, value interface{}) bool {

		user := value.(*Client)
		// Gonna change this from userID to username i dont feel like there is a need to give other clients the userID of one of the users Though i dont feel like its a security issue regardless
		msg := Message{
			UserID:  key.(string),
			Message: message,
		}
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshalling message to JSON: %v", err)
			return true
		}
		err = user.Conn.WriteMessage(websocket.TextMessage, msgJSON)
		if err != nil {

			log.Printf("Error sending message to user %v: %v", key, err)
			return true
		}
		return true
	})
}

func handlePartyInvite(ctx context.Context, w http.ResponseWriter, s db.Store, user auth.User, notification Notification) {
	userId := user.UserID
	partyID := notification.Sender_id
	partyKey := "party:" + partyID + ":members"
	userPartyKey := "user:" + userId + ":party"

	userBytes := []byte(userId)
	partyBytes := []byte(partyID)

	count, err := s.Count(ctx, partyKey)
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	// Party cant have more than 5 people so
	if count == 5 {
		http.Error(w, "Party Is Full", http.StatusBadRequest)
		return
	}

	// add user to party
	err = s.Add(ctx, "party:"+partyID+":user", userBytes, 24*time.Hour)
	if err != nil {
		return
	}

	// this is so we can Lookup what party a user is in Fast
	if err := s.Add(ctx, userPartyKey, partyBytes, 24*time.Hour); err != nil {
		return
	}

}
