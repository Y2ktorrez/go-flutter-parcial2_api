package socket

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

const (
	// Tiempo permitido para escribir un mensaje
	writeWait = 10 * time.Second

	// Tiempo para mantener la conexión viva
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

	for {
		var messageBytes []byte
		err := websocket.Message.Receive(c.conn, &messageBytes)
		if err != nil {
			if err.Error() != "EOF" {
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

		// Agregar información del cliente al mensaje
		incomingMessage.ProjectID = c.ProjectID
		incomingMessage.UserID = c.UserID
		incomingMessage.Username = c.Username

		// Obtener la sala y hacer broadcast del mensaje
		room := c.hub.GetRoom(c.ProjectID)
		if room != nil {
			processedMessage, err := json.Marshal(incomingMessage)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}
			room.Broadcast <- processedMessage
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
			if !ok {
				websocket.Message.Send(c.conn, "")
				return
			}

			if err := websocket.Message.Send(c.conn, string(message)); err != nil {
				log.Printf("Error sending websocket message: %v", err)
				return
			}

		case <-ticker.C:
			// Enviar ping para mantener la conexión viva
			if err := websocket.Message.Send(c.conn, `{"type":"ping"}`); err != nil {
				return
			}
		}
	}
}

// WebSocketHandler maneja las conexiones WebSocket usando gin
func WebSocketHandler(hub *Hub) gin.HandlerFunc {
	return gin.WrapH(websocket.Handler(func(ws *websocket.Conn) {
		// Obtener parámetros de la query string desde la URL
		req := ws.Request()
		projectID := req.URL.Query().Get("project_id")
		userID := req.URL.Query().Get("user_id")
		username := req.URL.Query().Get("username")

		if projectID == "" || userID == "" || username == "" {
			log.Printf("Missing required parameters: project_id=%s, user_id=%s, username=%s",
				projectID, userID, username)
			ws.Close()
			return
		}

		// Verificar si la sala existe y si está llena
		room := hub.GetRoom(projectID)
		if room != nil && len(room.Clients) >= room.MaxUsers {
			errorMsg := Message{
				Type: "error",
				Data: "La sala está llena. Máximo 4 usuarios permitidos",
			}
			if jsonMsg, err := json.Marshal(errorMsg); err == nil {
				websocket.Message.Send(ws, string(jsonMsg))
			}
			ws.Close()
			return
		}

		client := &Client{
			hub:       hub,
			conn:      ws,
			send:      make(chan []byte, 256),
			ProjectID: projectID,
			UserID:    userID,
			Username:  username,
		}

		client.hub.register <- client

		// Iniciar las goroutines para leer y escribir
		go client.writePump()
		client.readPump() // Esto bloquea hasta que la conexión se cierre
	}))
}
