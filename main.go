package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

// type Client struct {
// 	ID   string `json:"id"`
// 	// Conn *Conn

// 	Name string `json:"name"`
// 	Lobby *Lobby
// }

// type Lobby struct {
//   ID      string
//   Clients []*Client
// }

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
  fmt.Printf("Starting server on addr http://%s\n", *addr)
	hub := newHub()
	go hub.run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		err := serveWs(hub, w, r)
		if err != nil { log.Println(err) }
	})
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil { log.Fatal("ListenAndServe: ", err) }
}