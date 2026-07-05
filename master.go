package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// DATA STRUCTURES

type Job struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
}

const (
	StatusIdle = "IDLE"
	StatusBusy = "BUSY"
)

// Worker represents a connected node
type Worker struct {
	ID         string
	Conn       *websocket.Conn
	Status     string
	CurrentJob string // Tracks the ID of the job currently being processed
	mu         sync.Mutex
}

// MasterHub manages workers and routes jobs
type MasterHub struct {
	workers    map[string]*Worker
	workerList []string
	rrIndex    int
	jobQueue   chan Job
	activeJobs map[string]Job // Active jobs
	register   chan *Worker
	unregister chan *Worker
	mu         sync.RWMutex
}

func NewMasterHub() *MasterHub {
	return &MasterHub{
		workers:    make(map[string]*Worker),
		workerList: make([]string, 0),
		jobQueue:   make(chan Job, 100),
		activeJobs: make(map[string]Job),
		register:   make(chan *Worker),
		unregister: make(chan *Worker),
	}
}

// LOAD BALANCER & DISPATCHER

func (h *MasterHub) RunDispatcher() {
	for {
		select {

		case worker := <-h.register:
			h.mu.Lock()
			h.workers[worker.ID] = worker
			h.workerList = append(h.workerList, worker.ID)
			h.mu.Unlock()
			fmt.Printf("[REGISTRY] Worker %s connected.\n", worker.ID)

		// Event: A worker disconnects or crashes
		case worker := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.workers[worker.ID]; ok {

				// FAULT TOLERANCE Logic
				worker.mu.Lock()
				if worker.Status == StatusBusy && worker.CurrentJob != "" {
					// The worker died while processing! Rescue the job.
					failedJobID := worker.CurrentJob
					rescuedJob := h.activeJobs[failedJobID]

					fmt.Printf("[FAULT TOLERANCE] Worker %s crashed! Re-queueing Job %s.\n", worker.ID, failedJobID)

					// Remove from active jobs map and push back to queue
					delete(h.activeJobs, failedJobID)
					go func() { h.jobQueue <- rescuedJob }()
				}
				worker.mu.Unlock()

				delete(h.workers, worker.ID)
				h.workerList = rebuildWorkerList(h.workers)
				h.rrIndex = 0
			}
			h.mu.Unlock()
			fmt.Printf("[REGISTRY] Worker %s disconnected.\n", worker.ID)

		// Event: A new ML job arrives in the queue
		case job := <-h.jobQueue:
			h.mu.Lock()
			if len(h.workers) == 0 {
				fmt.Printf("[ALERT] No workers available for Job %s. Dropping.\n", job.ID)
				go func(j Job) {
					time.Sleep(time.Second)
					h.jobQueue <- j
				}(job)

				h.mu.Unlock()
				continue
			}

			assigned := false
			totalWorkers := len(h.workerList)

			for i := 0; i < totalWorkers; i++ {
				currentIndex := (h.rrIndex + i) % totalWorkers
				workerID := h.workerList[currentIndex]
				worker := h.workers[workerID]

				worker.mu.Lock()
				if worker.Status == StatusIdle {

					// STATE TRACKING START
					worker.Status = StatusBusy
					worker.CurrentJob = job.ID // Assign job to worker
					h.activeJobs[job.ID] = job // Add to Master's active jobs
					// STATE TRACKING END

					fmt.Printf("[DISPATCHER] Routing Job %s to Worker %s\n", job.ID, worker.ID)

					err := worker.Conn.WriteJSON(job)
					if err != nil {
						fmt.Println("Error dispatching job:", err)
					}

					assigned = true
					worker.mu.Unlock()
					h.rrIndex = (currentIndex + 1) % totalWorkers
					break
				}
				worker.mu.Unlock()
			}

			if !assigned {
				fmt.Printf("[QUEUE] All workers busy. Re-queueing job %s.\n", job.ID)
				go func(requeuedJob Job) {
					time.Sleep(1 * time.Second)
					h.jobQueue <- requeuedJob
				}(job)
			}
			h.mu.Unlock()
		}
	}
}

func rebuildWorkerList(workerMap map[string]*Worker) []string {
	list := make([]string, 0, len(workerMap))
	for id := range workerMap {
		list = append(list, id)
	}
	return list
}

// HTTP ROUTING

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Attach handleWorker as a method to MasterHub so it can access the registry
func (h *MasterHub) handleWorker(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// Create a unique ID for the worker
	workerID := uuid.New().String()
	worker := &Worker{
		ID:     workerID,
		Conn:   conn,
		Status: StatusIdle,
	}

	h.register <- worker

	// Listen for completion messages from this worker
	defer func() {
		h.unregister <- worker
		worker.Conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// If worker says it's done, set it back to Idle and clear the active jobs map
		if string(msg) == "DONE" {
			worker.mu.Lock()

			// Remove the completed job from the active jobs map
			h.mu.Lock()
			delete(h.activeJobs, worker.CurrentJob)
			h.mu.Unlock()

			worker.Status = StatusIdle
			worker.CurrentJob = ""

			worker.mu.Unlock()
			fmt.Printf("[REGISTRY] Worker %s completed task and is now IDLE.\n", worker.ID)
		}
	}
}

// Simple endpoint to trigger jobs manually
func (h *MasterHub) submitJob(w http.ResponseWriter, r *http.Request) {
	jobID := "JOB-" + uuid.New().String()
	job := Job{
		ID:      jobID,
		Payload: "Process ML embeddings",
	}
	h.jobQueue <- job
	fmt.Fprintf(w, "Submitted %s to queue\n", jobID)
}

func main() {
	hub := NewMasterHub()
	go hub.RunDispatcher()

	http.HandleFunc("/ws", hub.handleWorker)
	http.HandleFunc("/submit", hub.submitJob)

	fmt.Println("Master Node starting on localhost:8080..")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
