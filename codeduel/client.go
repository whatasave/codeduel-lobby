package codeduel

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
	go func() {
		err := s.handleClient(connection, lobby, user)
		if err != nil {
			_ = fmt.Errorf(err.Error())
		}
	}()
	return connection, nil
}

func (s *APIServer) handleClient(connection *websocket.Conn, lobby *Lobby, user *User) error {
	connection.SetReadLimit(maxMessageSize)
	user.Connection = connection
	err := SendPacket(connection, PacketOutLobby{
		LobbyID:  lobby.Id,
		Settings: lobby.Settings,
		Owner:    lobby.Owner,
		Users:    lobby.Users,
		State:    lobby.State,
	})
	if err != nil {
		return fmt.Errorf("error sending lobby packet: %v", err)

	}
	for {
		var packet any
		err := ReadPacket(connection, &packet)
		if err != nil {
			log.Printf("error while reading packet: %v\n", err)
			closeMessage := websocket.FormatCloseMessage(Timeout, "connection timed out")
			_ = connection.WriteMessage(websocket.CloseMessage, closeMessage)
			break
		}
		_ = s.handlePacket(packet, lobby, user)
	}
	return connection.Close()
}

func (s *APIServer) handlePacket(packet any, lobby *Lobby, user *User) error {
	switch packet := packet.(type) {
	case *PacketInSettings:
		s.handlePacketSettings(*packet, lobby, user)
	case *PacketInUserStatus:
		s.handlePacketUserStatus(*packet, lobby, user)
	case *PacketInStartLobby:
		s.handlePacketStartLobby(*packet, lobby, user)
	case *PacketInCheck:
		return s.handlePacketCheck(*packet, lobby, user)
	}
	return nil
}

func (s *APIServer) handlePacketSettings(packet PacketInSettings, lobby *Lobby, _ *User) {
	lobby.SetSettings(packet.Settings)
}

func (s *APIServer) handlePacketUserStatus(packet PacketInUserStatus, lobby *Lobby, user *User) {
	err := lobby.SetReadyState(user, packet.Status)
	if err != nil {
		log.Printf("error while setting user state: %v\n", err)
	}
}

func (s *APIServer) handlePacketStartLobby(_ PacketInStartLobby, lobby *Lobby, user *User) {
	if lobby.Owner.Id != user.Id {
		log.Printf("user %v is not the owner of the lobby\n", user)
		return
	}
	err := s.StartLobby(lobby, context.Background())
	if err != nil {
		log.Printf("error while starting lobby: %v\n", err)
	}
}

func (s *APIServer) handlePacketCheck(packet PacketInCheck, lobby *Lobby, user *User) error {
	result, err := lobby.RunTest(user, s.Runner, packet.Language, packet.Code)
	if err != nil {
		stringErr := fmt.Sprintf("err while running code: %v", err)
		return SendPacket(user.Connection, PacketOutCheckResult{Error: &stringErr, Result: result})
	}
	return SendPacket(user.Connection, PacketOutCheckResult{Result: result})
}

func (s *APIServer) handlePacketSubmit(packet PacketInCheck, lobby *Lobby, user *User) error {
	result, err := lobby.Submit(user, s.Runner, packet.Language, packet.Code)
	if err != nil {
		stringErr := fmt.Sprintf("err while running code: %v", err)
		return SendPacket(user.Connection, PacketOutSubmitResult{Error: &stringErr, Result: result})
	}
	return SendPacket(user.Connection, PacketOutSubmitResult{Result: result})
}
