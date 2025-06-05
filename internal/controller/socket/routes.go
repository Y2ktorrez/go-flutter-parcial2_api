package socket

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configura las rutas para WebSocket
func SetupRoutes(router *gin.Engine) {
	handler := NewHandler()

	// Grupo de rutas para WebSocket
	ws := router.Group("/ws")
	{
		// Conexión WebSocket principal
		// URL: ws://localhost:8080/ws/connect?project_id=123&user_id=456&username=Juan
		ws.GET("/connect", handler.HandleWebSocket())

		// Información de una sala específica
		ws.GET("/room/:project_id", handler.GetRoomInfo)

		// Enviar mensaje a una sala via REST
		ws.POST("/room/:project_id/message", handler.SendMessage)

		// Obtener todas las salas activas
		ws.GET("/rooms", handler.GetActiveRooms)

		// Expulsar usuario de una sala (para administradores)
		ws.DELETE("/room/:project_id/user/:user_id", handler.KickUser)
	}
}
