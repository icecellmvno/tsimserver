package database

import (
	"fmt"
	"log"
	"tsimserver/config"
	"tsimserver/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes database connection
func Connect() error {
	cfg := config.AppConfig.Database

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// Migrate runs database migrations
func Migrate() error {
	err := DB.AutoMigrate(
		// First create core authentication models
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Session{},
		&models.UserRole{},
		&models.RolePermission{},

		// Then create site and device management models
		&models.Site{},
		&models.DeviceGroup{},
		&models.Device{},
		&models.SIMCard{},
		&models.DeviceStatus{},

		// Then create dependent models
		&models.SMSMessage{},
		&models.USSDCommand{},
		&models.Alarm{},

		// Finally create world data models
		&models.Region{},
		&models.Subregion{},
		&models.Country{},
		&models.State{},
		&models.City{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// Close closes database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB returns database instance
func GetDB() *gorm.DB {
	return DB
}

// Reset drops all tables and recreates them
func Reset() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Drop all tables (in reverse dependency order)
	err := DB.Migrator().DropTable(
		&models.City{},
		&models.State{},
		&models.Country{},
		&models.Subregion{},
		&models.Region{},
		&models.Alarm{},
		&models.USSDCommand{},
		&models.SMSMessage{},
		&models.DeviceStatus{},
		&models.SIMCard{},
		&models.Device{},
		&models.DeviceGroup{},
		&models.Site{},
		&models.RolePermission{},
		&models.UserRole{},
		&models.Permission{},
		&models.Role{},
		&models.Session{},
		&models.User{},
	)
	if err != nil {
		return fmt.Errorf("failed to drop tables: %v", err)
	}

	// Run migrations again
	return Migrate()
}

// Rollback drops the most recently created tables
func Rollback() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Drop auth-related tables (most recent)
	err := DB.Migrator().DropTable(
		&models.RolePermission{},
		&models.UserRole{},
		&models.Permission{},
		&models.Role{},
	)
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %v", err)
	}

	log.Println("Auth-related tables dropped successfully")
	return nil
}
