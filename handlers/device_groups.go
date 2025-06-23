package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetDeviceGroups returns all device groups with pagination
func GetDeviceGroups(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	siteID := c.Query("site_id", "")
	groupType := c.Query("group_type", "")
	operator := c.Query("operator", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.DeviceGroup{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Filter by site
	if siteID != "" {
		query = query.Where("site_id = ?", siteID)
	}

	// Filter by group type
	if groupType != "" {
		query = query.Where("group_type = ?", groupType)
	}

	// Filter by operator
	if operator != "" {
		query = query.Where("operator = ?", operator)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get device groups with pagination
	var deviceGroups []models.DeviceGroup
	result := query.Offset(offset).Limit(limit).
		Preload("Site").
		Preload("Devices").
		Find(&deviceGroups)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch device groups",
		})
	}

	return c.JSON(fiber.Map{
		"device_groups": deviceGroups,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetDeviceGroup returns a specific device group by ID
func GetDeviceGroup(c *fiber.Ctx) error {
	groupIDStr := c.Params("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid device group ID",
		})
	}

	var deviceGroup models.DeviceGroup
	result := database.DB.Preload("Site").
		Preload("Devices.SIMCards").
		Where("id = ?", uint(groupID)).
		First(&deviceGroup)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device group not found",
		})
	}

	return c.JSON(deviceGroup)
}

// CreateDeviceGroup creates a new device group
func CreateDeviceGroup(c *fiber.Ctx) error {
	var deviceGroup models.DeviceGroup

	if err := c.BodyParser(&deviceGroup); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if deviceGroup.Name == "" || deviceGroup.SiteID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name and site ID are required",
		})
	}

	// Check if site exists
	var site models.Site
	if err := database.DB.Where("id = ?", deviceGroup.SiteID).First(&site).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Site not found",
		})
	}

	// Check if group name is unique within the site
	var existingGroup models.DeviceGroup
	if result := database.DB.Where("site_id = ? AND name = ?", deviceGroup.SiteID, deviceGroup.Name).First(&existingGroup); result.Error == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Device group name already exists in this site",
		})
	}

	// Create device group
	if err := database.DB.Create(&deviceGroup).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create device group",
		})
	}

	// Load relations for response
	database.DB.Preload("Site").First(&deviceGroup, deviceGroup.ID)

	return c.Status(201).JSON(deviceGroup)
}

// UpdateDeviceGroup updates an existing device group
func UpdateDeviceGroup(c *fiber.Ctx) error {
	groupIDStr := c.Params("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid device group ID",
		})
	}

	var deviceGroup models.DeviceGroup
	if err := database.DB.Where("id = ?", uint(groupID)).First(&deviceGroup).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device group not found",
		})
	}

	var updateData models.DeviceGroup
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if name is unique within the site (if name is being updated)
	if updateData.Name != "" && updateData.Name != deviceGroup.Name {
		var existingGroup models.DeviceGroup
		if result := database.DB.Where("site_id = ? AND name = ? AND id != ?",
			deviceGroup.SiteID, updateData.Name, deviceGroup.ID).First(&existingGroup); result.Error == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Device group name already exists in this site",
			})
		}
		deviceGroup.Name = updateData.Name
	}

	// Update fields
	if updateData.Description != "" {
		deviceGroup.Description = updateData.Description
	}
	if updateData.GroupType != "" {
		deviceGroup.GroupType = updateData.GroupType
	}
	if updateData.Operator != "" {
		deviceGroup.Operator = updateData.Operator
	}
	deviceGroup.IsActive = updateData.IsActive

	if err := database.DB.Save(&deviceGroup).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update device group",
		})
	}

	// Load relations for response
	database.DB.Preload("Site").First(&deviceGroup, deviceGroup.ID)

	return c.JSON(deviceGroup)
}

