package strykzaws

import (
	"fmt"

	"net/http"
	"os"

	//"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/config"
	//"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	//"github.com/joho/godotenv"
)

var (
	bucketName string
	s3Region   string
	uploadDir  string
	s3Client   *s3.Client
)

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
