package socket

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Tiempo permitido para escribir un mensaje
	writeWait = 10 * time.Second

	// Tiempo para leer el siguiente mensaje pong
	pongWait = 60 * time.Second

	// Enviar pings con este período
	pingPeriod = (pongWait * 9) / 10

	// Tamaño máximo del mensaje
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Upgrader para convertir HTTP a WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Permitir conexiones desde cualquier origen (ajustar según necesidades)
		return true
	},
}

// Client es un intermediario entre la conexión websocket y el hub
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	ProjectID string
	UserID    string
	Username  string
}

// readPump bombea mensajes de la conexión websocket al hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading websocket message: %v", err)
			}
			break
		}

		messageBytes = bytes.TrimSpace(bytes.Replace(messageBytes, newline, space, -1))

		// Parsear el mensaje recibido
		var incomingMessage Message
		if err := json.Unmarshal(messageBytes, &incomingMessage); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		log.Printf("Received message from client %s: type=%s", c.UserID, incomingMessage.Type)

		// Agregar información del cliente al mensaje
		incomingMessage.ProjectID = c.ProjectID
		incomingMessage.UserID = c.UserID
		incomingMessage.Username = c.Username

		// Obtener la sala y hacer broadcast del mensaje
		room := c.hub.GetRoom(c.ProjectID)
		if room != nil {
			log.Printf("Broadcasting message to room %s", c.ProjectID)
			// Usar el método BroadcastToRoom en lugar de acceder directamente al canal
			if err := room.BroadcastToRoom(incomingMessage); err != nil {
				log.Printf("Error broadcasting message: %v", err)
			}
		} else {
			log.Printf("Room not found for project: %s", c.ProjectID)
		}
	}
}

// writePump bombea mensajes del hub a la conexión websocket
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Agregar mensajes en cola al mensaje actual
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WebSocketHandler maneja las conexiones WebSocket usando gin
func WebSocketHandler(hub *Hub) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Obtener parámetros de la query string
		projectID := c.Query("project_id")
		userID := c.Query("user_id")
		username := c.Query("username")

		if projectID == "" || userID == "" || username == "" {
			log.Printf("Missing required parameters: project_id=%s, user_id=%s, username=%s",
				projectID, userID, username)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing required parameters: project_id, user_id, username",
			})
			return
		}

		// Verificar si la sala existe y si está llena
		room := hub.GetRoom(projectID)
		if room != nil {
			room.mutex.RLock()
			clientsCount := len(room.Clients)
			room.mutex.RUnlock()

			if clientsCount >= room.MaxUsers {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "La sala está llena. Máximo 4 usuarios permitidos",
				})
				return
			}
		}

		// Upgrade HTTP a WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		client := &Client{
			hub:       hub,
			conn:      conn,
			send:      make(chan []byte, 256),
			ProjectID: projectID,
			UserID:    userID,
			Username:  username,
		}

		client.hub.register <- client

		// Iniciar las goroutines para leer y escribir
		go client.writePump()
		go client.readPump()
	})
}