// DeleteDeviceGroup deletes a device group
func DeleteDeviceGroup(c *fiber.Ctx) error {
	groupIDStr := c.Params("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid device group ID",
		})
	}

	// Check if device group has devices
	var deviceCount int64
	database.DB.Model(&models.Device{}).Where("device_group_id = ?", uint(groupID)).Count(&deviceCount)

	if deviceCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Cannot delete device group with devices",
			"message": "Please move or delete all devices first",
		})
	}

	// Delete device group
	if err := database.DB.Delete(&models.DeviceGroup{}, uint(groupID)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete device group",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Device group deleted successfully",
	})
}

// GetDeviceGroupStats returns statistics for a specific device group
func GetDeviceGroupStats(c *fiber.Ctx) error {
	groupIDStr := c.Params("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid device group ID",
		})
	}

	// Count devices by status
	var totalDevices, activeDevices, onlineDevices, availableDevices int64

	database.DB.Model(&models.Device{}).Where("device_group_id = ?", uint(groupID)).Count(&totalDevices)
	database.DB.Model(&models.Device{}).Where("device_group_id = ? AND is_active = ?", uint(groupID), true).Count(&activeDevices)
	database.DB.Model(&models.Device{}).Where("device_group_id = ? AND operator_status = ?", uint(groupID), "online").Count(&onlineDevices)
	database.DB.Model(&models.Device{}).Where("device_group_id = ? AND is_available = ?", uint(groupID), true).Count(&availableDevices)

	// Count SIM cards
	var totalSIMCards, activeSIMCards int64
	database.DB.Raw(`
		SELECT COUNT(*) FROM sim_cards s
		JOIN devices d ON s.device_id = d.device_id
		WHERE d.device_group_id = ?
	`, uint(groupID)).Scan(&totalSIMCards)

	database.DB.Raw(`
		SELECT COUNT(*) FROM sim_cards s
		JOIN devices d ON s.device_id = d.device_id
		WHERE d.device_group_id = ? AND s.is_active = true
	`, uint(groupID)).Scan(&activeSIMCards)

	// Count SMS messages today
	var smsToday, smsPending, smsDelivered, smsFailed int64
	database.DB.Raw(`
		SELECT COUNT(*) FROM sms_messages s
		JOIN devices d ON s.device_id = d.device_id
		WHERE d.device_group_id = ? AND DATE(s.created_at) = CURRENT_DATE
	`, uint(groupID)).Scan(&smsToday)

	database.DB.Raw(`
		SELECT COUNT(*) FROM sms_messages s
		JOIN devices d ON s.device_id = d.device_id
		WHERE d.device_group_id = ? AND s.status = 'pending'
	`, uint(groupID)).Scan(&smsPending)

	database.DB.Raw(`
		SELECT COUNT(*) FROM sms_messages s
		JOIN devices d ON s.device_id = d.device_id
		WHERE d.device_group_id = ? AND s.status = 'delivered'
	`, uint(groupID)).Scan(&smsDelivered)

	database.DB.Raw(`
		SELECT COUNT(*) FROM sms_messages s
		JOIN devices d ON s.device_id = d.device_id
		WHERE d.device_group_id = ? AND s.status = 'failed'
	`, uint(groupID)).Scan(&smsFailed)

	return c.JSON(fiber.Map{
		"device_group_id": uint(groupID),
		"devices": fiber.Map{
			"total":     totalDevices,
			"active":    activeDevices,
			"online":    onlineDevices,
			"available": availableDevices,
		},
		"sim_cards": fiber.Map{
			"total":  totalSIMCards,
			"active": activeSIMCards,
		},
		"sms": fiber.Map{
			"today":     smsToday,
			"pending":   smsPending,
			"delivered": smsDelivered,
			"failed":    smsFailed,
		},
	})
}

// GetDeviceGroupOperators returns available operators
func GetDeviceGroupOperators(c *fiber.Ctx) error {
	var operators []struct {
		Operator string `json:"operator"`
		Count    int64  `json:"group_count"`
	}

	result := database.DB.Model(&models.DeviceGroup{}).
		Select("operator, COUNT(*) as count").
		Where("operator IS NOT NULL AND operator != ''").
		Group("operator").
		Find(&operators)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch operators",
		})
	}

	return c.JSON(fiber.Map{
		"operators": operators,
	})
}
