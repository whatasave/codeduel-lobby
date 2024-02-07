package codeduel

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	InternalServerError = 1011
	Timeout             = 4400
	Unauthorized        = 4401
	Forbidden           = 4403
	NotFound            = 4404
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func RejectConnection(response http.ResponseWriter, request *http.Request, code int, message string) error {
	connection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		return err
	}
	closeMessage := websocket.FormatCloseMessage(code, message)
	if err := connection.WriteMessage(websocket.CloseMessage, closeMessage); err != nil {
		return err
	}
	return connection.Close()
}

func StartWebSocket(response http.ResponseWriter, request *http.Request, lobby *Lobby, user *User) (*websocket.Conn, error) {
	connection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		return nil, err
	}
	go handleClient(connection, lobby, user)
	return connection, nil
}

func handleClient(connection *websocket.Conn, lobby *Lobby, user *User) {
	defer connection.Close()
	connection.SetReadLimit(maxMessageSize)
	user.Connection = connection
	SendPacket(connection, PacketOutLobby{
		LobbyID:  lobby.Id,
		Settings: lobby.Settings,
		Users:    lobby.Users,
		State:    lobby.State,
	})
	for {
		var packet any
		err := ReadPacket(connection, &packet)
		if err != nil {
			log.Printf("error while reading packet: %v\n", err)
			closeMessage := websocket.FormatCloseMessage(Timeout, "connection timed out")
			connection.WriteMessage(websocket.CloseMessage, closeMessage)
			break
		}
		handlePacket(packet, lobby, user)
	}
}

func handlePacket(packet any, lobby *Lobby, user *User) {
	switch packet.(type) {
	case *PacketInSettings:
		handlePacketSettings(*packet.(*PacketInSettings), lobby, user)
	case *PacketInUserStatus:
		handlePacketUserStatus(*packet.(*PacketInUserStatus), lobby, user)
	case *PacketInStartLobby:
		handlePacketStartLobby(*packet.(*PacketInStartLobby), lobby, user)
	}
}

func handlePacketSettings(packet PacketInSettings, lobby *Lobby, user *User) {
	lobby.SetSettings(packet.Settings)
}

func handlePacketUserStatus(packet PacketInUserStatus, lobby *Lobby, user *User) {
	err := lobby.SetState(user, packet.Status)
	if err != nil {
		log.Printf("error while setting user state: %v\n", err)
	}
}

func handlePacketStartLobby(packet PacketInStartLobby, lobby *Lobby, user *User) {
	fmt.Println("start")
	if lobby.Owner.Id != user.Id {
		log.Printf("user %v is not the owner of the lobby\n", user)
		return
	}
	err := lobby.Start()
	if err != nil {
		log.Printf("error while starting lobby: %v\n", err)
	}
}
