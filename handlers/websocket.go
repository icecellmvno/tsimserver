package handlers

import (
	"log"
	"tsimserver/websocket"

	"github.com/gofiber/fiber/v2"
	websocketLib "github.com/gofiber/websocket/v2"
)

var Hub *websocket.Hub

// InitWebSocketHub initializes the WebSocket hub
func InitWebSocketHub() {
	Hub = websocket.NewHub()
	go Hub.Run()
	log.Println("WebSocket hub started")
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(c *fiber.Ctx) error {
	// Check if the request is a WebSocket upgrade
	if websocketLib.IsWebSocketUpgrade(c) {
		return websocketLib.New(func(conn *websocketLib.Conn) {
			// Create new client
			client := websocket.NewClient(conn, Hub)

			// Register client
			Hub.Register <- client

			// Start client pumps in goroutines
			go client.WritePump()
			client.ReadPump() // This blocks until connection closes
		})(c)
	}

	return fiber.ErrUpgradeRequired
}

// GetConnectedClientsCount returns the number of connected WebSocket clients
func GetConnectedClientsCount() int {
	if Hub == nil {
		return 0
	}
	return Hub.GetConnectedClientsCount()
}

// GetActiveChannelsCount returns the number of active channels/devices
func GetActiveChannelsCount() int {
	if Hub == nil {
		return 0
	}
	return Hub.GetActiveChannelsCount()
}

// GetChannelStats returns statistics for each channel/device
func GetChannelStats() map[string]int {
	if Hub == nil {
		return make(map[string]int)
	}
	return Hub.GetChannelStats()
}
