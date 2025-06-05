package v1

import (
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, userService services.UserService, projectService services.ProjectService) {
	v1 := router.Group("/api/v1")
	{
		userHandler := NewUserHandler(userService)
		users := v1.Group("/users")
		{
			users.POST("/", userHandler.Create)
			users.GET("/:id", userHandler.GetByID)
			users.GET("/", userHandler.GetAll)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		projectHandler := NewProjectHandler(projectService)
		projects := v1.Group("/projects")
		{
			projects.POST("/", projectHandler.Create)
			projects.GET("/:id", projectHandler.GetByID)
			projects.GET("/", projectHandler.GetAll)
			projects.PUT("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)
		}
	}
}
