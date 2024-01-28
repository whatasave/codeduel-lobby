package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xedom/codeduel-lobby/types"
	"github.com/xedom/codeduel-lobby/utils"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize = 512                 // Maximum message size allowed from peer.
)

type Client struct {
	lobby *Lobby

	ID     string `json:"id"`
	Owner  bool   `json:"owner"`
	Token  string `json:"token"`
	Status string `json:"status"`
	Code   string `json:"code"`

	conn *websocket.Conn
	send chan []byte
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (c *Client) read() {
	defer func() {
		c.lobby.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error from client.read: %v", err)
			}
			break
		}

		// fmt.Printf("[CLIENT: %s] %s\n", c.ID, message)

		parsedMsg, err := utils.ParseMessage(message)
		if (err != nil) {
			fmt.Printf("Error parsing message: %v\n", err)
			continue
		}
		// fmt.Printf("Parsed message: %v\n", parsedMsg)
		// handleUserMessage(c, parsedMsg)

		fmt.Printf("msg type: %T\n", parsedMsg)
		switch parsedMsg.(type) {
		case *types.SettingsPacketIn:
			settings := *parsedMsg.(*types.SettingsPacketIn)
			fmt.Printf("test: %v\n", settings.LockLobby)
			c.SetLobbySettings(settings)
		case *types.PlayerStatusPacketIn:
			status := *parsedMsg.(*types.PlayerStatusPacketIn)
			c.SetStatus(status.Status)
		case *types.StartLobbyPacketIn:
			start := *parsedMsg.(*types.StartLobbyPacketIn)
			c.StartLobby(start.Start)
		default:
			fmt.Printf("Unknown message type: %v\n", parsedMsg)
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// c.hub.broadcast <- message
	}
}

func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SetStatus(status string) {
	fmt.Printf("Setting status to %s\n", status)
	if c.Status == status {
		return
	}

	if status != types.READY && status != types.NOT_READY && status != types.DONE {
		return
	} // TODO: find a better way to do this

	// if the client is in a match, he can only set his status to DONE
	if c.Status == types.IN_MATCH && status != types.DONE {
		return
	}

	if c.Status == types.DONE {
		return
	}

	c.Status = status
	c.lobby.broadcast <- []byte(fmt.Sprintf("User %s is %s", c.Token, status)) // TODO: send a better message to the clients
}

func (c *Client) SetLobbySettings(settings types.SettingsPacketIn) {
	fmt.Printf("Setting lobby settings: %v\n", settings)
	if !c.Owner {
		return
	}
	fmt.Printf("Owner: %v\n", c.Owner)
	if c.lobby.status != types.STARTING {
		return
	}

	c.lobby.isLocked = settings.LockLobby

	if settings.MaxPlayers > 0 && settings.MaxPlayers >= len(c.lobby.clients) {
		c.lobby.maxPlayers = settings.MaxPlayers
	}
	if settings.MaxDuration > 0 {
		c.lobby.maxDuration = settings.MaxDuration
	}
	if settings.AllowedLangs != nil && len(settings.AllowedLangs) > 0 {
		c.lobby.allowedLangs = settings.AllowedLangs
	}

	c.lobby.broadcast <- []byte(fmt.Sprintf("Lobby settings changed for lobby: isLocked %v, maxPlayers %d, maxDuration %d, allowedLangs %v", c.lobby.isLocked, c.lobby.maxPlayers, c.lobby.maxDuration, c.lobby.allowedLangs)) // TODO: send a better message to the clients
}

func (c *Client) StartLobby(start bool) {
	fmt.Printf("Starting lobby: %v\n", start)
	if !c.Owner {
		return
	}
	if c.lobby.status != types.STARTING {
		return
	}
	for client := range c.lobby.clients {
		if client.Status != types.READY {
			c.lobby.broadcast <- []byte(fmt.Sprintf("Lobby cannot start because all players must be ready"))
			return
		}
	}

	if start {
		fmt.Printf("About to start lobby\n")
		c.lobby.StartMatch()
	}
}

func serveWs(hub *Lobby, w http.ResponseWriter, r *http.Request, lobby string) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	cookie, err := r.Cookie("jwt")
	if err != nil {
		log.Fatal(err)
	}
	token := cookie.Value
	fmt.Printf("Token: %s\n", token)

	isLobbyEmpty := len(hub.clients) == 0

	client := &Client{
		lobby: hub,

		ID:     RandomID(),
		Owner:  isLobbyEmpty,
		Token:  token,
		Status: types.NOT_READY,
		Code:   lobby,

		conn: conn,
		send: make(chan []byte, 256),
	}
	client.lobby.register <- client

	go client.write()
	go client.read()

	return nil
}

func RandomID() string {
	num := rand.Intn(1000000)
	str := strconv.Itoa(num)
	hash := md5.Sum([]byte(str))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr[:8]
}
