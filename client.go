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

		parsedMsg, err := utils.PaserseMessage(message)
		fmt.Printf("Parsed message: %v\n", parsedMsg)

		fmt.Printf("[CLIENT: %s] %s\n", c.ID, message)
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
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

	client := &Client{
		hub: hub,

		ID:     RandomID(),
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
