package models

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type Player struct {
	Player_id string `json:"player_id"`
	JoinedAt  int64  `json:"joined_at"`
}

type Match struct {
	Id      string   `json:"id"`
	Players []Player `json:"players"`
}

type Notification struct {
	Sender_id    string `json:"sender_id"`
	Recepient_id string `json: "recepient_id"`
	Type         string `json:"type"`
	Data         string `json:"data"`
}

type RegisterRequest struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmpassword"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshToken struct {
}

type Tokens struct {
	Auth_token    string
	Refresh_token string
}

type AuthClaims struct {
	UserName string `json:"userName"`
	UserId   string `json:"userId"`
	jwt.RegisteredClaims
}
