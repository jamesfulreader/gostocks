package httpserver

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type WebSocketMessage struct {
	Action  string      `json:"action"`
	Symbol  string      `json:"symbol,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type SubscriptionManager struct {
	clients     map[*websocket.Conn]bool
	subscribers map[string]map[*websocket.Conn]bool
	broadcast   chan WebSocketMessage
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	subscribe   chan struct {
		conn   *websocket.Conn
		symbol string
	}
	mu sync.RWMutex
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		clients:     make(map[*websocket.Conn]bool),
		subscribers: make(map[string]map[*websocket.Conn]bool),
		broadcast:   make(chan WebSocketMessage),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
		subscribe: make(chan struct {
			conn   *websocket.Conn
			symbol string
		}),
	}
}

func (sm *SubscriptionManager) Run() {
	for {
		select {
		case conn := <-sm.register:
			sm.mu.Lock()
			sm.clients[conn] = true
			sm.mu.Unlock()
		case conn := <-sm.unregister:
			sm.mu.Lock()
			if _, ok := sm.clients[conn]; ok {
				delete(sm.clients, conn)
				conn.Close()
				// Remove from subscribers
				for symbol, subs := range sm.subscribers {
					delete(subs, conn)
					if len(subs) == 0 {
						delete(sm.subscribers, symbol)
					}
				}
			}
			sm.mu.Unlock()
		case sub := <-sm.subscribe:
			sm.mu.Lock()
			if sm.subscribers[sub.symbol] == nil {
				sm.subscribers[sub.symbol] = make(map[*websocket.Conn]bool)
			}
			sm.subscribers[sub.symbol][sub.conn] = true
			sm.mu.Unlock()
		case message := <-sm.broadcast:
			sm.mu.RLock()
			if message.Symbol != "" {
				// Send to subscribers of this symbol
				if subs, ok := sm.subscribers[message.Symbol]; ok {
					for conn := range subs {
						conn.WriteJSON(message)
					}
				}
			} else {
				// Broadcast to all
				for conn := range sm.clients {
					conn.WriteJSON(message)
				}
			}
			sm.mu.RUnlock()
		}
	}
}

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}

	s.subManager.register <- conn

	defer func() {
		s.subManager.unregister <- conn
	}()

	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("websocket read error: %v", err)
			break
		}

		if msg.Action == "subscribe" && msg.Symbol != "" {
			s.subManager.subscribe <- struct {
				conn   *websocket.Conn
				symbol string
			}{conn, msg.Symbol}
			log.Printf("Client subscribed to %s", msg.Symbol)
		}
	}
}

// StartTicker simulates real-time updates
func (s *Server) StartTicker() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			// In a real app, we would fetch real data here.
			// For now, we'll just iterate over active subscriptions and send a dummy update.
			s.subManager.mu.RLock()
			symbols := make([]string, 0, len(s.subManager.subscribers))
			for sym := range s.subManager.subscribers {
				symbols = append(symbols, sym)
			}
			s.subManager.mu.RUnlock()

			for _, sym := range symbols {
				// Fetch latest quote (this is blocking, ideally should be async or cached)
				// Using a background context for the ticker
				q, err := s.provider.Quote(context.Background(), sym)
				if err == nil {
					s.subManager.broadcast <- WebSocketMessage{
						Action:  "update",
						Symbol:  sym,
						Payload: q,
					}
				}
			}
		}
	}()
}
