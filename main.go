package main

import (
	"fmt"
	"log"
	"net/http"
	"strykz/auth"
	"strykz/db"
	"strykz/que"
	"strykz/social"
	"strykz/strykzaws"
)

func main() {

	//todo setup cors
	client := db.InitRedis()
	store := db.NewRedisInstance(client)

	// just the testing db username and password will change it in prod and make it a env

	db.InitDB()

	defer db.CloseDB()

	publisher, err := que.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	http.HandleFunc("/register", auth.Rate(auth.Register(), 2, 5))
	//http.HandleFunc("/login", auth.Cors(auth.Rate(auth.Login(), 5, 10)))

	http.HandleFunc("/login", auth.Login())

	http.HandleFunc("/logout", auth.Logout())
	http.HandleFunc("/invite", auth.Rate(auth.AuthMiddleware(social.PartyInvite(store)), 5, 10))
	http.HandleFunc("/online", auth.Cors(auth.WSAuth(social.SetOnlineStatus(store))))
	http.HandleFunc("/updatepfp", strykzaws.ChangeProfilePicture())

	// uncomment later testing stuff http.HandleFunc("/online", auth.Rate(auth.AuthMiddleware(social.SetOnlineStatus()), 5, 10))
	http.HandleFunc("/que", auth.Rate(auth.WSAuth(que.QuePlayer()), 5, 10))

	fmt.Println("Server started on http://localhost:8081")

	http.ListenAndServe(":8081", nil)

}
