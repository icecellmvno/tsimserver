package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tsimserver/auth"
	"tsimserver/cache"
	"tsimserver/config"
	"tsimserver/database"
	"tsimserver/handlers"
	"tsimserver/middleware"
	"tsimserver/queue"
	"tsimserver/seeders"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	if err := config.Load("config.yaml"); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run database migrations
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed world database
	if err := seeders.SeedWorldDatabase(); err != nil {
		log.Printf("Warning: Failed to seed world database: %v", err)
	}

	// Initialize Casbin
	if err := auth.InitCasbin(); err != nil {
		log.Fatal("Failed to initialize Casbin:", err)
	}

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

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ServerHeader: "TsimServer",
		AppName:      "TsimCloud Server v1.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// WebSocket endpoint
	app.Get(config.AppConfig.WebSocket.Endpoint, handlers.WebSocketHandler)

	// API Routes
	setupRoutes(app)

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	log.Printf("TsimCloud Server starting on %s", serverAddr)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(serverAddr); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal
	<-c
	log.Println("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}

func setupRoutes(app *fiber.App) {
	// API v1 group
	v1 := app.Group("/api/v1")

	// Health check
	v1.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"server":  "TsimCloud Server",
			"version": "1.0.0",
		})
	})

	// Authentication routes (public)
	auth := v1.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/register", handlers.Register)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/logout", middleware.AuthRequired(), handlers.Logout)
	auth.Get("/profile", middleware.AuthRequired(), handlers.GetProfile)
	auth.Put("/profile", middleware.AuthRequired(), handlers.UpdateProfile)
	auth.Put("/change-password", middleware.AuthRequired(), handlers.ChangePassword)

	// Protected routes - Admin only
	adminRequired := middleware.AuthRequired()

	// User management routes (admin only)
	users := v1.Group("/users", adminRequired, middleware.RequirePermission("users", "read"))
	users.Get("/", handlers.GetUsers)
	users.Get("/:id", handlers.GetUser)
	users.Post("/", middleware.RequirePermission("users", "write"), handlers.CreateUser)
	users.Put("/:id", middleware.RequirePermission("users", "write"), handlers.UpdateUser)
	users.Delete("/:id", middleware.RequirePermission("users", "delete"), handlers.DeleteUser)
	users.Post("/:id/roles", middleware.RequirePermission("users", "write"), handlers.AssignRoleToUser)
	users.Delete("/:id/roles/:role_id", middleware.RequirePermission("users", "write"), handlers.RemoveRoleFromUser)
	users.Get("/:id/roles", handlers.GetUserRoles)
	users.Get("/:id/sessions", handlers.GetUserSessions)
	users.Delete("/:id/sessions/:session_id", middleware.RequirePermission("users", "write"), handlers.RevokeUserSession)

	// Role management routes (admin only)
	roles := v1.Group("/roles", adminRequired, middleware.RequirePermission("roles", "read"))
	roles.Get("/", handlers.GetRoles)
	roles.Get("/:id", handlers.GetRole)
	roles.Post("/", middleware.RequirePermission("roles", "write"), handlers.CreateRole)
	roles.Put("/:id", middleware.RequirePermission("roles", "write"), handlers.UpdateRole)
	roles.Delete("/:id", middleware.RequirePermission("roles", "delete"), handlers.DeleteRole)
	roles.Post("/:id/permissions", middleware.RequirePermission("roles", "write"), handlers.AssignPermissionToRole)
	roles.Delete("/:id/permissions/:permission_id", middleware.RequirePermission("roles", "write"), handlers.RemovePermissionFromRole)
	roles.Get("/:id/users", handlers.GetRoleUsers)

	// Permission management routes (admin only)
	permissions := v1.Group("/permissions", adminRequired, middleware.RequirePermission("permissions", "read"))
	permissions.Get("/", handlers.GetPermissions)
	permissions.Get("/:id", handlers.GetPermission)
	permissions.Post("/", middleware.RequirePermission("permissions", "write"), handlers.CreatePermission)
	permissions.Put("/:id", middleware.RequirePermission("permissions", "write"), handlers.UpdatePermission)
	permissions.Delete("/:id", middleware.RequirePermission("permissions", "delete"), handlers.DeletePermission)
	permissions.Get("/resources", handlers.GetPermissionResources)
	permissions.Get("/actions", handlers.GetPermissionActions)
	permissions.Post("/bulk", middleware.RequirePermission("permissions", "write"), handlers.BulkCreatePermissions)

	// Site management routes (admin only)
	sites := v1.Group("/sites", adminRequired, middleware.RequirePermission("sites", "read"))
	sites.Get("/", handlers.GetSites)
	sites.Get("/:id", handlers.GetSite)
	sites.Post("/", middleware.RequirePermission("sites", "write"), handlers.CreateSite)
	sites.Put("/:id", middleware.RequirePermission("sites", "write"), handlers.UpdateSite)
	sites.Delete("/:id", middleware.RequirePermission("sites", "delete"), handlers.DeleteSite)
	sites.Get("/:id/stats", handlers.GetSiteStats)
	sites.Get("/countries", handlers.GetSiteCountries)

	// Device group management routes (admin only)
	deviceGroups := v1.Group("/device-groups", adminRequired, middleware.RequirePermission("device_groups", "read"))
	deviceGroups.Get("/", handlers.GetDeviceGroups)
	deviceGroups.Get("/:id", handlers.GetDeviceGroup)
	deviceGroups.Post("/", middleware.RequirePermission("device_groups", "write"), handlers.CreateDeviceGroup)
	deviceGroups.Put("/:id", middleware.RequirePermission("device_groups", "write"), handlers.UpdateDeviceGroup)
	deviceGroups.Delete("/:id", middleware.RequirePermission("device_groups", "delete"), handlers.DeleteDeviceGroup)
	deviceGroups.Get("/:id/stats", handlers.GetDeviceGroupStats)
	deviceGroups.Get("/operators", handlers.GetDeviceGroupOperators)

	// Device routes (protected)
	devices := v1.Group("/devices", middleware.AuthRequired(), middleware.RequirePermission("devices", "read"))
	devices.Get("/", handlers.GetDevices)
	devices.Post("/", middleware.RequirePermission("devices", "write"), handlers.CreateDevice)
	devices.Get("/:id", handlers.GetDevice)
	devices.Put("/:id", middleware.RequirePermission("devices", "write"), handlers.UpdateDevice)
	devices.Delete("/:id", middleware.RequirePermission("devices", "delete"), handlers.DeleteDevice)
	devices.Post("/:id/disable", middleware.RequirePermission("devices", "write"), handlers.DisableDevice)
	devices.Post("/:id/enable", middleware.RequirePermission("devices", "write"), handlers.EnableDevice)
	devices.Post("/:id/sim/:simslot/disable", middleware.RequirePermission("devices", "write"), handlers.DisableSIM)
	devices.Post("/:id/sim/:simslot/enable", middleware.RequirePermission("devices", "write"), handlers.EnableSIM)
	devices.Get("/:id/statuses", handlers.GetDeviceStatuses)
	devices.Post("/:id/alarm", middleware.RequirePermission("alarms", "write"), handlers.SendAlarmToDevice)

	// SMS routes (protected)
	sms := v1.Group("/sms", middleware.AuthRequired(), middleware.RequirePermission("sms", "read"))
	sms.Post("/send", middleware.RequirePermission("sms", "write"), handlers.SendSMS)
	sms.Get("/incoming", handlers.GetIncomingSMS)
	sms.Get("/outgoing", handlers.GetOutgoingSMS)
	sms.Get("/stats", handlers.GetSMSStats)
	sms.Get("/device/:deviceId", handlers.GetSMSMessages)
	sms.Get("/:id", handlers.GetSMSMessage)
	sms.Delete("/:id", middleware.RequirePermission("sms", "delete"), handlers.DeleteSMSMessage)

	// SMS Gateway routes (protected)
	smsGateway := v1.Group("/sms-gateway", middleware.AuthRequired(), middleware.RequirePermission("sms", "write"))
	smsGateway.Post("/send", handlers.SendSMSViaGateway)
	smsGateway.Post("/test", middleware.RequirePermission("sms", "admin"), handlers.SendTestSMS)
	smsGateway.Post("/command", middleware.RequirePermission("devices", "admin"), handlers.SendTestCommand)
	smsGateway.Post("/dlr", handlers.ProcessDeliveryReport) // Internal endpoint for devices

	// USSD routes (protected)
	ussd := v1.Group("/ussd", middleware.AuthRequired(), middleware.RequirePermission("ussd", "read"))
	ussd.Post("/send", middleware.RequirePermission("ussd", "write"), handlers.SendUSSD)
	ussd.Get("/device/:deviceId", handlers.GetUSSDCommands)
	ussd.Get("/:id", handlers.GetUSSDCommand)
	ussd.Delete("/:id", middleware.RequirePermission("ussd", "delete"), handlers.DeleteUSSDCommand)

	// Alarm routes (protected)
	alarms := v1.Group("/alarms", middleware.AuthRequired(), middleware.RequirePermission("alarms", "read"))
	alarms.Get("/", handlers.GetAlarms)
	alarms.Get("/:id", handlers.GetAlarm)
	alarms.Post("/:id/resolve", middleware.RequirePermission("alarms", "write"), handlers.ResolveAlarm)
	alarms.Delete("/:id", middleware.RequirePermission("alarms", "delete"), handlers.DeleteAlarm)

	// Statistics routes (protected)
	stats := v1.Group("/stats", middleware.AuthRequired(), middleware.RequirePermission("stats", "read"))
	stats.Get("/dashboard", handlers.GetDashboardStats)
	stats.Get("/devices", handlers.GetDeviceStats)

	// World data routes (public read, auth required for write)
	// Regions
	regions := v1.Group("/regions", middleware.OptionalAuth())
	regions.Get("/", handlers.GetRegions)
	regions.Get("/:id", handlers.GetRegion)
	regions.Post("/", middleware.AuthRequired(), middleware.RequirePermission("regions", "write"), handlers.CreateRegion)
	regions.Put("/:id", middleware.AuthRequired(), middleware.RequirePermission("regions", "write"), handlers.UpdateRegion)
	regions.Delete("/:id", middleware.AuthRequired(), middleware.RequirePermission("regions", "delete"), handlers.DeleteRegion)
	regions.Get("/:id/subregions", handlers.GetRegionSubregions)
	regions.Get("/:id/countries", handlers.GetRegionCountries)

	// Subregions
	subregions := v1.Group("/subregions", middleware.OptionalAuth())
	subregions.Get("/", handlers.GetSubregions)
	subregions.Get("/:id", handlers.GetSubregion)
	subregions.Post("/", middleware.AuthRequired(), middleware.RequirePermission("subregions", "write"), handlers.CreateSubregion)
	subregions.Put("/:id", middleware.AuthRequired(), middleware.RequirePermission("subregions", "write"), handlers.UpdateSubregion)
	subregions.Delete("/:id", middleware.AuthRequired(), middleware.RequirePermission("subregions", "delete"), handlers.DeleteSubregion)
	subregions.Get("/:id/countries", handlers.GetSubregionCountries)

	// Countries
	countries := v1.Group("/countries", middleware.OptionalAuth())
	countries.Get("/", handlers.GetCountries)
	countries.Get("/:id", handlers.GetCountry)
	countries.Get("/iso/:iso", handlers.GetCountryByISO)
	countries.Post("/", middleware.AuthRequired(), middleware.RequirePermission("countries", "write"), handlers.CreateCountry)
	countries.Put("/:id", middleware.AuthRequired(), middleware.RequirePermission("countries", "write"), handlers.UpdateCountry)
	countries.Delete("/:id", middleware.AuthRequired(), middleware.RequirePermission("countries", "delete"), handlers.DeleteCountry)
	countries.Get("/:id/states", handlers.GetCountryStates)
	countries.Get("/:id/cities", handlers.GetCountryCities)

	// States
	states := v1.Group("/states", middleware.OptionalAuth())
	states.Get("/", handlers.GetStates)
	states.Get("/:id", handlers.GetState)
	states.Post("/", middleware.AuthRequired(), middleware.RequirePermission("states", "write"), handlers.CreateState)
	states.Put("/:id", middleware.AuthRequired(), middleware.RequirePermission("states", "write"), handlers.UpdateState)
	states.Delete("/:id", middleware.AuthRequired(), middleware.RequirePermission("states", "delete"), handlers.DeleteState)
	states.Get("/:id/cities", handlers.GetStateCities)

	// Cities
	cities := v1.Group("/cities", middleware.OptionalAuth())
	cities.Get("/", handlers.GetCities)
	cities.Get("/:id", handlers.GetCity)
	cities.Post("/", middleware.AuthRequired(), middleware.RequirePermission("cities", "write"), handlers.CreateCity)
	cities.Put("/:id", middleware.AuthRequired(), middleware.RequirePermission("cities", "write"), handlers.UpdateCity)
	cities.Delete("/:id", middleware.AuthRequired(), middleware.RequirePermission("cities", "delete"), handlers.DeleteCity)
	cities.Get("/search", handlers.SearchCitiesByCoordinates)
	cities.Get("/stats", handlers.GetCitiesStats)
}
