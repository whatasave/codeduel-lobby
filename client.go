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
	hub *Lobby

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
		c.hub.unregister <- c
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

		fmt.Printf("[CLIENT: %s] %s\n", c.ID, message)

		parsedMsg, err := utils.PaserseMessage(message)
		fmt.Printf("Parsed message: %v\n", parsedMsg)
		handleUserMessage(c, parsedMsg)

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
		hub: hub,

		ID:     RandomID(),
		Owner:  isLobbyEmpty,
		Token:  token,
		Status: types.NOT_READY,
		Code:   lobby,

		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

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

func handleUserMessage(c *Client, message types.ClientMessage) {
	fmt.Printf("Handling message: %v\n", message)
	c.SetStatus(message.Status)
	c.SetLobbySettings(message.Settings)
	c.StartLobby(message.StartLobby)
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

	c.Status = status
	c.hub.broadcast <- []byte(fmt.Sprintf("User %s is %s", c.Token, status)) // TODO: send a better message to the clients
}

func (c *Client) SetLobbySettings(settings types.Settings) {
	fmt.Printf("Setting lobby settings: %v\n", settings)
	if !c.Owner {
		return
	}
	fmt.Printf("Owner: %v\n", c.Owner)
	if c.hub.status != types.STARTING {
		return
	}
	fmt.Printf("Status: %s\n", c.hub.status)

	c.hub.isLocked = settings.LockLobby

	if settings.MaxPlayers > 0 && settings.MaxPlayers >= len(c.hub.clients) {
		c.hub.maxPlayers = settings.MaxPlayers
	}
	if settings.MaxDuration > 0 {
		c.hub.maxDuration = settings.MaxDuration
	}
	if settings.AllowedLangs != nil && len(settings.AllowedLangs) > 0 {
		c.hub.allowedLangs = settings.AllowedLangs
	}

	c.hub.broadcast <- []byte(fmt.Sprintf("Lobby settings changed for lobby: isLocked %v, maxPlayers %d, maxDuration %d, allowedLangs %v", c.hub.isLocked, c.hub.maxPlayers, c.hub.maxDuration, c.hub.allowedLangs)) // TODO: send a better message to the clients
}

func (c *Client) StartLobby(start bool) {
	fmt.Printf("Starting lobby: %v\n", start)
	if !c.Owner {
		return
	}
	if c.hub.status != types.STARTING {
		return
	}
	for client := range c.hub.clients {
		if client.Status != types.READY {
			c.hub.broadcast <- []byte(fmt.Sprintf("Lobby cannot start because all players must be ready"))
			return
		}
	}

	if start {
		fmt.Printf("About to start lobby\n")
		c.hub.StartMatch()
	}
}
