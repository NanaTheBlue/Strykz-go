package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	_ "image/jpeg"
	_ "image/png"

	"log"

	"strykz/auth"
	"strykz/db"
	"time"

	"github.com/gorilla/websocket"
)

type Notification struct {
	Sender_id         string `json:"sender"`
	Notification_type string `json:"notification_type"`
}

func heartBeat(p []byte) bool {
	if string(p) == "Pong" {
		return true
	}
	return false
}

func CheckNotifications(ctx context.Context) error {
	u, ok := ctx.Value(auth.UserKey).(auth.User)
	if !ok {
		fmt.Println("user not found in context")
		return nil
	}

	var notifications []Notification

	// should prob get the notification ID and put that in the struct can use it to map over the mf aswell
	fmt.Println(u.UserID)
	rows, err := db.Pool.Query(context.Background(), "SELECT sender_id, type FROM notifications WHERE recipient_id = $1", u.UserID)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var sender_id string
		var notification_type string
		err = rows.Scan(&sender_id, &notification_type)
		if err != nil {
			return err
		}
		notifications = append(notifications, Notification{
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

	if sendToClient(u.UserID, notifications) != nil {
		failed := errors.New("WRONG MESSAGE")
		return failed
	}

	return nil

}

// Maybe should make a function that marshalls json since im using it twice so far

func sendToClient(userID string, message []Notification) error {

	value, ok := onlineUsers.Load(userID)
	if !ok {
		fmt.Println("inside sendToClient")
	}

	msgJSON, errr := json.Marshal(message)
	if errr != nil {
		log.Printf("Error marshalling message to JSON: %v", errr)
		return errr
	}

	client := value.(*Client)
	err := client.Conn.WriteMessage(websocket.TextMessage, msgJSON)
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
		s.Publish(context.Background(), "onlineUsers", fmt.Sprintf("%s disconnected", userID))
		broadcast(fmt.Sprintf("%s disconnected", userID))
		conn.Close()
	}()
	messageTime := time.Now()

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
