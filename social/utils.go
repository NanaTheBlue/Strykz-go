package social

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strykz/auth"
	"strykz/db"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
	MB = 1 << 20
)

func heartBeat(p []byte) bool {
	if string(p) == "Pong" {
		return true
	}
	return false
}

func CheckNotifications(ctx context.Context) {
	u, ok := ctx.Value(auth.UserKey).(auth.User)
	if !ok {
		fmt.Println("user not found in context")
		return
	}

	fmt.Println(u.UserID)
	notifications := db.Pool.QueryRow(context.Background(), "SELECT sender_id, type FROM notifications WHERE recipient_id = $1", u.UserID)
	if notifications != nil {
		return
	}

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

func validateImage(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 1*MB)

	file, _, err := r.FormFile("file")

	if err != nil {
		return nil, err
	}
	defer file.Close()

	imgBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	// do a mime type check first since it just reads first 512 bytes so if it fails this we can save on processing power
	mimeType := http.DetectContentType(imgBytes)

	fmt.Println(mimeType)

	if mimeType != "image/jpeg" && mimeType != "image/png" {

		return nil, fmt.Errorf("unsupported image type: %s", mimeType)

	}

	// decoding Image if it passes this its probrably a real image id say
	_, format, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("invalid image: %v", err)
	}

	if format != "jpeg" && format != "png" {
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	return imgBytes, nil

}

func uploadToR2() {

	godotenv.Load(".env")
	bucketName = "pfp"
	var accountId = os.Getenv("accountID")
	var accessKeyId = os.Getenv("accessKey")
	var accessKeySecret = os.Getenv("secretKey")
	fmt.Fprintf(os.Stderr, "Account ID: %v\n", accountId)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	listObjectsOutput, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, object := range listObjectsOutput.Contents {
		obj, _ := json.MarshalIndent(object, "", "\t")
		fmt.Println(string(obj))
	}

	//  {
	//    "ChecksumAlgorithm": null,
	//    "ETag": "\"eb2b891dc67b81755d2b726d9110af16\"",
	//    "Key": "ferriswasm.png",
	//    "LastModified": "2022-05-18T17:20:21.67Z",
	//    "Owner": null,
	//    "Size": 87671,
	//    "StorageClass": "STANDARD"
	//  }

	listBucketsOutput, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, object := range listBucketsOutput.Buckets {
		obj, _ := json.MarshalIndent(object, "", "\t")
		fmt.Println(string(obj))
	}

}
