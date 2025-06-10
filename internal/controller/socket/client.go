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
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
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
	CheckOrigin:     func(r *http.Request) bool { return true },
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

// readPump (sin cambios)
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

		var incoming Message
		if err := json.Unmarshal(messageBytes, &incoming); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		incoming.ProjectID = c.ProjectID
		incoming.UserID = c.UserID
		incoming.Username = c.Username

		if room := c.hub.GetRoom(c.ProjectID); room != nil {
			if err := room.BroadcastToRoom(incoming); err != nil {
				log.Printf("Error broadcasting message: %v", err)
			}
		}
	}
}

/* ---------- SOLUCIÓN: writePump sin concatenar JSON ---------- */
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// helper para escribir un frame
	write := func(msg []byte) error {
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		return c.conn.WriteMessage(websocket.TextMessage, msg)
	}

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// canal cerrado → cerrar WS
				write(websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			if err := write(msg); err != nil {
				return
			}

			// drenar backlog sin concatenar
		DRAIN:
			for {
				select {
				case next := <-c.send:
					if err := write(next); err != nil {
						return
					}
				default:
					break DRAIN
				}
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WebSocketHandler (sin cambios)
func WebSocketHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Query("project_id")
		userID := c.Query("user_id")
		username := c.Query("username")

		if projectID == "" || userID == "" || username == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing required parameters: project_id, user_id, username",
			})
			return
		}

		if room := hub.GetRoom(projectID); room != nil {
			room.mutex.RLock()
			full := len(room.Clients) >= room.MaxUsers
			room.mutex.RUnlock()
			if full {
				c.JSON(http.StatusForbidden, gin.H{"error": "La sala está llena. Máximo 4 usuarios"})
				return
			}
		}

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
		go client.writePump()
		go client.readPump()
	}
}
