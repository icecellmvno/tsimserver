package main

import (
	"flag"
	"log"
	"tsimserver/config"
	"tsimserver/database"
	"tsimserver/seeders"
)

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "Path to config file")
		worldOnly  = flag.Bool("world", false, "Seed only world data")
		authOnly   = flag.Bool("auth", false, "Seed only auth data (roles, permissions)")
		siteOnly   = flag.Bool("site", false, "Seed only site and device group data")
		verify     = flag.Bool("verify", false, "Verify seeded data")
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

	if *verify {
		log.Println("Verifying seeded data...")
		if err := seeders.VerifyWorldData(); err != nil {
			log.Fatal("Data verification failed:", err)
		}
		log.Println("Data verification completed successfully")
		return
	}

	if *worldOnly {
		log.Println("Seeding world data...")
		if err := seeders.SeedWorldDatabase(); err != nil {
			log.Fatal("Failed to seed world data:", err)
		}
		log.Println("World data seeding completed successfully")
		return
	}

	if *authOnly {
		log.Println("Seeding auth data...")
		if err := seeders.SeedAuthData(); err != nil {
			log.Fatal("Failed to seed auth data:", err)
		}
		log.Println("Auth data seeding completed successfully")
		return
	}

	if *siteOnly {
		log.Println("Seeding site data...")
		if err := seeders.SeedSiteData(); err != nil {
			log.Fatal("Failed to seed site data:", err)
		}
		log.Println("Site data seeding completed successfully")
		return
	}

	// Seed all data
	log.Println("Seeding all data...")

	// Seed world data
	if err := seeders.SeedWorldDatabase(); err != nil {
		log.Printf("Warning: Failed to seed world data: %v", err)
	} else {
		log.Println("World data seeded successfully")
	}

	// Seed auth data
	if err := seeders.SeedAuthData(); err != nil {
		log.Printf("Warning: Failed to seed auth data: %v", err)
	} else {
		log.Println("Auth data seeded successfully")
	}

	// Seed site data
	if err := seeders.SeedSiteData(); err != nil {
		log.Printf("Warning: Failed to seed site data: %v", err)
	} else {
		log.Println("Site data seeded successfully")
	}

	log.Println("All seeding operations completed")
}
