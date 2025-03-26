package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strykz/auth"
	"strykz/social"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var err error
	// just the testing db username and password will change it in prod and make it a env
	conn, err := pgxpool.New(context.Background(), "postgres://postgres:8575@localhost:5432/strykz_database?sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	http.HandleFunc("/register", auth.Register(conn))
	http.HandleFunc("/login", auth.Login(conn))
	http.HandleFunc("/logout", auth.Logout(conn))
	http.HandleFunc("/invite", social.PartyInvite(conn))
	//http.HandleFunc("/protected", protected)
	fmt.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
