package handlers

import (
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetDashboardStats returns dashboard statistics
func GetDashboardStats(c *fiber.Ctx) error {
	var stats struct {
		TotalDevices     int64 `json:"total_devices"`
		OnlineDevices    int64 `json:"online_devices"`
		OfflineDevices   int64 `json:"offline_devices"`
		TotalSIMCards    int64 `json:"total_sim_cards"`
		ActiveSIMCards   int64 `json:"active_sim_cards"`
		TotalSMSToday    int64 `json:"total_sms_today"`
		TotalUSSDToday   int64 `json:"total_ussd_today"`
		UnresolvedAlarms int64 `json:"unresolved_alarms"`
		CriticalAlarms   int64 `json:"critical_alarms"`
	}

	// Get device counts
	database.DB.Model(&models.Device{}).Count(&stats.TotalDevices)
	database.DB.Model(&models.Device{}).Where("is_active = ?", true).Count(&stats.OnlineDevices)
	stats.OfflineDevices = stats.TotalDevices - stats.OnlineDevices

	// Get SIM card counts
	database.DB.Model(&models.SIMCard{}).Count(&stats.TotalSIMCards)
	database.DB.Model(&models.SIMCard{}).Where("is_active = ? AND is_enabled = ?", true, true).Count(&stats.ActiveSIMCards)

	// Get today's SMS count
	database.DB.Model(&models.SMSMessage{}).Where("DATE(created_at) = CURRENT_DATE").Count(&stats.TotalSMSToday)

	// Get today's USSD count
	database.DB.Model(&models.USSDCommand{}).Where("DATE(created_at) = CURRENT_DATE").Count(&stats.TotalUSSDToday)

	// Get alarm counts
	database.DB.Model(&models.Alarm{}).Where("resolved = ?", false).Count(&stats.UnresolvedAlarms)
	database.DB.Model(&models.Alarm{}).Where("resolved = ? AND severity = ?", false, "critical").Count(&stats.CriticalAlarms)

	return c.JSON(stats)
}

// GetDeviceStats returns device statistics
func GetDeviceStats(c *fiber.Ctx) error {
	var deviceStats []struct {
		DeviceID   string `json:"device_id"`
		DeviceName string `json:"device_name"`
		SiteName   string `json:"site_name"`
		GroupName  string `json:"group_name"`
		IsActive   bool   `json:"is_active"`
		SMSCount   int64  `json:"sms_count"`
		USSDCount  int64  `json:"ussd_count"`
		AlarmCount int64  `json:"alarm_count"`
		SIMCount   int64  `json:"sim_count"`
	}

	// Get all devices
	var devices []models.Device
	database.DB.Find(&devices)

	for _, device := range devices {
		var stat struct {
			DeviceID   string `json:"device_id"`
			DeviceName string `json:"device_name"`
			SiteName   string `json:"site_name"`
			GroupName  string `json:"group_name"`
			IsActive   bool   `json:"is_active"`
			SMSCount   int64  `json:"sms_count"`
			USSDCount  int64  `json:"ussd_count"`
			AlarmCount int64  `json:"alarm_count"`
			SIMCount   int64  `json:"sim_count"`
		}

		stat.DeviceID = device.DeviceID
		stat.DeviceName = device.DeviceName
		stat.SiteName = device.SiteName
		stat.GroupName = device.GroupName
		stat.IsActive = device.IsActive

		// Get counts for this device
		database.DB.Model(&models.SMSMessage{}).Where("device_id = ?", device.DeviceID).Count(&stat.SMSCount)
		database.DB.Model(&models.USSDCommand{}).Where("device_id = ?", device.DeviceID).Count(&stat.USSDCount)
		database.DB.Model(&models.Alarm{}).Where("device_id = ?", device.DeviceID).Count(&stat.AlarmCount)
		database.DB.Model(&models.SIMCard{}).Where("device_id = ?", device.DeviceID).Count(&stat.SIMCount)

		deviceStats = append(deviceStats, stat)
	}

	return c.JSON(fiber.Map{
		"devices": deviceStats,
		"count":   len(deviceStats),
	})
}
