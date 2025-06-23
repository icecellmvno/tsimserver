package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetAlarms returns all alarms
func GetAlarms(c *fiber.Ctx) error {
	deviceID := c.Query("device_id")
	resolved := c.Query("resolved")
	severity := c.Query("severity")
	limit := c.QueryInt("limit", 100)

	query := database.DB.Model(&models.Alarm{})

	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	if resolved != "" {
		isResolved := resolved == "true"
		query = query.Where("resolved = ?", isResolved)
	}

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	var alarms []models.Alarm
	result := query.Order("created_at DESC").Limit(limit).Find(&alarms)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch alarms",
		})
	}

	return c.JSON(fiber.Map{
		"alarms": alarms,
		"count":  len(alarms),
	})
}

// GetAlarm returns a specific alarm
func GetAlarm(c *fiber.Ctx) error {
	alarmIDStr := c.Params("id")
	alarmID, err := strconv.Atoi(alarmIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid alarm ID",
		})
	}

	var alarm models.Alarm
	if err := database.DB.Where("id = ?", alarmID).First(&alarm).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Alarm not found",
		})
	}

	return c.JSON(alarm)
}

// ResolveAlarm marks an alarm as resolved
func ResolveAlarm(c *fiber.Ctx) error {
	alarmIDStr := c.Params("id")
	alarmID, err := strconv.Atoi(alarmIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid alarm ID",
		})
	}

	var alarm models.Alarm
	if err := database.DB.Where("id = ?", alarmID).First(&alarm).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Alarm not found",
		})
	}

	alarm.Resolved = true
	if err := database.DB.Save(&alarm).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to resolve alarm",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Alarm resolved successfully",
		"alarm":   alarm,
	})
}

// DeleteAlarm deletes an alarm
func DeleteAlarm(c *fiber.Ctx) error {
	alarmIDStr := c.Params("id")
	alarmID, err := strconv.Atoi(alarmIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid alarm ID",
		})
	}

	if err := database.DB.Delete(&models.Alarm{}, alarmID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete alarm",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Alarm deleted successfully",
	})
}
