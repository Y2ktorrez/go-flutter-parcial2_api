package v1

import (
	"os"

	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/middleware"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, userService services.UserService, projectService services.ProjectService) {
	jwt := os.Getenv("JWT_SECRET")
	v1 := router.Group("/api/v1")
	{
		userHandler := NewUserHandler(userService)

		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/signup", userHandler.Signup)
		}
		users := v1.Group("/users")
		users.Use(middleware.JWTMiddleware(jwt))
		{
			users.POST("/", userHandler.Create)
			users.GET("/:id", userHandler.GetByID)
			users.GET("/", userHandler.GetAll)
			users.PATCH("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		projectHandler := NewProjectHandler(projectService)
		projects := v1.Group("/projects")
		projects.Use(middleware.JWTMiddleware(jwt))
		{
			projects.POST("/", projectHandler.Create)
			projects.GET("/:id", projectHandler.GetByID)
			projects.GET("/", projectHandler.GetAll)
			projects.PATCH("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)
		}
	}
}
