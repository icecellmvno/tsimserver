package seeders

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"tsimserver/config"
	"tsimserver/database"
	"tsimserver/models"
)

// SeedWorldDatabase seeds the world database from SQL file
func SeedWorldDatabase() error {
	cfg := config.AppConfig.Database

	// Check if world data already exists
	var count int64
	database.DB.Model(&models.Region{}).Count(&count)
	if count > 0 {
		log.Println("World data already exists, skipping seed")
		return nil
	}

	log.Println("Starting world database seeding...")

	// Execute the SQL file using psql command
	sqlFile := "dbsource/world.sql"
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		return fmt.Errorf("SQL file not found: %s", sqlFile)
	}

	// Build psql command
	psqlCmd := fmt.Sprintf("PGPASSWORD=%s psql -h %s -p %d -U %s -d %s -f %s",
		cfg.Password, cfg.Host, cfg.Port, cfg.User, cfg.Name, sqlFile)

	// Execute command
	cmd := exec.Command("bash", "-c", psqlCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute SQL file: %v", err)
	}

	log.Println("World database seeded successfully")
	return nil
}

// SeedWorldDatabaseDirect seeds using direct SQL execution (alternative method)
func SeedWorldDatabaseDirect() error {
	// Check if world data already exists
	var count int64
	database.DB.Model(&models.Region{}).Count(&count)
	if count > 0 {
		log.Println("World data already exists, skipping seed")
		return nil
	}

	log.Println("Starting world database seeding with direct SQL execution...")

	// Read SQL file
	sqlContent, err := os.ReadFile("dbsource/world.sql")
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %v", err)
	}

	// Execute SQL content
	if err := database.DB.Exec(string(sqlContent)).Error; err != nil {
		return fmt.Errorf("failed to execute SQL content: %v", err)
	}

	log.Println("World database seeded successfully with direct execution")
	return nil
}

// SeedSiteData seeds site and device group data
func SeedSiteData() error {
	log.Println("Starting site data seeding...")

	// Create default sites
	sites := []models.Site{
		{
			Name:        "Istanbul Hub",
			Description: "Main SMS hub for Turkey operations",
			Country:     "TR",
			CountryName: "Turkey",
			PhoneCode:   "+90",
			Address:     "Istanbul Technology Park, Sarıyer, Istanbul",
			Latitude:    41.1579,
			Longitude:   29.0503,
			ManagerName: "Ahmet Yilmaz",
			ContactInfo: "ahmet@company.com",
			IsActive:    true,
		},
		{
			Name:        "Ankara Backup",
			Description: "Backup SMS hub for Turkey operations",
			Country:     "TR",
			CountryName: "Turkey",
			PhoneCode:   "+90",
			Address:     "ODTU Teknokent, Ankara",
			Latitude:    39.8917,
			Longitude:   32.7834,
			ManagerName: "Mehmet Demir",
			ContactInfo: "mehmet@company.com",
			IsActive:    true,
		},
		{
			Name:        "London Hub",
			Description: "SMS hub for UK operations",
			Country:     "UK",
			CountryName: "United Kingdom",
			PhoneCode:   "+44",
			Address:     "Tech City, London",
			Latitude:    51.5074,
			Longitude:   -0.1278,
			ManagerName: "John Smith",
			ContactInfo: "john@company.com",
			IsActive:    true,
		},
	}

	for _, site := range sites {
		var existingSite models.Site
		if result := database.DB.Where("name = ?", site.Name).First(&existingSite); result.Error != nil {
			if err := database.DB.Create(&site).Error; err != nil {
				log.Printf("Failed to create site %s: %v", site.Name, err)
				return err
			}
			log.Printf("Created site: %s", site.Name)
		} else {
			log.Printf("Site already exists: %s", site.Name)
		}
	}

	// Create device groups for sites
	var createdSites []models.Site
	database.DB.Find(&createdSites)

	for _, site := range createdSites {
		deviceGroups := []models.DeviceGroup{}

		if site.Country == "TR" {
			deviceGroups = []models.DeviceGroup{
				{
					SiteID:      site.ID,
					Name:        "Turkcell Group",
					Description: "Devices with Turkcell SIM cards",
					GroupType:   "operator",
					Operator:    "Turkcell",
					IsActive:    true,
				},
				{
					SiteID:      site.ID,
					Name:        "Vodafone Group",
					Description: "Devices with Vodafone SIM cards",
					GroupType:   "operator",
					Operator:    "Vodafone",
					IsActive:    true,
				},
				{
					SiteID:      site.ID,
					Name:        "Turk Telekom Group",
					Description: "Devices with Türk Telekom SIM cards",
					GroupType:   "operator",
					Operator:    "Turk Telekom",
					IsActive:    true,
				},
			}
		} else if site.Country == "UK" {
			deviceGroups = []models.DeviceGroup{
				{
					SiteID:      site.ID,
					Name:        "EE Group",
					Description: "Devices with EE SIM cards",
					GroupType:   "operator",
					Operator:    "EE",
					IsActive:    true,
				},
				{
					SiteID:      site.ID,
					Name:        "Vodafone UK Group",
					Description: "Devices with Vodafone UK SIM cards",
					GroupType:   "operator",
					Operator:    "Vodafone UK",
					IsActive:    true,
				},
			}
		}

		for _, group := range deviceGroups {
			var existingGroup models.DeviceGroup
			if result := database.DB.Where("site_id = ? AND name = ?", group.SiteID, group.Name).First(&existingGroup); result.Error != nil {
				if err := database.DB.Create(&group).Error; err != nil {
					log.Printf("Failed to create device group %s: %v", group.Name, err)
					return err
				}
				log.Printf("Created device group: %s for site %s", group.Name, site.Name)
			} else {
				log.Printf("Device group already exists: %s", group.Name)
			}
		}
	}

	log.Println("Site data seeding completed successfully")
	return nil
}

