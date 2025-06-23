package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"tsimserver/cache"
	"tsimserver/database"
	"tsimserver/models"
	"tsimserver/queue"
	"tsimserver/types"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Client represents a WebSocket client
type Client struct {
	ID       string
	DeviceID string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	LastSeen time.Time
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	Clients map[string]*Client

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast message to all clients
	Broadcast chan []byte

	// Send message to specific device
	SendToDevice chan DeviceMessage

	// Mutex for thread safety
	mutex sync.RWMutex
}

// DeviceMessage represents a message to be sent to a specific device
type DeviceMessage struct {
	DeviceID string
	Message  []byte
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		Clients:      make(map[string]*Client),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Broadcast:    make(chan []byte),
		SendToDevice: make(chan DeviceMessage),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client.ID] = client
			h.mutex.Unlock()

			// Store connection in Redis
			cache.SetDeviceConnection(client.DeviceID, client.ID)
			cache.SetDeviceStatus(client.DeviceID, "online")

			log.Printf("Client registered: %s (Device: %s)", client.ID, client.DeviceID)

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)

				// Remove connection from Redis
				cache.RemoveDeviceConnection(client.DeviceID)
				cache.SetDeviceStatus(client.DeviceID, "offline")

				log.Printf("Client unregistered: %s (Device: %s)", client.ID, client.DeviceID)
			}
			h.mutex.Unlock()

		case message := <-h.Broadcast:
			h.mutex.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.ID)
				}
			}
			h.mutex.RUnlock()

		case deviceMsg := <-h.SendToDevice:
			h.sendToDevice(deviceMsg.DeviceID, deviceMsg.Message)
		}
	}
}

// sendToDevice sends message to specific device
func (h *Hub) sendToDevice(deviceID string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, client := range h.Clients {
		if client.DeviceID == deviceID {
			select {
			case client.Send <- message:
				log.Printf("Message sent to device %s", deviceID)
			default:
				close(client.Send)
				delete(h.Clients, client.ID)
				log.Printf("Failed to send message to device %s, client disconnected", deviceID)
			}
			return
		}
	}

	log.Printf("Device %s not found for message delivery", deviceID)
}

// SendMessageToDevice sends message to specific device by device ID
func (h *Hub) SendMessageToDevice(deviceID string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.SendToDevice <- DeviceMessage{
		DeviceID: deviceID,
		Message:  data,
	}

	return nil
}

// BroadcastMessage sends message to all connected clients
func (h *Hub) BroadcastMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.Broadcast <- data
	return nil
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error for client %s: %v", c.ID, err)
			break
		}

		c.LastSeen = time.Now()
		if err := c.handleMessage(message); err != nil {
			log.Printf("Error handling message from client %s: %v", c.ID, err)
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error for client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(data []byte) error {
	var msg types.WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	switch msg.Type {
	case "auth":
		return c.handleAuth(msg.Data)
	case "device_registration":
		return c.handleDeviceRegistration(msg.Data)
	case "device_status":
		return c.handleDeviceStatus(msg.Data)
	case "incoming_sms":
		return c.handleIncomingSMS(msg.Data)
	case "sms_delivery_report":
		return c.handleSMSDeliveryReport(msg.Data)
	case "ussd_result":
		return c.handleUSSDResult(msg.Data)
	case "phone_number_result":
		return c.handlePhoneNumberResult(msg.Data)
	case "alarm":
		return c.handleClientAlarm(msg.Data)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}

	return nil
}

// handleAuth handles authentication requests
func (c *Client) handleAuth(data json.RawMessage) error {
	var authReq types.AuthRequest
	if err := json.Unmarshal(data, &authReq); err != nil {
		return err
	}

	// Find device by connect key
	var device models.Device
	if err := database.DB.Where("connect_key = ?", authReq.ConnectKey).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Send auth failure
			response := types.AuthResponse{
				Type:    "auth_response",
				Success: false,
			}
			return c.sendMessage(response)
		}
		return err
	}

	// Update client with device info
	c.DeviceID = device.DeviceID

	// Update device last seen
	device.LastSeen = time.Now()
	database.DB.Save(&device)

	// Send auth success
	response := types.AuthResponse{
		Type:       "auth_response",
		Success:    true,
		SiteName:   device.SiteName,
		GroupName:  device.GroupName,
		DeviceName: device.DeviceName,
	}

	return c.sendMessage(response)
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return fmt.Errorf("client send channel is full")
	}
}

