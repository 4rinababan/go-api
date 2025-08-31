package sse

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SSEClient struct {
	UserID string
	Role   string
	ch     chan string
}

var (
	mu      sync.RWMutex
	clients = make(map[*SSEClient]struct{})
	byUser  = make(map[string]map[*SSEClient]struct{})
	byRole  = make(map[string]map[*SSEClient]struct{})
)

// --- registry helpers ---
func addClient(c *SSEClient) {
	mu.Lock()
	defer mu.Unlock()

	clients[c] = struct{}{}

	if c.UserID != "" {
		if _, ok := byUser[c.UserID]; !ok {
			byUser[c.UserID] = make(map[*SSEClient]struct{})
		}
		byUser[c.UserID][c] = struct{}{}
	}
	if c.Role != "" {
		if _, ok := byRole[c.Role]; !ok {
			byRole[c.Role] = make(map[*SSEClient]struct{})
		}
		byRole[c.Role][c] = struct{}{}
	}
}

func removeClient(c *SSEClient) {
	mu.Lock()
	defer mu.Unlock()

	delete(clients, c)

	if c.UserID != "" {
		if set, ok := byUser[c.UserID]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(byUser, c.UserID)
			}
		}
	}
	if c.Role != "" {
		if set, ok := byRole[c.Role]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(byRole, c.Role)
			}
		}
	}
	close(c.ch)
}

// --- broadcast APIs ---
func BroadcastAll(message string) {
	mu.RLock()
	defer mu.RUnlock()
	for c := range clients {
		select {
		case c.ch <- message:
		default:
		}
	}
}

func BroadcastToRole(role string, message string) {
	mu.RLock()
	defer mu.RUnlock()
	if set, ok := byRole[role]; ok {
		for c := range set {
			select {
			case c.ch <- message:
			default:
			}
		}
	}
}

func BroadcastToUser(userID string, message string) {
	mu.RLock()
	defer mu.RUnlock()
	if set, ok := byUser[userID]; ok {
		for c := range set {
			select {
			case c.ch <- message:
			default:
			}
		}
	}
}

// --- HTTP handler (net/http), dipakai juga oleh Gin) ---
func ServeHTTP(w http.ResponseWriter, r *http.Request, userID, role string) {
	// Header SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// (Opsional) agar aman lewat Nginx
	w.Header().Set("X-Accel-Buffering", "no")
	// (Atur sesuai domain FE kamu)
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	// Pastikan flusher tersedia
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	client := &SSEClient{
		UserID: userID,
		Role:   role,
		ch:     make(chan string, 8),
	}
	addClient(client)
	defer removeClient(client)

	// Set retry agar EventSource auto-reconnect (ms)
	fmt.Fprintf(w, "retry: 5000\n\n")
	flusher.Flush()

	// Heartbeat supaya koneksi tidak dianggap idle (proksi sering tutup idle)
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	ctx := r.Context()
	for {
		select {
		case msg := <-client.ch:
			// default event (message)
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-heartbeat.C:
			// komentar SSE (heartbeat)
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
