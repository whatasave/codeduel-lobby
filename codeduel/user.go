package codeduel

import "github.com/gorilla/websocket"

type UserId int32

type User struct {
	Id              UserId          `json:"id"`
	Username        string          `json:"username"`
	Name            string          `json:"name"`
	Avatar          string          `json:"avatar"`
	BackgroundImage string          `json:"backgroundImage"`
	Token           string          `json:"-"`
	Connection      *websocket.Conn `json:"-"`
}
