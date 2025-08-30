package main

import (
	"github.com/ary/go-api/config"
	"github.com/ary/go-api/routes"
	"github.com/ary/go-api/utils"

	"github.com/ary/go-api/ws"
	// "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// "github.com/swaggo/files"
)

// @title API Meisha Alumunium Kaca
// @version 1.0
// @description API untuk manajemen produk dan kategori toko
// @termsOfService http://swagger.io/terms/
// @contact.name Ari Dev Team
// @contact.email support@meishaalumuniumkaca.com
// @host localhost:8080
// @BasePath /api
func main() {
	r := gin.Default()
	utils.SetupMonitoring(r) // /metrics
	// r.Use(cors.Default())
	config.ConnectDB()
	config.ConnectRedis()
	go ws.H.Run() // ⬅️ jalanin hub websocket
	r.Static("/uploads", "./uploads")
	// r.Use(middlewares.CORSMiddleware())
	routes.RegisterRoutes(r)

	r.Run(":8080")
}
