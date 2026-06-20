package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader to upgrade HTTP connection to WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for local MVP
	},
}

func handleWorker(w http.ResponseWriter, r *http.Request) {
	// 1. Upgrade the connection:
	conn, err := upgrader.Upgrade(w, r, nil) // header is nil
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("New Worker Node connected!")

	// 2. Listen for messages from Worker (infinite loop)
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Worker disconnected", err)
			break
		}
		fmt.Printf("Received from Worker: %s\n", message)

		// 3. Acknowledge the worker
		response := []byte("Master acknowledges: Standby for tasks.") // Convert string to []byte
		err = conn.WriteMessage(messageType, response)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWorker) // Registers /ws route

	fmt.Println("Master Node starting on localhost:8080...")

	err := http.ListenAndServe(":8080", nil) // Waiting live on localhost:8080/ws

	// Receives GET/ws from any Worker
	// Go checks routing table: "/ws" -> handleWorker
	// Go executs handleWorker(w,r)

	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}

}
