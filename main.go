package main

import (
	"fmt"
	"net/http"
	"strykz/auth"
	"strykz/db"
	"strykz/que"
	"strykz/social"
)

func main() {
	//todo setup cors

	// just the testing db username and password will change it in prod and make it a env
	db.InitDB()

	defer db.CloseDB()

	http.HandleFunc("/register", auth.Rate(auth.Register(), 2, 5))
	http.HandleFunc("/login", auth.Rate(auth.Login(), 5, 10))
	http.HandleFunc("/logout", auth.Logout())
	http.HandleFunc("/invite", auth.Rate(auth.AuthMiddleware(social.PartyInvite()), 5, 10))
	http.HandleFunc("/online", social.SetOnlineStatus())
	http.HandleFunc("/updatepfp", social.ChangeProfilePicture())

	// uncomment later testing stuff http.HandleFunc("/online", auth.Rate(auth.AuthMiddleware(social.SetOnlineStatus()), 5, 10))
	http.HandleFunc("/que", auth.Rate(auth.AuthMiddleware(que.QuePlayer()), 5, 10))

	fmt.Println("Server started on http://localhost:8081")

	http.ListenAndServe(":8081", nil)

}
