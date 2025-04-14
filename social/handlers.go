package social

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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

// Fetch the variables from command line arguments
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

// should add other Data in here such as UserName and profile Pic url prob remove the UserID and only do username

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

func SetOnlineStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// very important that i change this line later
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		/*
			userID, ok := r.Context().Value("userID").(string)


			if !ok {
				http.Error(w, "Username not found", http.StatusInternalServerError)
				return
			}
		*/
		userID := "Bingus"

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Client Connected to Websocket")
		onlineUsers.Store(userID, &Client{
			UserID: userID,
			Conn:   ws,
		})

		reader(userID, ws)

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

		validateImage(w, r)
		fmt.Fprintf(os.Stderr, "Bing Bong ")

		/*
			godotenv.Load(".env")
			bucketName = "pfp"
			var accountId = os.Getenv("accountID")
			var accessKeyId = os.Getenv("accessKey")
			var accessKeySecret = os.Getenv("secretKey")
			fmt.Fprintf(os.Stderr, "Account ID: %v\n", accountId)
			validateImage(r)

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
		*/

	}
}
