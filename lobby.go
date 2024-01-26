package main

import (
	"fmt"
	"time"

	"github.com/xedom/codeduel-lobby/types"
)

type Lobby struct {
	clients map[*Client]bool

	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client

	status       string
	password     string
	isLocked     bool
	maxPlayers   int
	maxDuration  int
	endTimestamp int
	allowedLangs []string

	allowedPlayers []string
}

func newLobby() *Lobby {
	return &Lobby{
		clients: make(map[*Client]bool),

		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),

		status:       types.STARTING,
		isLocked:     false,
		maxPlayers:   2,
		maxDuration:  900,
		endTimestamp: 0,
		allowedLangs: []string{"ts", "py"},
	}
}

func (h *Lobby) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; !ok {
				continue
			}
			delete(h.clients, client)
			close(client.send)
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Lobby) StartMatch() {
	h.status = types.ONGOING
	h.endTimestamp = int(time.Now().Unix()) + h.maxDuration
	allowedPlayers := make([]string, 0)
	for client := range h.clients {
		if client.Status == types.READY {
			allowedPlayers = append(allowedPlayers, client.Token)
		}
	}
	h.allowedPlayers = allowedPlayers

	h.broadcast <- []byte(fmt.Sprintf("Match started: Lobby Settings {isLocked: %v, maxPlayers: %d, maxDuration: %d, allowedLangs: %v} - Allowed players: %v", h.isLocked, h.maxPlayers, h.maxDuration, h.allowedLangs, h.allowedPlayers)) // TODO: send a better message to the clients
}
