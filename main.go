package main

import (
	"go-flutter-parcial2_api/config"
	"go-flutter-parcial2_api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Inicializar configuración y base de datos
	config.LoadEnvVariables()
	config.ConnectDatabase()

	r := gin.Default()

	// Cargar rutas
	routes.SetupRoutes(r)

	r.Run() // por defecto en :8080
}
