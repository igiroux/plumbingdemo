// websockets.go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// IndexerService data structure
type IndexerService struct {
	channel  chan int
	started  bool
	mutex    sync.Mutex
	upgrader websocket.Upgrader
}

func (ix *IndexerService) handleEcho(w http.ResponseWriter, r *http.Request) {
	conn, _ := ix.upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	for {
		// Read message from browser
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		switch string(msg) {
		case "quit", "stop", "pause":
			ix.channel <- 0
		case "start", "go", "run":
			if !ix.started {
				go ix.run()
			} else {
				fmt.Println("Already started!")
			}
		default:
			// Print the message to the console
			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}

	}

}

// NewIndexer does not need comment
func NewIndexer() *IndexerService {
	ix := &IndexerService{
		channel: make(chan int),
		started: false,
		mutex:   sync.Mutex{},
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	return ix
}
func (ix *IndexerService) start() {
	ix.mutex.Lock()
	ix.started = true
	ix.mutex.Unlock()
}

func (ix *IndexerService) stop() {
	ix.mutex.Lock()
	ix.started = false
	ix.mutex.Unlock()
}

func (ix *IndexerService) run() {
	indexing := time.Tick(2 * time.Second)
	ix.start()
	for {
		select {
		case <-indexing:
			fmt.Println("Indexing")
		case <-ix.channel:
			fmt.Println("quit")
			ix.stop()
			return
		default:
			fmt.Print(".")
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (ix *IndexerService) handleStart(w http.ResponseWriter, r *http.Request) {
	if ix.started {
		fmt.Println("Already started!")
		return
	}
	go ix.run()
}

func (ix *IndexerService) handleStop(w http.ResponseWriter, r *http.Request) {
	if ix.started {
		ix.channel <- 0
	}
}

func main() {
	ix := NewIndexer()
	http.HandleFunc("/echo", ix.handleEcho)

	http.HandleFunc("/start", ix.handleStart)
	http.HandleFunc("/stop", ix.handleStop)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})
	go ix.run()
	fmt.Println("http://localhost:8082/")
	http.ListenAndServe(":8081", nil)
}
