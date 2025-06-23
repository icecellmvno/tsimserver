package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetSites returns all sites with pagination
func GetSites(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	country := c.Query("country", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Site{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR country_name ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Filter by country
	if country != "" {
		query = query.Where("country = ?", country)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get sites with pagination
	var sites []models.Site
	result := query.Offset(offset).Limit(limit).
		Preload("DeviceGroups").
		Find(&sites)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch sites",
		})
	}

	return c.JSON(fiber.Map{
		"sites": sites,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSite returns a specific site by ID
func GetSite(c *fiber.Ctx) error {
	siteIDStr := c.Params("id")
	siteID, err := strconv.ParseUint(siteIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	var site models.Site
	result := database.DB.Preload("DeviceGroups.Devices.SIMCards").
		Where("id = ?", uint(siteID)).
		First(&site)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Site not found",
		})
	}

	return c.JSON(site)
}

// CreateSite creates a new site
func CreateSite(c *fiber.Ctx) error {
	var site models.Site

	if err := c.BodyParser(&site); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if site.Name == "" || site.Country == "" || site.PhoneCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name, country, and phone code are required",
		})
	}

	// Create site
	if err := database.DB.Create(&site).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create site",
		})
	}

	return c.Status(201).JSON(site)
}

// UpdateSite updates an existing site
func UpdateSite(c *fiber.Ctx) error {
	siteIDStr := c.Params("id")
	siteID, err := strconv.ParseUint(siteIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	var site models.Site
	if err := database.DB.Where("id = ?", uint(siteID)).First(&site).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Site not found",
		})
	}

	var updateData models.Site
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if updateData.Name != "" {
		site.Name = updateData.Name
	}
	if updateData.Description != "" {
		site.Description = updateData.Description
	}
	if updateData.Country != "" {
		site.Country = updateData.Country
	}
	if updateData.CountryName != "" {
		site.CountryName = updateData.CountryName
	}
	if updateData.PhoneCode != "" {
		site.PhoneCode = updateData.PhoneCode
	}
	if updateData.Address != "" {
		site.Address = updateData.Address
	}
	if updateData.Latitude != 0 {
		site.Latitude = updateData.Latitude
	}
	if updateData.Longitude != 0 {
		site.Longitude = updateData.Longitude
	}
	if updateData.ManagerName != "" {
		site.ManagerName = updateData.ManagerName
	}
	if updateData.ContactInfo != "" {
		site.ContactInfo = updateData.ContactInfo
	}
	site.IsActive = updateData.IsActive

	if err := database.DB.Save(&site).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update site",
		})
	}

	return c.JSON(site)
}

// DeleteSite deletes a site
func DeleteSite(c *fiber.Ctx) error {
	siteIDStr := c.Params("id")
	siteID, err := strconv.ParseUint(siteIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Check if site has device groups
	var groupCount int64
	database.DB.Model(&models.DeviceGroup{}).Where("site_id = ?", uint(siteID)).Count(&groupCount)

	if groupCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Cannot delete site with device groups",
			"message": "Please delete all device groups first",
		})
	}

	// Delete site
	if err := database.DB.Delete(&models.Site{}, uint(siteID)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete site",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Site deleted successfully",
	})
}

// GetSiteStats returns statistics for a specific site
func GetSiteStats(c *fiber.Ctx) error {
	siteIDStr := c.Params("id")
	siteID, err := strconv.ParseUint(siteIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid site ID",
		})
	}

	// Get site device groups
	var deviceGroups []models.DeviceGroup
	database.DB.Where("site_id = ?", uint(siteID)).Find(&deviceGroups)

	var deviceGroupIDs []uint
	for _, dg := range deviceGroups {
		deviceGroupIDs = append(deviceGroupIDs, dg.ID)
	}

	// Count devices by status
	var totalDevices, activeDevices, onlineDevices int64

	if len(deviceGroupIDs) > 0 {
		database.DB.Model(&models.Device{}).Where("device_group_id IN ?", deviceGroupIDs).Count(&totalDevices)
		database.DB.Model(&models.Device{}).Where("device_group_id IN ? AND is_active = ?", deviceGroupIDs, true).Count(&activeDevices)
		database.DB.Model(&models.Device{}).Where("device_group_id IN ? AND operator_status = ?", deviceGroupIDs, "online").Count(&onlineDevices)
	}

	// Count SMS messages today
	var smsToday int64
	database.DB.Raw(`
		SELECT COUNT(*) FROM sms_messages s
		JOIN devices d ON s.device_id = d.device_id
		WHERE d.device_group_id IN ? AND DATE(s.created_at) = CURRENT_DATE
	`, deviceGroupIDs).Scan(&smsToday)

	return c.JSON(fiber.Map{
		"site_id":        uint(siteID),
		"device_groups":  len(deviceGroups),
		"total_devices":  totalDevices,
		"active_devices": activeDevices,
		"online_devices": onlineDevices,
		"sms_today":      smsToday,
	})
}

// GetSiteCountries returns available countries
func GetSiteCountries(c *fiber.Ctx) error {
	var countries []struct {
		Country     string `json:"country"`
		CountryName string `json:"country_name"`
		PhoneCode   string `json:"phone_code"`
		Count       int64  `json:"site_count"`
	}

	result := database.DB.Model(&models.Site{}).
		Select("country, country_name, phone_code, COUNT(*) as count").
		Group("country, country_name, phone_code").
		Find(&countries)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch countries",
		})
	}

	return c.JSON(fiber.Map{
		"countries": countries,
	})
}