// handleDeviceRegistration handles device registration
func (c *Client) handleDeviceRegistration(data json.RawMessage) error {
	var deviceReg types.DeviceRegistration
	if err := json.Unmarshal(data, &deviceReg); err != nil {
		return err
	}

	// Update device info
	var device models.Device
	if err := database.DB.Where("device_id = ?", deviceReg.Payload.DeviceID).First(&device).Error; err != nil {
		return err
	}

	device.DeviceName = deviceReg.Payload.DeviceName
	device.Model = deviceReg.Payload.Model
	device.AndroidVersion = deviceReg.Payload.AndroidVersion
	device.AppVersion = deviceReg.Payload.AppVersion
	device.LastSeen = time.Now()

	if err := database.DB.Save(&device).Error; err != nil {
		return err
	}

	// Update SIM cards
	if err := c.updateSIMCards(deviceReg.Payload.DeviceID, deviceReg.Payload.SIMCards); err != nil {
		return err
	}

	// Save device status
	return c.saveDeviceStatus(deviceReg.Payload)
}

// updateSIMCards updates device SIM cards
func (c *Client) updateSIMCards(deviceID string, simCards []types.SIMCardInfo) error {
	// Delete existing SIM cards
	database.DB.Where("device_id = ?", deviceID).Delete(&models.SIMCard{})

	// Add new SIM cards
	for _, simInfo := range simCards {
		simCard := models.SIMCard{
			DeviceID:       deviceID,
			Identifier:     simInfo.Identifier,
			IMSI:           simInfo.IMSI,
			IMEI:           simInfo.IMEI,
			Operator:       simInfo.Operator,
			PhoneNumber:    simInfo.PhoneNumber,
			SignalStrength: simInfo.SignalStrength,
			NetworkType:    simInfo.NetworkType,
			MCC:            simInfo.MCC,
			MNC:            simInfo.MNC,
			IsActive:       simInfo.IsActive,
		}

		if err := database.DB.Create(&simCard).Error; err != nil {
			return err
		}
	}

	return nil
}

// saveDeviceStatus saves device status to database
func (c *Client) saveDeviceStatus(payload types.DeviceRegistrationPayload) error {
	status := models.DeviceStatus{
		DeviceID:      payload.DeviceID,
		BatteryLevel:  payload.BatteryLevel,
		BatteryStatus: payload.BatteryStatus,
		Latitude:      payload.Latitude,
		Longitude:     payload.Longitude,
		Timestamp:     payload.Timestamp,
	}

	return database.DB.Create(&status).Error
}

// handleDeviceStatus handles device status updates
func (c *Client) handleDeviceStatus(data json.RawMessage) error {
	var deviceStatus types.DeviceStatus
	if err := json.Unmarshal(data, &deviceStatus); err != nil {
		return err
	}

	// Update SIM cards
	if err := c.updateSIMCards(deviceStatus.Payload.DeviceID, deviceStatus.Payload.SIMCards); err != nil {
		return err
	}

	// Save device status
	return c.saveDeviceStatus(deviceStatus.Payload)
}

// handleIncomingSMS handles incoming SMS messages
func (c *Client) handleIncomingSMS(data json.RawMessage) error {
	var incomingSMS types.IncomingSMS
	if err := json.Unmarshal(data, &incomingSMS); err != nil {
		return err
	}

	// Save SMS to database
	sms := models.SMSMessage{
		DeviceID:  c.DeviceID,
		Type:      "incoming",
		From:      incomingSMS.From,
		Message:   incomingSMS.Message,
		Timestamp: incomingSMS.Timestamp,
	}

	if err := database.DB.Create(&sms).Error; err != nil {
		return err
	}

	// Publish to queue for processing
	return queue.PublishMessage(queue.SMSQueue, map[string]interface{}{
		"type":      "incoming_sms",
		"device_id": c.DeviceID,
		"from":      incomingSMS.From,
		"message":   incomingSMS.Message,
		"timestamp": incomingSMS.Timestamp,
	})
}

