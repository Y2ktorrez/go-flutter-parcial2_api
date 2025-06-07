package app

import (
	"fmt"
	"os"
	"time"

	"github.com/Y2ktorrez/go-flutter-parcial2_api/config"
	v1 "github.com/Y2ktorrez/go-flutter-parcial2_api/internal/controller/http/v1"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/controller/socket"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/repositories"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/services"
	impl "github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/services/Impl"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
	db     *gorm.DB
}

func New(config *config.Config) (*App, error) {
	db, err := setupDatabase(config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	app := &App{
		router: r,
		db:     db,
	}

	app.setupRoutes()
	return app, nil
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}

func setupDatabase(config *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.GetDBURL()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the database
	if err := db.AutoMigrate(&entity.User{}, &entity.Project{}); err != nil {
		return nil, err
	}

	return db, nil
}

func (a *App) setupRoutes() {
	// Initialize repositories
	userRepo := repositories.NewUserRepository(a.db)
	projectRepo := repositories.NewProjectRepository(a.db)

	// Initialize services
	userService := services.NewUserService(userRepo, os.Getenv("JWT_SECRET"))
	projectService := impl.NewProjectService(projectRepo)

	// Setup routes
	v1.SetupRoutes(a.router, userService, projectService)

	socket.SetupRoutes(a.router)
}
