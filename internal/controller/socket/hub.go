package socket

import (
	"encoding/json"
	"log"
	"sync"
)

// Message representa un mensaje que se enviará por WebSocket
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	ProjectID string      `json:"project_id"`
	UserID    string      `json:"user_id"`
	Username  string      `json:"username"`
}

// Room representa una sala de chat
type Room struct {
	ID         string           `json:"id"`        // Project ID
	Clients    map[*Client]bool `json:"-"`         // Clientes conectados
	Broadcast  chan []byte      `json:"-"`         // Canal para broadcast
	Register   chan *Client     `json:"-"`         // Canal para registrar cliente
	Unregister chan *Client     `json:"-"`         // Canal para desregistrar cliente
	MaxUsers   int              `json:"max_users"` // Máximo 4 usuarios
	mutex      sync.RWMutex     `json:"-"`
}

// Hub mantiene el conjunto de clientes activos y les envía mensajes
type Hub struct {
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// NewHub crea una nueva instancia del hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// CreateRoom crea una nueva sala para un proyecto
func (h *Hub) CreateRoom(projectID string) *Room {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if room, exists := h.rooms[projectID]; exists {
		return room
	}

	room := &Room{
		ID:         projectID,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		MaxUsers:   4,
	}

	h.rooms[projectID] = room

	// Iniciar el goroutine para manejar la sala
	go room.run()

	return room
}

// GetRoom obtiene una sala por project ID
func (h *Hub) GetRoom(projectID string) *Room {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.rooms[projectID]
}

// RemoveRoom elimina una sala vacía
func (h *Hub) RemoveRoom(projectID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if room, exists := h.rooms[projectID]; exists && len(room.Clients) == 0 {
		delete(h.rooms, projectID)
		log.Printf("Sala %s eliminada", projectID)
	}
}

// Run inicia el hub principal
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			room := h.GetRoom(client.ProjectID)
			if room == nil {
				room = h.CreateRoom(client.ProjectID)
			}
			room.Register <- client

		case client := <-h.unregister:
			room := h.GetRoom(client.ProjectID)
			if room != nil {
				room.Unregister <- client
				// Si la sala está vacía, la eliminamos
				if len(room.Clients) == 0 {
					h.RemoveRoom(client.ProjectID)
				}
			}
		}
	}
}

// run maneja los eventos de una sala específica
func (r *Room) run() {
	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		select {
		case client := <-r.Register:
			r.mutex.Lock()
			// Verificar si la sala está llena
			if len(r.Clients) >= r.MaxUsers {
				r.mutex.Unlock()
				// Enviar mensaje de sala llena y cerrar conexión
				message := Message{
					Type: "error",
					Data: "La sala está llena. Máximo 4 usuarios.",
				}
				if jsonMessage, err := json.Marshal(message); err == nil {
					select {
					case client.send <- jsonMessage:
					default:
						close(client.send)
					}
				}
				continue
			}

			r.Clients[client] = true
			r.mutex.Unlock()

			// Notificar que un usuario se unió
			message := Message{
				Type:      "user_joined",
				ProjectID: r.ID,
				UserID:    client.UserID,
				Username:  client.Username,
				Data: map[string]interface{}{
					"message":     client.Username + " se unió a la sala",
					"users_count": len(r.Clients),
					"users":       r.GetConnectedUsers(),
				},
			}

			if jsonMessage, err := json.Marshal(message); err == nil {
				r.Broadcast <- jsonMessage
			}

			log.Printf("Cliente %s conectado a la sala %s. Usuarios conectados: %d",
				client.UserID, r.ID, len(r.Clients))

		case client := <-r.Unregister:
			r.mutex.Lock()
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.send)

				// Notificar que un usuario se desconectó
				message := Message{
					Type:      "user_left",
					ProjectID: r.ID,
					UserID:    client.UserID,
					Username:  client.Username,
					Data: map[string]interface{}{
						"message":     client.Username + " dejó la sala",
						"users_count": len(r.Clients),
						"users":       r.GetConnectedUsers(),
					},
				}

				if jsonMessage, err := json.Marshal(message); err == nil && len(r.Clients) > 0 {
					r.Broadcast <- jsonMessage
				}

				log.Printf("Cliente %s desconectado de la sala %s. Usuarios conectados: %d",
					client.UserID, r.ID, len(r.Clients))
			}
			r.mutex.Unlock()

		case message := <-r.Broadcast:
			r.mutex.RLock()
			for client := range r.Clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.Clients, client)
				}
			}
			r.mutex.RUnlock()
		}
	}
}

// GetConnectedUsers retorna la lista de usuarios conectados en la sala
func (r *Room) GetConnectedUsers() []map[string]string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]map[string]string, 0, len(r.Clients))
	for client := range r.Clients {
		users = append(users, map[string]string{
			"user_id":  client.UserID,
			"username": client.Username,
		})
	}
	return users
}

// BroadcastToRoom envía un mensaje a todos los clientes de la sala
func (r *Room) BroadcastToRoom(message Message) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case r.Broadcast <- jsonMessage:
		return nil
	default:
		return nil // Si el canal está bloqueado, no hacemos nada
	}
}
