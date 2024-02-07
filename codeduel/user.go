package codeduel

import "github.com/gorilla/websocket"

type UserId int32

type User struct {
	Id         		UserId `json:"id"`
	Username		string `json:"username"`
	Email			string `json:"email"`
	Avatar			string `json:"avatar"`
	Token			string `json:"token"`
	TokenExpiresAt	string `json:"expires_at"`
	Connection 		*websocket.Conn
}
