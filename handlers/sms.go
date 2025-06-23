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

// SendSMS sends an SMS message
func SendSMS(c *fiber.Ctx) error {
	var smsReq struct {
		DeviceID string `json:"device_id"`
		Target   string `json:"target"`
		Message  string `json:"message"`
		SimSlot  int    `json:"sim_slot"`
	}

	if err := c.BodyParser(&smsReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate internal log ID
	internalLogID := rand.Intn(999999) + 100000

	// Save SMS to database
	sms := models.SMSMessage{
		DeviceID:      smsReq.DeviceID,
		Type:          "outgoing",
		Target:        smsReq.Target,
		Message:       smsReq.Message,
		SimSlot:       smsReq.SimSlot,
		InternalLogID: internalLogID,
		Status:        "pending",
		Timestamp:     time.Now().Unix(),
	}

	if err := database.DB.Create(&sms).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save SMS",
		})
	}

	// Send SMS command to device
	smsCmd := types.SendSMSCommand{
		Type:          "send_sms",
		Target:        smsReq.Target,
		SimSlot:       smsReq.SimSlot,
		Message:       smsReq.Message,
		InternalLogID: internalLogID,
	}

	if err := Hub.SendMessageToDevice(smsReq.DeviceID, smsCmd); err != nil {
		// Update SMS status to failed
		sms.Status = "failed"
		sms.ErrorMessage = err.Error()
		database.DB.Save(&sms)

		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to send SMS command",
		})
	}

	// Publish to queue
	queue.PublishSMSCommand(smsReq.DeviceID, smsReq.Target, smsReq.Message, smsReq.SimSlot, internalLogID)

	return c.Status(201).JSON(fiber.Map{
		"message":         "SMS sent successfully",
		"internal_log_id": internalLogID,
		"sms_id":          sms.ID,
	})
}

// GetSMSMessages returns SMS messages for a device
func GetSMSMessages(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")

	var messages []models.SMSMessage
	result := database.DB.Where("device_id = ?", deviceID).Order("created_at DESC").Limit(100).Find(&messages)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch SMS messages",
		})
	}

	return c.JSON(fiber.Map{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetSMSMessage returns a specific SMS message
func GetSMSMessage(c *fiber.Ctx) error {
	smsIDStr := c.Params("id")
	smsID, err := strconv.Atoi(smsIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid SMS ID",
		})
	}

	var sms models.SMSMessage
	if err := database.DB.Where("id = ?", smsID).First(&sms).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "SMS message not found",
		})
	}

	return c.JSON(sms)
}

// GetIncomingSMS returns incoming SMS messages
func GetIncomingSMS(c *fiber.Ctx) error {
	deviceID := c.Query("device_id")
	limit := c.QueryInt("limit", 50)

	query := database.DB.Where("type = ?", "incoming")
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	var messages []models.SMSMessage
	result := query.Order("created_at DESC").Limit(limit).Find(&messages)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch incoming SMS messages",
		})
	}

	return c.JSON(fiber.Map{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetOutgoingSMS returns outgoing SMS messages
func GetOutgoingSMS(c *fiber.Ctx) error {
	deviceID := c.Query("device_id")
	limit := c.QueryInt("limit", 50)

	query := database.DB.Where("type = ?", "outgoing")
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	var messages []models.SMSMessage
	result := query.Order("created_at DESC").Limit(limit).Find(&messages)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch outgoing SMS messages",
		})
	}

	return c.JSON(fiber.Map{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetSMSStats returns SMS statistics
func GetSMSStats(c *fiber.Ctx) error {
	deviceID := c.Query("device_id")

	var stats struct {
		TotalIncoming  int64 `json:"total_incoming"`
		TotalOutgoing  int64 `json:"total_outgoing"`
		TotalDelivered int64 `json:"total_delivered"`
		TotalFailed    int64 `json:"total_failed"`
	}

	baseQuery := database.DB.Model(&models.SMSMessage{})
	if deviceID != "" {
		baseQuery = baseQuery.Where("device_id = ?", deviceID)
	}

	// Get incoming count
	baseQuery.Where("type = ?", "incoming").Count(&stats.TotalIncoming)

	// Get outgoing count
	baseQuery.Where("type = ?", "outgoing").Count(&stats.TotalOutgoing)

	// Get delivered count
	baseQuery.Where("type = ? AND status = ?", "outgoing", "DELIVRD").Count(&stats.TotalDelivered)

	// Get failed count
	baseQuery.Where("type = ? AND (status = ? OR status = ?)", "outgoing", "failed", "UNDELIV").Count(&stats.TotalFailed)

	return c.JSON(stats)
}

// DeleteSMSMessage deletes an SMS message
func DeleteSMSMessage(c *fiber.Ctx) error {
	smsIDStr := c.Params("id")
	smsID, err := strconv.Atoi(smsIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid SMS ID",
		})
	}

	if err := database.DB.Delete(&models.SMSMessage{}, smsID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete SMS message",
		})
	}

	return c.JSON(fiber.Map{
		"message": "SMS message deleted successfully",
	})
}
