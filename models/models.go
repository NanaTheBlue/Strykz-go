package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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

type NotificationType string

const (
	FriendRequest     NotificationType = "FriendRequest"
	PartyInvite       NotificationType = "PartyInvite"
	BlockNotification NotificationType = "BlockNotification"
)

type Notification struct {
	ID          string           `json:"id"`
	SenderID    string           `json:"sender_id"`
	RecipientID string           `json:"recipient_id"`
	Type        NotificationType `json:"type"`
	Data        string           `json:"data"`
	Status      string           `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
}
type FriendRequestInput struct {
	SenderID    string
	RecipientID string
}

type BlockRequest struct {
	BlockerID string `json:"blocker_id"`
	BlockedID string `json:"blocked_id"`
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
