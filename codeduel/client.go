package codeduel

import (
	"context"
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

func (s *APIServer) StartWebSocket(response http.ResponseWriter, request *http.Request, lobby *Lobby, user *User) (*websocket.Conn, error) {
	connection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		return nil, err
	}
	go s.handleClient(connection, lobby, user)
	return connection, nil
}

func (s *APIServer) handleClient(connection *websocket.Conn, lobby *Lobby, user *User) {
	defer connection.Close()
	connection.SetReadLimit(maxMessageSize)
	user.Connection = connection
	SendPacket(connection, PacketOutLobby{
		LobbyID:  lobby.Id,
		Settings: lobby.Settings,
		Owner:    lobby.Owner,
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
		s.handlePacket(packet, lobby, user)
	}
}

func (s *APIServer) handlePacket(packet any, lobby *Lobby, user *User) {
	switch packet := packet.(type) {
	case *PacketInSettings:
		s.handlePacketSettings(*packet, lobby, user)
	case *PacketInUserStatus:
		s.handlePacketUserStatus(*packet, lobby, user)
	case *PacketInStartLobby:
		s.handlePacketStartLobby(*packet, lobby, user)
	case *PacketInCheck:
		s.handlePacketCheck(*packet, lobby, user)
	}
}

func (s *APIServer) handlePacketSettings(packet PacketInSettings, lobby *Lobby, _ *User) {
	lobby.SetSettings(packet.Settings)
}

func (s *APIServer) handlePacketUserStatus(packet PacketInUserStatus, lobby *Lobby, user *User) {
	err := lobby.SetState(user, packet.Status)
	if err != nil {
		log.Printf("error while setting user state: %v\n", err)
	}
}

func (s *APIServer) handlePacketStartLobby(_ PacketInStartLobby, lobby *Lobby, user *User) {
	if lobby.Owner.Id != user.Id {
		log.Printf("user %v is not the owner of the lobby\n", user)
		return
	}
	err := s.startLobby(lobby, context.Background())
	if err != nil {
		log.Printf("error while starting lobby: %v\n", err)
	}
}

func (s *APIServer) handlePacketCheck(packet PacketInCheck, lobby *Lobby, user *User) {
	state, ok := lobby.State.(GameLobbyState)
	if !ok {
		log.Printf("lobby is not in game state\n")
		return
	}
	input := []string{}
	for _, testCase := range state.Challenge.TestCases {
		input = append(input, testCase.Input)
	}
	result, err := s.Runner.Run(packet.Language, packet.Code, input)
	if err != nil {
		error := fmt.Sprintf("error while running code: %v", err)
		log.Println(error)
		SendPacket(user.Connection, PacketOutCheckResult{Error: &error, Result: result})
	}
	SendPacket(user.Connection, PacketOutCheckResult{Result: result})
}
