package codeduel

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize = 1024                // Maximum message size allowed from peer.
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
	connection.SetReadDeadline(time.Now().Add(pongWait))
	connection.SetPongHandler(func(string) error { connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, bytes, err := connection.ReadMessage()
		if err != nil {
			return
		}
		var packet any
		err = UnmarshalPacket(bytes, &packet)
		if err != nil {
			log.Printf("error while parsing packet: %v\n", err)
			continue
		}
		handlePacket(packet, lobby, user)
	}
}

func handlePacket(packet any, lobby *Lobby, user *User) {
	log.Printf("received packet from %v: %v\n", user, packet)
	switch packet.(type) {
	case PacketInSettings:
		handlePacketSettings(packet.(PacketInSettings), lobby, user)
	case PacketInUserStatus:
		handlePacketUserStatus(packet.(PacketInUserStatus), lobby, user)
	case PacketInStartLobby:
		handlePacketStartLobby(packet.(PacketInStartLobby), lobby, user)
	default:
		log.Printf("received unknown packet: %v\n", packet)
	}
}

func handlePacketSettings(packet PacketInSettings, lobby *Lobby, user *User) {
	lobby.SetSettings(packet.Settings)
}

func handlePacketUserStatus(packet PacketInUserStatus, lobby *Lobby, user *User) {
	// TODO
}

func handlePacketStartLobby(packet PacketInStartLobby, lobby *Lobby, user *User) {
	// TODO
}
