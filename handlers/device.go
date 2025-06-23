package handlers

import (
	"strconv"
	"time"
	"tsimserver/database"
	"tsimserver/models"
	"tsimserver/queue"
	"tsimserver/types"

	"github.com/gofiber/fiber/v2"
)

// GetDevices returns all devices
func GetDevices(c *fiber.Ctx) error {
	var devices []models.Device

	result := database.DB.Preload("SIMCards").Find(&devices)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch devices",
		})
	}

	return c.JSON(fiber.Map{
		"devices": devices,
		"count":   len(devices),
	})
}

// GetDevice returns a specific device
func GetDevice(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	var device models.Device
	result := database.DB.Preload("SIMCards").Preload("DeviceStatuses").Where("device_id = ?", deviceID).First(&device)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
		})
	}

	return c.JSON(device)
}

// CreateDevice creates a new device
func CreateDevice(c *fiber.Ctx) error {
	var device models.Device

	if err := c.BodyParser(&device); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()
	device.IsActive = true

	if err := database.DB.Create(&device).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create device",
		})
	}

	return c.Status(201).JSON(device)
}

// UpdateDevice updates an existing device
func UpdateDevice(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	var device models.Device
	if err := database.DB.Where("device_id = ?", deviceID).First(&device).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
		})
	}

	var updateData models.Device
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	device.DeviceName = updateData.DeviceName
	device.SiteName = updateData.SiteName
	device.GroupName = updateData.GroupName
	device.IsActive = updateData.IsActive
	device.UpdatedAt = time.Now()

	if err := database.DB.Save(&device).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update device",
		})
	}

	return c.JSON(device)
}

// DeleteDevice deletes a device
func DeleteDevice(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	if err := database.DB.Where("device_id = ?", deviceID).Delete(&models.Device{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete device",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Device deleted successfully",
	})
}

// DisableDevice disables a device
func DisableDevice(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	// Update device status in database
	if err := database.DB.Model(&models.Device{}).Where("device_id = ?", deviceID).Update("is_active", false).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to disable device",
		})
	}

	// Send disable command to device
	disableCmd := types.DisableDeviceCommand{
		Type:     "disable_device",
		DeviceID: deviceID,
	}

	if err := Hub.SendMessageToDevice(deviceID, disableCmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send disable command",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Device disabled successfully",
	})
}

// EnableDevice enables a device
func EnableDevice(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	// Update device status in database
	if err := database.DB.Model(&models.Device{}).Where("device_id = ?", deviceID).Update("is_active", true).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to enable device",
		})
	}

	// Send enable command to device
	enableCmd := types.EnableDeviceCommand{
		Type:     "enable_device",
		DeviceID: deviceID,
	}

	if err := Hub.SendMessageToDevice(deviceID, enableCmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send enable command",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Device enabled successfully",
	})
}

// DisableSIM disables a SIM card
func DisableSIM(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	simSlotStr := c.Params("simslot")

	simSlot, err := strconv.Atoi(simSlotStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid SIM slot",
		})
	}

	// Update SIM status in database
	if err := database.DB.Model(&models.SIMCard{}).Where("device_id = ? AND identifier = ?", deviceID, simSlot).Update("is_enabled", false).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to disable SIM",
		})
	}

	// Send disable SIM command to device
	disableCmd := types.DisableSIMCommand{
		Type:     "disable_sim",
		DeviceID: deviceID,
		SimSlot:  simSlot,
	}

	if err := Hub.SendMessageToDevice(deviceID, disableCmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send disable SIM command",
		})
	}

	return c.JSON(fiber.Map{
		"message": "SIM disabled successfully",
	})
}

// EnableSIM enables a SIM card
func EnableSIM(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	simSlotStr := c.Params("simslot")

	simSlot, err := strconv.Atoi(simSlotStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid SIM slot",
		})
	}

	// Update SIM status in database
	if err := database.DB.Model(&models.SIMCard{}).Where("device_id = ? AND identifier = ?", deviceID, simSlot).Update("is_enabled", true).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to enable SIM",
		})
	}

	// Send enable SIM command to device
	enableCmd := types.EnableSIMCommand{
		Type:     "enable_sim",
		DeviceID: deviceID,
		SimSlot:  simSlot,
	}

	if err := Hub.SendMessageToDevice(deviceID, enableCmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send enable SIM command",
		})
	}

	return c.JSON(fiber.Map{
		"message": "SIM enabled successfully",
	})
}

// GetDeviceStatuses returns device status history
func GetDeviceStatuses(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	var statuses []models.DeviceStatus
	result := database.DB.Where("device_id = ?", deviceID).Order("created_at DESC").Limit(100).Find(&statuses)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch device statuses",
		})
	}

	return c.JSON(fiber.Map{
		"statuses": statuses,
		"count":    len(statuses),
	})
}

// SendAlarmToDevice sends an alarm to a specific device
func SendAlarmToDevice(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	var alarmReq struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}

	if err := c.BodyParser(&alarmReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create alarm message
	alarm := types.ServerAlarm{
		Type:    "alarm",
		Title:   alarmReq.Title,
		Message: alarmReq.Message,
	}

	// Send alarm to device
	if err := Hub.SendMessageToDevice(deviceID, alarm); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send alarm",
		})
	}

	// Save alarm to database
	dbAlarm := models.Alarm{
		DeviceID:  deviceID,
		Type:      "server",
		AlarmType: "manual",
		Title:     alarmReq.Title,
		Message:   alarmReq.Message,
		Severity:  "medium",
		Timestamp: time.Now().Unix(),
	}

	if err := database.DB.Create(&dbAlarm).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save alarm",
		})
	}

	// Publish to queue
	queue.PublishAlarm(deviceID, "manual", alarmReq.Message, "medium")

	return c.JSON(fiber.Map{
		"message": "Alarm sent successfully",
	})
}