// VerifyWorldData verifies that world data was seeded correctly
func VerifyWorldData() error {
	var regionCount, subregionCount, countryCount, stateCount, cityCount int64

	database.DB.Model(&models.Region{}).Count(&regionCount)
	database.DB.Model(&models.Subregion{}).Count(&subregionCount)
	database.DB.Model(&models.Country{}).Count(&countryCount)
	database.DB.Model(&models.State{}).Count(&stateCount)
	database.DB.Model(&models.City{}).Count(&cityCount)

	log.Printf("World Data Verification:")
	log.Printf("- Regions: %d", regionCount)
	log.Printf("- Subregions: %d", subregionCount)
	log.Printf("- Countries: %d", countryCount)
	log.Printf("- States: %d", stateCount)
	log.Printf("- Cities: %d", cityCount)

	if regionCount == 0 || countryCount == 0 {
		return fmt.Errorf("world data seems incomplete")
	}

	return nil
}

// SeedAuthData seeds authentication related data (roles, permissions, default admin user)
func SeedAuthData() error {
	// Create default roles
	roles := []models.Role{
		{
			Name:        "admin",
			DisplayName: "Administrator",
			Description: "Full system access",
			IsActive:    true,
		},
		{
			Name:        "manager",
			DisplayName: "Manager",
			Description: "Device and SMS management",
			IsActive:    true,
		},
		{
			Name:        "operator",
			DisplayName: "Operator",
			Description: "Limited operations access",
			IsActive:    true,
		},
		{
			Name:        "viewer",
			DisplayName: "Viewer",
			Description: "Read-only access",
			IsActive:    true,
		},
	}

	// Create roles if they don't exist
	for _, role := range roles {
		var existingRole models.Role
		if err := database.DB.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if err := database.DB.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %v", role.Name, err)
			}
			log.Printf("Created role: %s", role.Name)
		}
	}

	// Create default permissions
	permissions := []models.Permission{
		// User management
		{Name: "users.read", DisplayName: "Read Users", Resource: "users", Action: "read", IsActive: true},
		{Name: "users.write", DisplayName: "Write Users", Resource: "users", Action: "write", IsActive: true},
		{Name: "users.delete", DisplayName: "Delete Users", Resource: "users", Action: "delete", IsActive: true},

		// Role management
		{Name: "roles.read", DisplayName: "Read Roles", Resource: "roles", Action: "read", IsActive: true},
		{Name: "roles.write", DisplayName: "Write Roles", Resource: "roles", Action: "write", IsActive: true},
		{Name: "roles.delete", DisplayName: "Delete Roles", Resource: "roles", Action: "delete", IsActive: true},

		// Permission management
		{Name: "permissions.read", DisplayName: "Read Permissions", Resource: "permissions", Action: "read", IsActive: true},
		{Name: "permissions.write", DisplayName: "Write Permissions", Resource: "permissions", Action: "write", IsActive: true},
		{Name: "permissions.delete", DisplayName: "Delete Permissions", Resource: "permissions", Action: "delete", IsActive: true},

		// Device management
		{Name: "devices.read", DisplayName: "Read Devices", Resource: "devices", Action: "read", IsActive: true},
		{Name: "devices.write", DisplayName: "Write Devices", Resource: "devices", Action: "write", IsActive: true},
		{Name: "devices.delete", DisplayName: "Delete Devices", Resource: "devices", Action: "delete", IsActive: true},

		// SMS management
		{Name: "sms.read", DisplayName: "Read SMS", Resource: "sms", Action: "read", IsActive: true},
		{Name: "sms.write", DisplayName: "Write SMS", Resource: "sms", Action: "write", IsActive: true},
		{Name: "sms.delete", DisplayName: "Delete SMS", Resource: "sms", Action: "delete", IsActive: true},

		// USSD management
		{Name: "ussd.read", DisplayName: "Read USSD", Resource: "ussd", Action: "read", IsActive: true},
		{Name: "ussd.write", DisplayName: "Write USSD", Resource: "ussd", Action: "write", IsActive: true},
		{Name: "ussd.delete", DisplayName: "Delete USSD", Resource: "ussd", Action: "delete", IsActive: true},

		// Alarm management
		{Name: "alarms.read", DisplayName: "Read Alarms", Resource: "alarms", Action: "read", IsActive: true},
		{Name: "alarms.write", DisplayName: "Write Alarms", Resource: "alarms", Action: "write", IsActive: true},
		{Name: "alarms.delete", DisplayName: "Delete Alarms", Resource: "alarms", Action: "delete", IsActive: true},

		// Statistics
		{Name: "stats.read", DisplayName: "Read Statistics", Resource: "stats", Action: "read", IsActive: true},

		// World data management
		{Name: "regions.read", DisplayName: "Read Regions", Resource: "regions", Action: "read", IsActive: true},
		{Name: "regions.write", DisplayName: "Write Regions", Resource: "regions", Action: "write", IsActive: true},
		{Name: "regions.delete", DisplayName: "Delete Regions", Resource: "regions", Action: "delete", IsActive: true},

		{Name: "countries.read", DisplayName: "Read Countries", Resource: "countries", Action: "read", IsActive: true},
		{Name: "countries.write", DisplayName: "Write Countries", Resource: "countries", Action: "write", IsActive: true},
		{Name: "countries.delete", DisplayName: "Delete Countries", Resource: "countries", Action: "delete", IsActive: true},

		{Name: "states.read", DisplayName: "Read States", Resource: "states", Action: "read", IsActive: true},
		{Name: "states.write", DisplayName: "Write States", Resource: "states", Action: "write", IsActive: true},
		{Name: "states.delete", DisplayName: "Delete States", Resource: "states", Action: "delete", IsActive: true},

		{Name: "cities.read", DisplayName: "Read Cities", Resource: "cities", Action: "read", IsActive: true},
		{Name: "cities.write", DisplayName: "Write Cities", Resource: "cities", Action: "write", IsActive: true},
		{Name: "cities.delete", DisplayName: "Delete Cities", Resource: "cities", Action: "delete", IsActive: true},

		{Name: "subregions.read", DisplayName: "Read Subregions", Resource: "subregions", Action: "read", IsActive: true},
		{Name: "subregions.write", DisplayName: "Write Subregions", Resource: "subregions", Action: "write", IsActive: true},
		{Name: "subregions.delete", DisplayName: "Delete Subregions", Resource: "subregions", Action: "delete", IsActive: true},

		// Site management
		{Name: "sites.read", DisplayName: "Read Sites", Resource: "sites", Action: "read", IsActive: true},
		{Name: "sites.write", DisplayName: "Write Sites", Resource: "sites", Action: "write", IsActive: true},
		{Name: "sites.delete", DisplayName: "Delete Sites", Resource: "sites", Action: "delete", IsActive: true},

		// Device group management
		{Name: "device_groups.read", DisplayName: "Read Device Groups", Resource: "device_groups", Action: "read", IsActive: true},
		{Name: "device_groups.write", DisplayName: "Write Device Groups", Resource: "device_groups", Action: "write", IsActive: true},
		{Name: "device_groups.delete", DisplayName: "Delete Device Groups", Resource: "device_groups", Action: "delete", IsActive: true},

		// Admin-level permissions
		{Name: "sms.admin", DisplayName: "SMS Admin", Resource: "sms", Action: "admin", IsActive: true},
		{Name: "devices.admin", DisplayName: "Device Admin", Resource: "devices", Action: "admin", IsActive: true},
	}

	// Create permissions if they don't exist
	for _, permission := range permissions {
		var existingPermission models.Permission
		if err := database.DB.Where("name = ?", permission.Name).First(&existingPermission).Error; err != nil {
			if err := database.DB.Create(&permission).Error; err != nil {
				return fmt.Errorf("failed to create permission %s: %v", permission.Name, err)
			}
			log.Printf("Created permission: %s", permission.Name)
		}
	}

	// Create default admin user
	var adminUser models.User
	if err := database.DB.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		adminUser = models.User{
			Username:  "admin",
			Email:     "admin@tsimserver.com",
			Password:  "admin123",
			FirstName: "System",
			LastName:  "Administrator",
			IsActive:  true,
		}

		// Hash password
		if err := adminUser.HashPassword(); err != nil {
			return fmt.Errorf("failed to hash admin password: %v", err)
		}

		if err := database.DB.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}
		log.Println("Created default admin user (admin/admin123)")
	}

	// Assign admin role to admin user
	var adminRole models.Role
	if err := database.DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("admin role not found: %v", err)
	}

	// Check if user already has admin role
	var userRole models.UserRole
	if err := database.DB.Where("user_id = ? AND role_id = ?", adminUser.ID, adminRole.ID).First(&userRole).Error; err != nil {
		userRole = models.UserRole{
			UserID: adminUser.ID,
			RoleID: adminRole.ID,
		}
		if err := database.DB.Create(&userRole).Error; err != nil {
			return fmt.Errorf("failed to assign admin role: %v", err)
		}
		log.Println("Assigned admin role to admin user")
	}

	// Assign all permissions to admin role
	var allPermissions []models.Permission
	database.DB.Find(&allPermissions)

	for _, permission := range allPermissions {
		var rolePermission models.RolePermission
		if err := database.DB.Where("role_id = ? AND permission_id = ?", adminRole.ID, permission.ID).First(&rolePermission).Error; err != nil {
			rolePermission = models.RolePermission{
				RoleID:       adminRole.ID,
				PermissionID: permission.ID,
			}
			database.DB.Create(&rolePermission)
		}
	}

	log.Println("Auth data seeding completed successfully")
	return nil
}
