package social

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	MB = 1 << 20
)

func reader(userID string, conn *websocket.Conn) {
	defer func() {
		onlineUsers.Delete(userID)
		broadcast(fmt.Sprintf("%s has left the chat.", userID))
		conn.Close()
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

		broadcast(string(p))

	}
}
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

/*
	func checkSize(r *http.Request) int {
		size, err := r.FormFile("file")
		if err != nil {

		}

}
*/
func validateImage(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1*MB)

	image, _, err := r.FormFile("file")

	if err != nil {
		return err
	}
	defer image.Close()

	bytes, err := io.ReadAll(image)
	if err != nil {
		log.Fatal(err)
	}

	mimeType := http.DetectContentType(bytes)
	fmt.Println(mimeType)

	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {

		return fmt.Errorf("unsupported image type: %s", mimeType)
	}

	// resize the image to whataever i want so we know its a image for real for real on god on gang
	return nil

}
