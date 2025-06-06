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
	done       chan struct{}    `json:"-"` // Canal para terminar la goroutine
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
		Broadcast:  make(chan []byte, 256), // Buffer para evitar bloqueos
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		MaxUsers:   4,
		done:       make(chan struct{}),
	}

	h.rooms[projectID] = room

	// Iniciar el goroutine para manejar la sala
	go room.run(h)

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

	if room, exists := h.rooms[projectID]; exists {
		room.mutex.RLock()
		clientsCount := len(room.Clients)
		room.mutex.RUnlock()

		if clientsCount == 0 {
			// Señalar que la sala debe cerrarse
			close(room.done)
			delete(h.rooms, projectID)
			log.Printf("Sala %s eliminada", projectID)
		}
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
			}
		}
	}
}

// run maneja los eventos de una sala específica
func (r *Room) run(hub *Hub) {
	defer func() {
		// Limpiar canales al finalizar
		r.mutex.Lock()
		for client := range r.Clients {
			close(client.send)
		}
		r.mutex.Unlock()

		// Cerrar canales de la sala
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
			usersCount := len(r.Clients)
			r.mutex.Unlock()

			// Notificar que un usuario se unió
			message := Message{
				Type:      "user_joined",
				ProjectID: r.ID,
				UserID:    client.UserID,
				Username:  client.Username,
				Data: map[string]interface{}{
					"message":     client.Username + " se unió a la sala",
					"users_count": usersCount,
					"users":       r.GetConnectedUsers(),
				},
			}

			r.broadcastMessage(message)
			log.Printf("Cliente %s conectado a la sala %s. Usuarios conectados: %d",
				client.UserID, r.ID, usersCount)

		case client := <-r.Unregister:
			r.mutex.Lock()
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.send)
				usersCount := len(r.Clients)
				r.mutex.Unlock()

				// Notificar que un usuario se desconectó
				if usersCount > 0 { // Solo notificar si quedan usuarios
					message := Message{
						Type:      "user_left",
						ProjectID: r.ID,
						UserID:    client.UserID,
						Username:  client.Username,
						Data: map[string]interface{}{
							"message":     client.Username + " dejó la sala",
							"users_count": usersCount,
							"users":       r.GetConnectedUsers(),
						},
					}
					r.broadcastMessage(message)
				}

				log.Printf("Cliente %s desconectado de la sala %s. Usuarios conectados: %d",
					client.UserID, r.ID, usersCount)

				// Si no quedan usuarios, programar eliminación de la sala
				if usersCount == 0 {
					go func() {
						// Esperar un poco antes de eliminar la sala por si alguien se reconecta
						// time.Sleep(30 * time.Second)
						hub.RemoveRoom(r.ID)
					}()
				}
			} else {
				r.mutex.Unlock()
			}

		case message := <-r.Broadcast:
			r.mutex.RLock()
			clientsToSend := make([]*Client, 0, len(r.Clients))
			for client := range r.Clients {
				clientsToSend = append(clientsToSend, client)
			}
			r.mutex.RUnlock()

			log.Printf("Broadcasting to %d clients in room %s", len(clientsToSend), r.ID)

			// Enviar mensaje a todos los clientes
			var disconnectedClients []*Client
			for _, client := range clientsToSend {
				select {
				case client.send <- message:
					log.Printf("Message sent to client %s", client.UserID)
				default:
					// Cliente no puede recibir mensajes, marcar para desconectar
					log.Printf("Client %s channel full, marking for disconnect", client.UserID)
					disconnectedClients = append(disconnectedClients, client)
				}
			}

			// Limpiar clientes desconectados
			if len(disconnectedClients) > 0 {
				r.mutex.Lock()
				for _, client := range disconnectedClients {
					if _, ok := r.Clients[client]; ok {
						close(client.send)
						delete(r.Clients, client)
						log.Printf("Removed disconnected client %s", client.UserID)
					}
				}
				r.mutex.Unlock()
			}

		case <-r.done:
			// Sala marcada para eliminación
			return
		}
	}
}

// broadcastMessage envía un mensaje a todos los clientes de la sala
func (r *Room) broadcastMessage(message Message) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	// Usar select con default para evitar bloqueos
	select {
	case r.Broadcast <- jsonMessage:
	default:
		log.Printf("Broadcast channel full, dropping message")
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
		log.Printf("Error marshaling message: %v", err)
		return err
	}

	log.Printf("Sending message to broadcast channel for room %s, type: %s", r.ID, message.Type)

	select {
	case r.Broadcast <- jsonMessage:
		log.Printf("Message successfully queued for broadcast in room %s", r.ID)
		return nil
	default:
		log.Printf("Broadcast channel full for room %s, message dropped", r.ID)
		return nil
	}
}
