package handlers

import (
	"math/rand"
	"strconv"
	"time"
	"tsimserver/database"
	"tsimserver/models"
	"tsimserver/queue"
	"tsimserver/types"

	"github.com/gofiber/fiber/v2"
)

// SendUSSD sends a USSD command
func SendUSSD(c *fiber.Ctx) error {
	var ussdReq struct {
		DeviceID string `json:"device_id"`
		USSDCode string `json:"ussd_code"`
		SimSlot  int    `json:"sim_slot"`
	}

	if err := c.BodyParser(&ussdReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate internal log ID
	internalLogID := rand.Intn(999999) + 100000

	// Save USSD command to database
	ussd := models.USSDCommand{
		DeviceID:      ussdReq.DeviceID,
		USSDCode:      ussdReq.USSDCode,
		SimSlot:       ussdReq.SimSlot,
		InternalLogID: internalLogID,
		Status:        "pending",
		Timestamp:     time.Now().Unix(),
	}

	if err := database.DB.Create(&ussd).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save USSD command",
		})
	}

	// Send USSD command to device
	ussdCmd := types.USSDCommand{
		Type:          "ussd_command",
		USSDCode:      ussdReq.USSDCode,
		SimSlot:       ussdReq.SimSlot,
		InternalLogID: internalLogID,
	}

	if err := Hub.SendMessageToDevice(ussdReq.DeviceID, ussdCmd); err != nil {
		// Update USSD status to failed
		ussd.Status = "failed"
		ussd.ErrorMessage = err.Error()
		database.DB.Save(&ussd)

		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send USSD command",
		})
	}

	// Publish to queue
	queue.PublishUSSDCommand(ussdReq.DeviceID, ussdReq.USSDCode, ussdReq.SimSlot, internalLogID)

	return c.Status(201).JSON(fiber.Map{
		"message":         "USSD command sent successfully",
		"internal_log_id": internalLogID,
		"ussd_id":         ussd.ID,
	})
}

// GetUSSDCommands returns USSD commands for a device
func GetUSSDCommands(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")

	var commands []models.USSDCommand
	result := database.DB.Where("device_id = ?", deviceID).Order("created_at DESC").Limit(100).Find(&commands)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch USSD commands",
		})
	}

	return c.JSON(fiber.Map{
		"commands": commands,
		"count":    len(commands),
	})
}

// GetUSSDCommand returns a specific USSD command
func GetUSSDCommand(c *fiber.Ctx) error {
	ussdIDStr := c.Params("id")
	ussdID, err := strconv.Atoi(ussdIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid USSD ID",
		})
	}

	var ussd models.USSDCommand
	if err := database.DB.Where("id = ?", ussdID).First(&ussd).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "USSD command not found",
		})
	}

	return c.JSON(ussd)
}

// DeleteUSSDCommand deletes a USSD command
func DeleteUSSDCommand(c *fiber.Ctx) error {
	ussdIDStr := c.Params("id")
	ussdID, err := strconv.Atoi(ussdIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid USSD ID",
		})
	}

	if err := database.DB.Delete(&models.USSDCommand{}, ussdID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete USSD command",
		})
	}

	return c.JSON(fiber.Map{
		"message": "USSD command deleted successfully",
	})
}

// CheckBalance sends balance check USSD command
func CheckBalance(c *fiber.Ctx) error {
	var balanceReq struct {
		DeviceID string `json:"device_id"`
		SimSlot  int    `json:"sim_slot"`
		USSDCode string `json:"ussd_code"`
	}

	if err := c.BodyParser(&balanceReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate internal log ID
	internalLogID := rand.Intn(999999) + 100000

	// Send balance check command to device
	balanceCmd := types.CheckBalanceCommand{
		Type:          "check_balance",
		SimSlot:       balanceReq.SimSlot,
		USSDCode:      balanceReq.USSDCode,
		InternalLogID: internalLogID,
	}

	if err := Hub.SendMessageToDevice(balanceReq.DeviceID, balanceCmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send balance check command",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":         "Balance check sent successfully",
		"internal_log_id": internalLogID,
	})
}

// DiscoverPhoneNumber sends phone number discovery USSD command
func DiscoverPhoneNumber(c *fiber.Ctx) error {
	var phoneReq struct {
		DeviceID string `json:"device_id"`
		SimSlot  int    `json:"sim_slot"`
		USSDCode string `json:"ussd_code"`
	}

	if err := c.BodyParser(&phoneReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate internal log ID
	internalLogID := rand.Intn(999999) + 100000

	// Send phone number discovery command to device
	phoneCmd := types.DiscoverPhoneNumberCommand{
		Type:          "discover_phone_number",
		SimSlot:       phoneReq.SimSlot,
		USSDCode:      phoneReq.USSDCode,
		InternalLogID: internalLogID,
	}

	if err := Hub.SendMessageToDevice(phoneReq.DeviceID, phoneCmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send phone number discovery command",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":         "Phone number discovery sent successfully",
		"internal_log_id": internalLogID,
	})
}
