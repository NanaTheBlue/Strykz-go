package main

import (
	"fmt"
	"net/http"
	"os"
	"strykz/auth"
	"strykz/db"
	"strykz/social"
)

func main() {
	var err error
	// just the testing db username and password will change it in prod and make it a env
	db.InitDB()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.CloseDB()

	http.HandleFunc("/register", auth.Register())
	http.HandleFunc("/login", auth.Login())
	http.HandleFunc("/logout", auth.Logout())
	http.HandleFunc("/invite", social.PartyInvite())
	//http.HandleFunc("/protected", protected)
	fmt.Println("Server started on http://localhost:8080")

	http.ListenAndServe(":8080", nil)

}
