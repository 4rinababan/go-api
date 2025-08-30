package ws

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID string
	Role   string
	Conn   *websocket.Conn
	Send   chan []byte
}

type TargetMessage struct {
	UserID  string
	Message []byte
}

type RoleMessage struct {
	Role    string
	Message []byte
}

type Hub struct {
	Clients    map[*Client]bool
	ClientsMu  sync.RWMutex
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	TargetSend chan TargetMessage
	RoleSend   chan RoleMessage
}

var H = Hub{
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[*Client]bool),
	TargetSend: make(chan TargetMessage),
	RoleSend:   make(chan RoleMessage),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.ClientsMu.Lock()
			h.Clients[client] = true
			h.ClientsMu.Unlock()
			fmt.Println("✅ Client registered:", client.UserID, "| Role:", client.Role)

		case client := <-h.Unregister:
			h.ClientsMu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				fmt.Println("❌ Client unregistered:", client.UserID)
			}
			h.ClientsMu.Unlock()

		case message := <-h.Broadcast:
			h.ClientsMu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.ClientsMu.RUnlock()

		case target := <-h.TargetSend:
			h.ClientsMu.RLock()
			for client := range h.Clients {
				if client.UserID == target.UserID {
					select {
					case client.Send <- target.Message:
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			}
			h.ClientsMu.RUnlock()

		case roleMsg := <-h.RoleSend:
			h.ClientsMu.RLock()
			for client := range h.Clients {
				if client.Role == roleMsg.Role {
					select {
					case client.Send <- roleMsg.Message:
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			}
			h.ClientsMu.RUnlock()
		}
	}
}