// handleSMSDeliveryReport handles SMS delivery reports
func (c *Client) handleSMSDeliveryReport(data json.RawMessage) error {
	var dlr types.SMSDeliveryReport
	if err := json.Unmarshal(data, &dlr); err != nil {
		return err
	}

	// Update SMS status in database
	var sms models.SMSMessage
	if err := database.DB.Where("device_id = ? AND internal_log_id = ?", c.DeviceID, dlr.ID).First(&sms).Error; err != nil {
		return err
	}

	sms.Status = dlr.Stat
	sms.DeliveryReport = fmt.Sprintf("sub:%d dlvrd:%d submit_date:%s done_date:%s stat:%s err:%s",
		dlr.Sub, dlr.Dlvrd, dlr.SubmitDate, dlr.DoneDate, dlr.Stat, dlr.Err)

	return database.DB.Save(&sms).Error
}

// handleUSSDResult handles USSD command results
func (c *Client) handleUSSDResult(data json.RawMessage) error {
	var ussdResult types.USSDResult
	if err := json.Unmarshal(data, &ussdResult); err != nil {
		return err
	}

	// Update USSD command in database
	var ussd models.USSDCommand
	if err := database.DB.Where("device_id = ? AND internal_log_id = ?", c.DeviceID, ussdResult.InternalLogID).First(&ussd).Error; err != nil {
		return err
	}

	ussd.Success = ussdResult.Success
	ussd.Result = ussdResult.Result
	ussd.ErrorMessage = ussdResult.ErrorMessage
	ussd.Status = "completed"
	ussd.Timestamp = ussdResult.Timestamp

	return database.DB.Save(&ussd).Error
}

// handlePhoneNumberResult handles phone number discovery results
func (c *Client) handlePhoneNumberResult(data json.RawMessage) error {
	var phoneResult types.PhoneNumberResult
	if err := json.Unmarshal(data, &phoneResult); err != nil {
		return err
	}

	// Update phone number in SIM card
	var simCard models.SIMCard
	if err := database.DB.Where("device_id = ? AND id = ?", c.DeviceID, phoneResult.InternalLogID).First(&simCard).Error; err != nil {
		return err
	}

	if phoneResult.Success {
		simCard.PhoneNumber = phoneResult.PhoneNumber
		return database.DB.Save(&simCard).Error
	}

	return nil
}

// handleClientAlarm handles alarms from clients
func (c *Client) handleClientAlarm(data json.RawMessage) error {
	var clientAlarm types.ClientAlarm
	if err := json.Unmarshal(data, &clientAlarm); err != nil {
		return err
	}

	// Save alarm to database
	alarm := models.Alarm{
		DeviceID:  c.DeviceID,
		Type:      "client",
		AlarmType: clientAlarm.AlarmType,
		Message:   clientAlarm.Message,
		Severity:  "medium", // Default severity
		Timestamp: clientAlarm.Timestamp,
	}

	if err := database.DB.Create(&alarm).Error; err != nil {
		return err
	}

	// Publish alarm to queue
	return queue.PublishAlarm(c.DeviceID, clientAlarm.AlarmType, clientAlarm.Message, "medium")
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:       uuid.New().String(),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      hub,
		LastSeen: time.Now(),
	}
}

// GetConnectedClientsCount returns the number of connected clients
func (h *Hub) GetConnectedClientsCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.Clients)
}

// GetActiveChannelsCount returns the number of active devices
func (h *Hub) GetActiveChannelsCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	devices := make(map[string]bool)
	for _, client := range h.Clients {
		if client.DeviceID != "" {
			devices[client.DeviceID] = true
		}
	}
	return len(devices)
}

// GetChannelStats returns statistics for each device
func (h *Hub) GetChannelStats() map[string]int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	deviceStats := make(map[string]int)
	for _, client := range h.Clients {
		if client.DeviceID != "" {
			deviceStats[client.DeviceID]++
		}
	}
	return deviceStats
}
