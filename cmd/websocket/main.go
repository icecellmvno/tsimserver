package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tsimserver/cache"
	"tsimserver/config"
	"tsimserver/database"
	"tsimserver/handlers"
	"tsimserver/queue"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "Path to config file")
		port       = flag.Int("port", 8081, "WebSocket server port")
	)
	flag.Parse()

	// Load configuration
	if err := config.Load(*configPath); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database (for session validation)
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize Redis
	if err := cache.Connect(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer cache.Close()

	// Initialize RabbitMQ
	if err := queue.Connect(); err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer queue.Close()

	// Initialize WebSocket hub
	handlers.InitWebSocketHub()

	// Create Fiber app for WebSocket only
	app := fiber.New(fiber.Config{
		ServerHeader: "TsimServer-WebSocket",
		AppName:      "TsimCloud WebSocket Server v1.0",
	})

	// Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "websocket",
			"version": "1.0.0",
		})
	})

	// WebSocket endpoint
	app.Get("/ws", handlers.WebSocketHandler)

	// WebSocket statistics endpoint
	app.Get("/ws/stats", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"connected_clients": handlers.GetConnectedClientsCount(),
			"active_channels":   handlers.GetActiveChannelsCount(),
		})
	})

	// Start server
	serverAddr := fmt.Sprintf(":%d", *port)
	log.Printf("TsimCloud WebSocket Server starting on %s", serverAddr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(serverAddr); err != nil {
			log.Fatal("WebSocket server failed to start:", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down WebSocket server...")

	if err := app.Shutdown(); err != nil {
		log.Fatal("WebSocket server forced to shutdown:", err)
	}

	log.Println("WebSocket server exited gracefully")
}
