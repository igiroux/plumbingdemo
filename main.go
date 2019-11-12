// websockets.go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// IndexerState safely holds the state of the indexer
type IndexerState struct {
	mutex sync.Mutex
	paused bool
}

func (s *IndexerState) setPaused(paused bool) {
	s.mutex.Lock()
	s.paused = paused
	s.mutex.Unlock()
}

func (s *IndexerState) isPaused() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.paused
}

func (s *IndexerState) isRunning() bool {
	return ! s.isPaused()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// IndexerService data structure
type IndexerService struct {
	state IndexerState
	upgrader websocket.Upgrader
}



// NewIndexer does not need comment
func NewIndexer() *IndexerService {
	ix := &IndexerService{
		state:       IndexerState{paused: true},
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	return ix
}

func (ix *IndexerService) resume() {
	ix.state.setPaused(false)
}

func (ix *IndexerService) pause() {
	ix.state.setPaused(true)
}

func (ix *IndexerService) indexDoc(){
	fmt.Println("I")
	time.Sleep(100 * time.Millisecond)
}

func (ix *IndexerService) run() {
	tick := time.Tick(2 * time.Second)
	ix.state.setPaused(false)
	for {
		if ix.state.isRunning() {
			ix.indexDoc()
		}
		<- tick
	} // for ever
}

func (ix *IndexerService) progressBar(){
	tick := time.Tick(200 * time.Millisecond)
	for {
		if ix.state.isRunning() {
			fmt.Print(".")
		}
		<- tick
	}
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
			ix.pause()
		case "start", "go", "run", "resume":
			ix.resume()
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

func (ix *IndexerService) handleResume(w http.ResponseWriter, r *http.Request) {
	if ix.state.isRunning() {
		fmt.Println("Already running!")
		return
	}
	ix.resume()
}

func (ix *IndexerService) handlePause(w http.ResponseWriter, r *http.Request) {
	if ix.state.isPaused() {
		fmt.Println("Already on hold!")
		return
	}
	ix.pause()
}

func main() {
	ix := NewIndexer()
	http.HandleFunc("/echo", ix.handleEcho)

	http.HandleFunc("/resume", ix.handleResume)
	http.HandleFunc("/pause", ix.handlePause)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})
	go ix.run()
	go ix.progressBar()
	port := ":18080"
	fmt.Sprintln("http://localhost%s/", port)
	http.ListenAndServe(port, nil)
}
