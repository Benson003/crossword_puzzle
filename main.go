package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"crossword/engine"
)

//go:embed frontend/dist/*
var assets embed.FS

var (
	globalDict *engine.Dictionary
	isSolving  bool
	solveMu    sync.Mutex
	hub        = &Registry{Clients: make(map[string]*websocket.Conn)}
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

type Registry struct {
	sync.Mutex
	Clients map[string]*websocket.Conn
}

func (reg *Registry) Broadcast(msg any) {
	reg.Lock()
	defer reg.Unlock()
	for id, conn := range reg.Clients {
		if err := conn.WriteJSON(msg); err != nil {
			conn.Close()
			delete(reg.Clients, id)
		}
	}
}

func handleTrigger(w http.ResponseWriter, r *http.Request) {
	solveMu.Lock()
	if isSolving {
		solveMu.Unlock()
		http.Error(w, "Busy", 429)
		return
	}
	isSolving = true
	solveMu.Unlock()

	puzzle := engine.NewPuzzle()
	vars, constraints := puzzle.GenerateAndExport(time.Now().Unix())

	// 1. Send the layout immediately
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars)

	// 2. Wrap the solver in a small delay or separate goroutine
	// to give the frontend time to render the black squares.
	go func() {
		defer func() {
			solveMu.Lock()
			isSolving = false
			solveMu.Unlock()
		}()

		// GIVE THE UI A BREATHER TO RENDER THE GRID
		time.Sleep(200 * time.Millisecond)

		solver := engine.NewGeneratorSolver(vars, constraints, globalDict.Words)

		solver.OnPlace = func(id int, word string) {
			fmt.Printf("[SOLVER] Placing %s at ID %d\n", word, id) // Trace it here
			hub.Broadcast(map[string]any{"type": "PLACE", "id": id, "word": word})
			time.Sleep(30 * time.Millisecond) // Slowed down slightly for visual clarity
		}

		solver.OnBacktrack = func(id int, word string) {
			fmt.Printf("[SOLVER] Backtracking %s at ID %d\n", word, id) // Trace it here
			hub.Broadcast(map[string]any{"type": "BACK", "id": id, "word": ""})
			time.Sleep(15 * time.Millisecond)
		}

		success := solver.Solve()
		res := "FAIL"
		if success {
			res = "SUCCESS"
		}
		hub.Broadcast(map[string]any{"type": "DONE", "word": res})
	}()
}

// Create a wrapper to fix MIME types
func mimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the file is a .js file, force the correct MIME type
		if path := r.URL.Path; len(path) > 3 && path[len(path)-3:] == ".js" {
			w.Header().Set("Content-Type", "application/javascript")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	dict, _ := engine.NewDictionary("./dictionary.json")
	globalDict = dict

	r := chi.NewRouter()
	r.Post("/api/solve", handleTrigger)
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Upgrade error:", err)
			return
		}

		id := r.RemoteAddr
		hub.Lock()
		hub.Clients[id] = conn
		hub.Unlock()

		fmt.Printf("[WS] Client connected: %s\n", id)

		// Keep the goroutine alive so the connection stays in the hub
		defer func() {
			hub.Lock()
			delete(hub.Clients, id)
			hub.Unlock()
			conn.Close()
			fmt.Printf("[WS] Client disconnected: %s\n", id)
		}()

		for {
			// We don't expect messages from client, but we must read to detect closure
			_, _, err := conn.ReadMessage()
			if err != nil {
				break // Exit loop if client disconnects or error occurs
			}
		}
	})

	dist, _ := fs.Sub(assets, "frontend/dist")
fileServer:=http.FileServer(http.FS(dist))

r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        
        // Manual MIME override for JS and CSS
        if len(path) > 3 && path[len(path)-3:] == ".js" {
            w.Header().Set("Content-Type", "application/javascript")
        } else if len(path) > 4 && path[len(path)-4:] == ".css" {
            w.Header().Set("Content-Type", "text/css")
        }
        
        fileServer.ServeHTTP(w, r)
    })
	
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}
