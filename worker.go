package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	// 1. Define the MasterNode URL
	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
		Path:   "/ws",
	}
	fmt.Printf("Connecting to Master at %s\n", u.String()) // Print the URL string -> ws://localhost:8080/ws

	// 2. Open a websocket connection to localhost:8080/ws
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial Error: ", err)
	}
	defer conn.Close()

	// 3. Send readiness message to the Master
	readyMessage := "I am ready for a task"
	err = conn.WriteMessage(websocket.TextMessage, []byte(readyMessage))
	if err != nil {
		log.Fatal("Write error:", err)
	}
	fmt.Println("Sent:", readyMessage)

	// 4. Wait for Master's response
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Read error:", err)
	}
	fmt.Printf("Received: %s\n", message)
	fmt.Println("Connected to Master. Standing by...")

	// Dialer converts ws://localhost:8080/ws into
	/*
		GET /ws HTTP/1.1
		Host: localhost:8080
		Connection: Upgrade
		Upgrade: websocket
		Sec-WebSocket-Version: 13
		Sec-WebSocket-Key: xyz123
	*/
	// and send it to Master Node
}
