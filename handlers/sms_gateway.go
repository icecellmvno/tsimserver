package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"
	"tsimserver/database"
	"tsimserver/models"
	"tsimserver/queue"
	"tsimserver/websocket"

	"github.com/gofiber/fiber/v2"
)

// Global WebSocket hub reference
var wsHub *websocket.Hub

// InitializeWebSocketHub sets the websocket hub reference
func InitializeWebSocketHub(hub *websocket.Hub) {
	wsHub = hub
}

// SMSSendRequest represents SMS send request
type SMSSendRequest struct {
	Target        string `json:"target" validate:"required"`
	Message       string `json:"message" validate:"required"`
	Country       string `json:"country"`                         // Optional: specific country
	Operator      string `json:"operator"`                        // Optional: specific operator
	Priority      int    `json:"priority" validate:"min=1,max=5"` // 1-5, higher is more priority
	ScheduledAt   string `json:"scheduled_at"`                    // Optional: ISO timestamp
	IsTestMessage bool   `json:"is_test_message"`                 // Admin test flag
}

// SMSGatewayResponse represents SMS Gateway response
type SMSGatewayResponse struct {
	Success       bool    `json:"success"`
	MessageID     uint    `json:"message_id"`
	DeviceID      string  `json:"device_id"`
	SimSlot       int     `json:"sim_slot"`
	EstimatedCost float64 `json:"estimated_cost"`
	Message       string  `json:"message"`
}

// SendSMSViaGateway sends SMS through the intelligent gateway system
func SendSMSViaGateway(c *fiber.Ctx) error {
	var req SMSSendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate phone number
	if !isValidPhoneNumber(req.Target) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid phone number format",
		})
	}

	// Set default priority
	if req.Priority == 0 {
		req.Priority = 1
	}

	// Get user ID for admin test messages
	var adminUserID *uint
	if req.IsTestMessage {
		if userID, ok := c.Locals("user_id").(uint); ok {
			adminUserID = &userID
		}
	}

	// Find best device and SIM for sending SMS
	device, simCard, err := findBestDeviceForSMS(req.Target, req.Country, req.Operator)
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"error":   "No available device found",
			"details": err.Error(),
		})
	}

	// Parse scheduled time if provided
	var scheduledAt *time.Time
	if req.ScheduledAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.ScheduledAt); err == nil {
			scheduledAt = &parsed
		}
	}

	// Create SMS message record
	smsMessage := models.SMSMessage{
		DeviceID:      device.DeviceID,
		Type:          "outgoing",
		Target:        req.Target,
		Message:       req.Message,
		SimSlot:       int(simCard.ID), // Use SIMCard ID as slot reference
		Status:        "pending",
		Priority:      req.Priority,
		ScheduledAt:   scheduledAt,
		IsTestMessage: req.IsTestMessage,
		AdminUserID:   adminUserID,
		Timestamp:     time.Now().Unix(),
	}

	if err := database.DB.Create(&smsMessage).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create SMS record",
		})
	}

	// Send SMS to device via WebSocket
	if err := sendSMSToDevice(device.DeviceID, smsMessage); err != nil {
		// Update SMS status to failed
		smsMessage.Status = "failed"
		smsMessage.ErrorMessage = err.Error()
		database.DB.Save(&smsMessage)

		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to send SMS to device",
			"details": err.Error(),
		})
	}

	// Update SMS status to sent
	smsMessage.Status = "sent"
	database.DB.Save(&smsMessage)

	return c.JSON(SMSGatewayResponse{
		Success:       true,
		MessageID:     smsMessage.ID,
		DeviceID:      device.DeviceID,
		SimSlot:       int(simCard.ID),
		EstimatedCost: calculateSMSCost(req.Target, len(req.Message)),
		Message:       "SMS sent successfully",
	})
}

