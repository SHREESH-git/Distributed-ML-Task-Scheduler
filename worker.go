package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Job struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
}

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	fmt.Printf("Connecting to Master at %s\n", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial Error: ", err)
	}
	defer conn.Close()

	fmt.Println("Connected to Master. Standing by for tasks...")

	for {
		var job Job
		err := conn.ReadJSON(&job)
		if err != nil {
			log.Println("Disconnected from Master:", err)
			break
		}

		fmt.Printf("--> Received %s: %s\n", job.ID, job.Payload)

		// Simulate heavy ML workload bypassing GIL equivalent in Go
		fmt.Printf("--> Processing %s..\n", job.ID)
		time.Sleep(30 * time.Second)
		fmt.Printf("--> Finished %s!\n", job.ID)

		// Tell Master we are ready for the next job
		err = conn.WriteMessage(websocket.TextMessage, []byte("DONE"))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
