package codeduel

import "github.com/gorilla/websocket"

type UserId int32

type User struct {
	Id         UserId
	Connection *websocket.Conn
}