// SendTestSMS sends a test SMS (admin only)
func SendTestSMS(c *fiber.Ctx) error {
	var req struct {
		DeviceID string `json:"device_id" validate:"required"`
		Target   string `json:"target" validate:"required"`
		Message  string `json:"message" validate:"required"`
		SimSlot  int    `json:"sim_slot"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get admin user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Check if device exists and is available
	var device models.Device
	if err := database.DB.Where("device_id = ?", req.DeviceID).First(&device).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
		})
	}

	if !device.IsReadyForSMS() {
		return c.Status(400).JSON(fiber.Map{
			"error": "Device is not ready for SMS",
			"details": fmt.Sprintf("Device status: active=%v, available=%v, operator=%s, battery=%d%%",
				device.IsActive, device.IsAvailable, device.OperatorStatus, device.BatteryLevel),
		})
	}

	// Create test SMS message
	testMessage := models.SMSMessage{
		DeviceID:      device.DeviceID,
		Type:          "outgoing",
		Target:        req.Target,
		Message:       fmt.Sprintf("[TEST] %s", req.Message),
		SimSlot:       req.SimSlot,
		Status:        "pending",
		Priority:      5, // Highest priority for test messages
		IsTestMessage: true,
		AdminUserID:   &userID,
		Timestamp:     time.Now().Unix(),
	}

	if err := database.DB.Create(&testMessage).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create test SMS record",
		})
	}

	// Send test SMS to device
	if err := sendSMSToDevice(device.DeviceID, testMessage); err != nil {
		testMessage.Status = "failed"
		testMessage.ErrorMessage = err.Error()
		database.DB.Save(&testMessage)

		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to send test SMS",
			"details": err.Error(),
		})
	}

	testMessage.Status = "sent"
	database.DB.Save(&testMessage)

	return c.JSON(fiber.Map{
		"success":    true,
		"message_id": testMessage.ID,
		"message":    "Test SMS sent successfully",
	})
}

// SendTestCommand sends a test command to device (admin only)
func SendTestCommand(c *fiber.Ctx) error {
	var req struct {
		DeviceID    string                 `json:"device_id" validate:"required"`
		CommandType string                 `json:"command_type" validate:"required"` // "ussd", "alarm", "status_request"
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if device exists
	var device models.Device
	if err := database.DB.Where("device_id = ?", req.DeviceID).First(&device).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
		})
	}

	if device.OperatorStatus != "online" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Device is not online",
		})
	}

	// Create command based on type
	var command interface{}
	switch req.CommandType {
	case "ussd":
		ussdCode, ok := req.Payload["ussd_code"].(string)
		if !ok {
			return c.Status(400).JSON(fiber.Map{
				"error": "USSD code is required for USSD command",
			})
		}
		command = map[string]interface{}{
			"type": "send_ussd",
			"payload": map[string]interface{}{
				"ussd_code": ussdCode,
				"sim_slot":  req.Payload["sim_slot"],
			},
		}
	case "alarm":
		command = map[string]interface{}{
			"type": "test_alarm",
			"payload": map[string]interface{}{
				"title":   "Test Alarm from Admin",
				"message": req.Payload["message"],
				"type":    "server",
			},
		}
	case "status_request":
		command = map[string]interface{}{
			"type": "status_request",
			"payload": map[string]interface{}{
				"request_battery":  true,
				"request_location": true,
				"request_sim_info": true,
			},
		}
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid command type",
		})
	}

	// Send command to device via WebSocket
	if err := sendCommandToDevice(device.DeviceID, command); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to send command to device",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("%s command sent successfully to device %s", req.CommandType, device.DeviceID),
	})
}

// ProcessDeliveryReport processes SMS delivery reports from devices
func ProcessDeliveryReport(c *fiber.Ctx) error {
	var dlr struct {
		DeviceID       string `json:"device_id"`
		InternalLogID  int    `json:"internal_log_id"`
		Status         string `json:"status"` // "delivered", "failed"
		DeliveryReport string `json:"delivery_report"`
		ErrorMessage   string `json:"error_message"`
		Timestamp      int64  `json:"timestamp"`
	}

	if err := c.BodyParser(&dlr); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid delivery report format",
		})
	}

	// Find SMS message by device ID and internal log ID
	var smsMessage models.SMSMessage
	result := database.DB.Where("device_id = ? AND internal_log_id = ?", dlr.DeviceID, dlr.InternalLogID).
		First(&smsMessage)

	if result.Error != nil {
		log.Printf("SMS message not found for DLR: device=%s, log_id=%d", dlr.DeviceID, dlr.InternalLogID)
		return c.Status(404).JSON(fiber.Map{
			"error": "SMS message not found",
		})
	}

	// Update SMS message status
	smsMessage.Status = dlr.Status
	smsMessage.DeliveryReport = dlr.DeliveryReport
	smsMessage.ErrorMessage = dlr.ErrorMessage

	if dlr.Status == "delivered" {
		now := time.Now()
		smsMessage.DeliveredAt = &now
	}

	if err := database.DB.Save(&smsMessage).Error; err != nil {
		log.Printf("Failed to update SMS message: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update SMS status",
		})
	}

	// Send delivery report to RabbitMQ queue
	go func() {
		dlrData := map[string]interface{}{
			"message_id":      smsMessage.ID,
			"device_id":       dlr.DeviceID,
			"target":          smsMessage.Target,
			"status":          dlr.Status,
			"delivery_report": dlr.DeliveryReport,
			"error_message":   dlr.ErrorMessage,
			"delivered_at":    smsMessage.DeliveredAt,
			"is_test_message": smsMessage.IsTestMessage,
			"timestamp":       dlr.Timestamp,
		}

		if err := queue.PublishToQueue("deliveryreport", dlrData); err != nil {
			log.Printf("Failed to publish DLR to queue: %v", err)
		}
	}()

	log.Printf("DLR processed: message_id=%d, status=%s", smsMessage.ID, dlr.Status)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Delivery report processed",
	})
}

// Helper functions

// findBestDeviceForSMS finds the best available device and SIM for sending SMS
func findBestDeviceForSMS(targetNumber, country, operator string) (*models.Device, *models.SIMCard, error) {
	// Determine target country from phone number if not provided
	if country == "" {
		country = getCountryFromPhoneNumber(targetNumber)
	}

	// Build query for finding suitable devices
	query := database.DB.Joins("JOIN device_groups ON devices.device_group_id = device_groups.id").
		Joins("JOIN sites ON device_groups.site_id = sites.id").
		Where("devices.is_active = ? AND devices.is_available = ? AND devices.operator_status = ? AND devices.battery_level >= ?",
			true, true, "online", 10)

	// Filter by country if specified
	if country != "" {
		query = query.Where("sites.country = ?", country)
	}

	// Filter by operator if specified
	if operator != "" {
		query = query.Where("device_groups.operator = ?", operator)
	}

	// Order by priority: battery level desc, signal strength desc, last seen desc
	query = query.Order("devices.battery_level DESC, devices.signal_strength DESC, devices.last_seen DESC")

	var devices []models.Device
	if err := query.Preload("SIMCards").Find(&devices).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to query devices: %v", err)
	}

	// Find device with available SIM card
	for _, device := range devices {
		for _, simCard := range device.SIMCards {
			if simCard.IsActive && simCard.IsEnabled && simCard.SignalStrength > 0 {
				return &device, &simCard, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("no available device with active SIM card found")
}

// sendSMSToDevice sends SMS to device via WebSocket
func sendSMSToDevice(deviceID string, smsMessage models.SMSMessage) error {
	if wsHub == nil {
		return fmt.Errorf("WebSocket hub not initialized")
	}

	message := map[string]interface{}{
		"type": "send_sms",
		"payload": map[string]interface{}{
			"target":          smsMessage.Target,
			"message":         smsMessage.Message,
			"sim_slot":        smsMessage.SimSlot,
			"internal_log_id": smsMessage.ID, // Use message ID as internal log ID
		},
	}

	return wsHub.SendMessageToDevice(deviceID, message)
}

// sendCommandToDevice sends command to device via WebSocket
func sendCommandToDevice(deviceID string, command interface{}) error {
	if wsHub == nil {
		return fmt.Errorf("WebSocket hub not initialized")
	}

	return wsHub.SendMessageToDevice(deviceID, command)
}

// isValidPhoneNumber validates phone number format
func isValidPhoneNumber(phoneNumber string) bool {
	// Remove spaces and common separators
	cleaned := strings.ReplaceAll(phoneNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Check if it starts with + and has at least 10 digits
	if strings.HasPrefix(cleaned, "+") && len(cleaned) >= 11 {
		for _, char := range cleaned[1:] {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	}

	return false
}

// getCountryFromPhoneNumber determines country from phone number
func getCountryFromPhoneNumber(phoneNumber string) string {
	if !strings.HasPrefix(phoneNumber, "+") {
		return ""
	}

	// Simple country code mapping
	countryMap := map[string]string{
		"+90": "TR", // Turkey
		"+1":  "US", // USA/Canada
		"+44": "UK", // UK
		"+49": "DE", // Germany
		"+33": "FR", // France
		// Add more as needed
	}

	for code, country := range countryMap {
		if strings.HasPrefix(phoneNumber, code) {
			return country
		}
	}

	return ""
}

// calculateSMSCost calculates estimated SMS cost
func calculateSMSCost(target string, messageLength int) float64 {
	// Simple cost calculation - 160 characters = 1 SMS unit
	units := (messageLength + 159) / 160

	// Base cost per SMS unit (example pricing)
	baseCost := 0.05 // $0.05 per SMS unit

	// International SMS cost multiplier
	if strings.HasPrefix(target, "+") && !strings.HasPrefix(target, "+90") {
		baseCost *= 3.0 // 3x cost for international
	}

	return float64(units) * baseCost
}
