package socket

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler maneja las conexiones WebSocket
type Handler struct {
	hub *Hub
}

// NewHandler crea una nueva instancia del handler
func NewHandler() *Handler {
	hub := NewHub()
	go hub.Run() // Iniciar el hub en una goroutine separada

	return &Handler{
		hub: hub,
	}
}

// HandleWebSocket retorna el handler de gin para WebSocket
func (h *Handler) HandleWebSocket() gin.HandlerFunc {
	return WebSocketHandler(h.hub)
}

// GetRoomInfo obtiene información de una sala
func (h *Handler) GetRoomInfo(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id es requerido"})
		return
	}

	// Validar autenticación
	userID := c.GetHeader("X-User-ID")
	if userID == "" || !h.isUserAuthenticated(c, userID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	room := h.hub.GetRoom(projectID)
	if room == nil {
		c.JSON(http.StatusOK, gin.H{
			"project_id":      projectID,
			"users_count":     0,
			"max_users":       4,
			"connected_users": []interface{}{},
			"is_full":         false,
		})
		return
	}

	room.mutex.RLock()
	usersCount := len(room.Clients)
	connectedUsers := room.GetConnectedUsers()
	room.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"project_id":      projectID,
		"users_count":     usersCount,
		"max_users":       room.MaxUsers,
		"connected_users": connectedUsers,
		"is_full":         usersCount >= room.MaxUsers,
	})
}

// SendMessage envía un mensaje a una sala específica (endpoint REST)
func (h *Handler) SendMessage(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id es requerido"})
		return
	}

	var req struct {
		Type     string      `json:"type" binding:"required"`
		Data     interface{} `json:"data"`
		UserID   string      `json:"user_id" binding:"required"`
		Username string      `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar autenticación
	if !h.isUserAuthenticated(c, req.UserID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	room := h.hub.GetRoom(projectID)
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sala no encontrada"})
		return
	}

	message := Message{
		Type:      req.Type,
		Data:      req.Data,
		ProjectID: projectID,
		UserID:    req.UserID,
		Username:  req.Username,
	}

	if err := room.BroadcastToRoom(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error enviando mensaje"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mensaje enviado correctamente",
		"type":    req.Type,
	})
}

// GetActiveRooms obtiene todas las salas activas
func (h *Handler) GetActiveRooms(c *gin.Context) {
	// Validar que el usuario esté autenticado
	userID := c.GetHeader("X-User-ID")
	if userID == "" || !h.isUserAuthenticated(c, userID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	h.hub.mutex.RLock()
	rooms := make([]map[string]interface{}, 0, len(h.hub.rooms))

	for projectID, room := range h.hub.rooms {
		room.mutex.RLock()
		roomInfo := map[string]interface{}{
			"project_id":      projectID,
			"users_count":     len(room.Clients),
			"max_users":       room.MaxUsers,
			"is_full":         len(room.Clients) >= room.MaxUsers,
			"connected_users": room.GetConnectedUsers(),
		}
		room.mutex.RUnlock()
		rooms = append(rooms, roomInfo)
	}
	h.hub.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
		"total": len(rooms),
	})
}

// KickUser expulsa a un usuario de una sala (solo para administradores)
func (h *Handler) KickUser(c *gin.Context) {
	projectID := c.Param("project_id")
	targetUserID := c.Param("user_id")

	if projectID == "" || targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id y user_id son requeridos"})
		return
	}

	// Validar autenticación del usuario que quiere expulsar
	adminUserID := c.GetHeader("X-User-ID")
	if adminUserID == "" || !h.isUserAuthenticated(c, adminUserID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// Aquí deberías validar que el usuario tenga permisos de administrador
	// if !h.isUserAdmin(adminUserID, projectID) { ... }

	room := h.hub.GetRoom(projectID)
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sala no encontrada"})
		return
	}

	// Buscar el cliente a expulsar
	room.mutex.RLock()
	var targetClient *Client
	for client := range room.Clients {
		if client.UserID == targetUserID {
			targetClient = client
			break
		}
	}
	room.mutex.RUnlock()

	if targetClient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado en la sala"})
		return
	}

	// Enviar mensaje de expulsión antes de desconectar
	kickMessage := Message{
		Type:      "kicked",
		ProjectID: projectID,
		UserID:    targetUserID,
		Username:  targetClient.Username,
		Data: map[string]interface{}{
			"reason": "Expulsado por administrador",
			"admin":  adminUserID,
		},
	}

	room.BroadcastToRoom(kickMessage)

	// Desconectar al usuario
	h.hub.unregister <- targetClient

	c.JSON(http.StatusOK, gin.H{
		"message": "Usuario expulsado correctamente",
		"user_id": targetUserID,
	})
}

// isUserAuthenticated valida si el usuario está autenticado
func (h *Handler) isUserAuthenticated(c *gin.Context, userID string) bool {
	// Implementar según tu sistema de autenticación
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	// Ejemplo con Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		// Aquí validarías el JWT token
		// return validateJWTToken(token, userID)
		return token != "" && len(token) > 10 // Placeholder
	}

	return false
}

// isUserAdmin verifica si el usuario es administrador del proyecto (implementar según necesidades)
func (h *Handler) isUserAdmin(userID, projectID string) bool {
	// Implementar validación de permisos de administrador
	// Podría consultar la base de datos para verificar roles
	return false // Placeholder
}
