package main

import (
	"log"
	"tsimserver/config"
	"tsimserver/database"
	"tsimserver/seeders"
)

func main() {
	// Load configuration
	if err := config.Load("../../config.yaml"); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run database migrations for world tables
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Starting world database seeding...")

	// Seed world database
	if err := seeders.SeedWorldDatabase(); err != nil {
		log.Fatal("Failed to seed world database:", err)
	}

	// Verify seeded data
	if err := seeders.VerifyWorldData(); err != nil {
		log.Fatal("Failed to verify world data:", err)
	}

	log.Println("World database seeding completed successfully!")
}
