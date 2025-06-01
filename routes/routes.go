package routes

import (
	"go-flutter-parcial2_api/controllers"
	"go-flutter-parcial2_api/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// Aquí puedes agregar más rutas protegidas o públicas

	protected := r.Group("/protected")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Ruta protegida!"})
		})
	}
}
