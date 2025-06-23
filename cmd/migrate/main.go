package main

import (
	"flag"
	"log"
	"tsimserver/config"
	"tsimserver/database"
)

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "Path to config file")
		rollback   = flag.Bool("rollback", false, "Rollback last migration")
		reset      = flag.Bool("reset", false, "Reset all migrations")
	)
	flag.Parse()

	// Load configuration
	if err := config.Load(*configPath); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	if *reset {
		log.Println("Resetting database...")
		if err := database.Reset(); err != nil {
			log.Fatal("Failed to reset database:", err)
		}
		log.Println("Database reset completed")
		return
	}

	if *rollback {
		log.Println("Rolling back migrations...")
		if err := database.Rollback(); err != nil {
			log.Fatal("Failed to rollback migrations:", err)
		}
		log.Println("Migration rollback completed")
		return
	}

	// Run migrations
	log.Println("Running database migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migrations completed successfully")
}
