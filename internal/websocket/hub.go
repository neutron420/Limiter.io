package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan []byte
	ProjectID string
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

type BroadcastMessage struct {
	ProjectID string
	Payload   []byte
}

type Hub struct {
	clients    map[string]map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan BroadcastMessage, 1000),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.ProjectID] == nil {
				h.clients[client.ProjectID] = make(map[*Client]bool)
			}
			h.clients[client.ProjectID][client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered for project: %s", client.ProjectID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.ProjectID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.ProjectID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered for project: %s", client.ProjectID)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.ProjectID]
			for client := range clients {
				select {
				case client.Send <- message.Payload:
				default:
					close(client.Send)
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients[message.ProjectID], client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(projectID string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal WebSocket broadcast: %v", err)
		return
	}

	select {
	case h.broadcast <- BroadcastMessage{ProjectID: projectID, Payload: data}:
	default:
		// If broadcast channel is full, drop message to prevent blocking API handlers
	}
}
