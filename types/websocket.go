package types

import "encoding/json"

// WebSocketMessage represents a generic WebSocket message
type WebSocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// AuthRequest represents authentication request from client
type AuthRequest struct {
	Type       string `json:"type"`
	ConnectKey string `json:"connectkey"`
}

// AuthResponse represents authentication response to client
type AuthResponse struct {
	Type       string `json:"type"`
	Success    bool   `json:"success"`
	SiteName   string `json:"sitename,omitempty"`
	GroupName  string `json:"groupname,omitempty"`
	DeviceName string `json:"devicename,omitempty"`
}

// SIMCardInfo represents SIM card information
type SIMCardInfo struct {
	Identifier     string `json:"identifier"`
	IMSI           string `json:"imsi"`
	IMEI           string `json:"imei"`
	Operator       string `json:"operator"`
	PhoneNumber    string `json:"phoneNumber"`
	SignalStrength int    `json:"signalStrength"`
	NetworkType    string `json:"networkType"`
	MCC            string `json:"mcc"`
	MNC            string `json:"mnc"`
	IsActive       bool   `json:"isActive"`
}

// DeviceRegistrationPayload represents device registration payload
type DeviceRegistrationPayload struct {
	DeviceID       string        `json:"device_id"`
	DeviceName     string        `json:"device_name"`
	Model          string        `json:"model"`
	AndroidVersion string        `json:"android_version"`
	AppVersion     string        `json:"app_version"`
	BatteryLevel   int           `json:"batteryLevel"`
	BatteryStatus  string        `json:"batteryStatus"`
	Latitude       float64       `json:"latitude"`
	Longitude      float64       `json:"longitude"`
	Timestamp      int64         `json:"timestamp"`
	SIMCards       []SIMCardInfo `json:"simCards"`
}

// DeviceRegistration represents device registration message
type DeviceRegistration struct {
	Type    string                    `json:"type"`
	Payload DeviceRegistrationPayload `json:"payload"`
}

// DeviceStatus represents device status message
type DeviceStatus struct {
	Type    string                    `json:"type"`
	Payload DeviceRegistrationPayload `json:"payload"`
}

// SendSMSCommand represents SMS sending command from server
type SendSMSCommand struct {
	Type          string `json:"type"`
	Target        string `json:"target"`
	SimSlot       int    `json:"simSlot"`
	Message       string `json:"message"`
	InternalLogID int    `json:"internalLogId"`
}

// IncomingSMS represents incoming SMS from client
type IncomingSMS struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// SMSDeliveryReport represents SMS delivery report from client
type SMSDeliveryReport struct {
	Type       string `json:"type"`
	ID         int    `json:"id"`
	SimSlot    int    `json:"simSlot"`
	Sub        int    `json:"sub"`
	Dlvrd      int    `json:"dlvrd"`
	SubmitDate string `json:"submit_date"`
	DoneDate   string `json:"done_date"`
	Stat       string `json:"stat"`
	Err        string `json:"err"`
	Text       string `json:"text"`
}

// USSDCommand represents USSD command from server
type USSDCommand struct {
	Type          string `json:"type"`
	USSDCode      string `json:"ussdCode"`
	SimSlot       int    `json:"simSlot"`
	InternalLogID int    `json:"internalLogId"`
}

// USSDResult represents USSD result from client
type USSDResult struct {
	Type          string `json:"type"`
	InternalLogID int    `json:"internalLogId"`
	Success       bool   `json:"success"`
	Result        string `json:"result"`
	ErrorMessage  string `json:"errorMessage"`
	Timestamp     int64  `json:"timestamp"`
}

// DisableDeviceCommand represents device disable command
type DisableDeviceCommand struct {
	Type     string `json:"type"`
	DeviceID string `json:"deviceId"`
}

// EnableDeviceCommand represents device enable command
type EnableDeviceCommand struct {
	Type     string `json:"type"`
	DeviceID string `json:"deviceId"`
}

// DisableSIMCommand represents SIM disable command
type DisableSIMCommand struct {
	Type     string `json:"type"`
	DeviceID string `json:"deviceId"`
	SimSlot  int    `json:"simSlot"`
}

// EnableSIMCommand represents SIM enable command
type EnableSIMCommand struct {
	Type     string `json:"type"`
	DeviceID string `json:"deviceId"`
	SimSlot  int    `json:"simSlot"`
}

// ClientAlarm represents alarm from client
type ClientAlarm struct {
	Type      string `json:"type"`
	AlarmType string `json:"alarmType"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// ServerAlarm represents alarm from server to client
type ServerAlarm struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// CheckBalanceCommand represents balance check command
type CheckBalanceCommand struct {
	Type          string `json:"type"`
	SimSlot       int    `json:"simSlot"`
	USSDCode      string `json:"ussdCode"`
	InternalLogID int    `json:"internalLogId"`
}

// DiscoverPhoneNumberCommand represents phone number discovery command
type DiscoverPhoneNumberCommand struct {
	Type          string `json:"type"`
	SimSlot       int    `json:"simSlot"`
	USSDCode      string `json:"ussdCode"`
	InternalLogID int    `json:"internalLogId"`
}

// PhoneNumberResult represents phone number discovery result
type PhoneNumberResult struct {
	Type          string `json:"type"`
	InternalLogID int    `json:"internalLogId"`
	Success       bool   `json:"success"`
	PhoneNumber   string `json:"phoneNumber"`
	ErrorMessage  string `json:"errorMessage"`
	Timestamp     int64  `json:"timestamp"`
}
