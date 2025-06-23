package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Site represents a physical location or SMS hub
type Site struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Country     string    `json:"country" gorm:"not null"`    // TR, US, UK, etc.
	CountryName string    `json:"country_name"`               // Turkey, United States, etc.
	PhoneCode   string    `json:"phone_code" gorm:"not null"` // +90, +1, +44, etc.
	Address     string    `json:"address"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	ManagerName string    `json:"manager_name"`
	ContactInfo string    `json:"contact_info"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	DeviceGroups []DeviceGroup `json:"device_groups" gorm:"foreignKey:SiteID"`
}

// DeviceGroup represents a group of devices (by operator, floor, department, etc.)
type DeviceGroup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SiteID      uint      `json:"site_id" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	GroupType   string    `json:"group_type"` // operator, floor, department, area
	Operator    string    `json:"operator"`   // Turkcell, Vodafone, Türk Telekom, etc.
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Site    Site     `json:"site" gorm:"foreignKey:SiteID"`
	Devices []Device `json:"devices" gorm:"foreignKey:DeviceGroupID"`
}

// Device represents an Android SMS Gateway device
type Device struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	DeviceGroupID  *uint          `json:"device_group_id"`
	DeviceID       string         `json:"device_id" gorm:"uniqueIndex;not null"`
	DeviceName     string         `json:"device_name" gorm:"not null"`
	Model          string         `json:"model"`
	AndroidVersion string         `json:"android_version"`
	AppVersion     string         `json:"app_version"`
	ConnectKey     string         `json:"connect_key" gorm:"uniqueIndex;not null"`
	BatteryLevel   int            `json:"battery_level" gorm:"default:0"`         // 0-100
	BatteryStatus  string         `json:"battery_status"`                         // charging, discharging, full
	OperatorStatus string         `json:"operator_status" gorm:"default:offline"` // online, offline, connecting
	SignalStrength int            `json:"signal_strength" gorm:"default:0"`       // 0-4 or 0-100
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	IsAvailable    bool           `json:"is_available" gorm:"default:false"` // Available for SMS sending
	LastSeen       time.Time      `json:"last_seen"`
	IPAddress      string         `json:"ip_address"`
	Location       string         `json:"location"`
	Latitude       float64        `json:"latitude"`
	Longitude      float64        `json:"longitude"`
	SiteName       string         `json:"site_name"`  // Legacy field
	GroupName      string         `json:"group_name"` // Legacy field
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	DeviceGroup    *DeviceGroup   `json:"device_group" gorm:"foreignKey:DeviceGroupID"`
	SIMCards       []SIMCard      `json:"sim_cards" gorm:"foreignKey:DeviceID;references:DeviceID"`
	DeviceStatuses []DeviceStatus `json:"device_statuses" gorm:"foreignKey:DeviceID;references:DeviceID"`
	SMSMessages    []SMSMessage   `json:"sms_messages" gorm:"foreignKey:DeviceID;references:DeviceID"`
	USSDCommands   []USSDCommand  `json:"ussd_commands" gorm:"foreignKey:DeviceID;references:DeviceID"`
	Alarms         []Alarm        `json:"alarms" gorm:"foreignKey:DeviceID;references:DeviceID"`
}

// IsReadyForSMS checks if device is ready to send SMS
func (d *Device) IsReadyForSMS() bool {
	return d.IsActive &&
		d.IsAvailable &&
		d.OperatorStatus == "online" &&
		d.BatteryLevel >= 10 // Minimum 10% battery required
}

// SIMCard represents a SIM card
type SIMCard struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	DeviceID       string         `json:"device_id" gorm:"not null"`
	Identifier     string         `json:"identifier"`
	IMSI           string         `json:"imsi"`
	IMEI           string         `json:"imei"`
	Operator       string         `json:"operator"`
	PhoneNumber    string         `json:"phone_number"`
	SignalStrength int            `json:"signal_strength"`
	NetworkType    string         `json:"network_type"`
	MCC            string         `json:"mcc"`
	MNC            string         `json:"mnc"`
	IsActive       bool           `json:"is_active"`
	IsEnabled      bool           `json:"is_enabled" gorm:"default:true"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Device *Device `json:"device" gorm:"foreignKey:DeviceID;references:DeviceID"`
}

// DeviceStatus represents device status updates
type DeviceStatus struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	DeviceID      string    `json:"device_id" gorm:"not null"`
	BatteryLevel  int       `json:"battery_level"`
	BatteryStatus string    `json:"battery_status"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Timestamp     int64     `json:"timestamp"`
	CreatedAt     time.Time `json:"created_at"`

	// Relations
	Device *Device `json:"device" gorm:"foreignKey:DeviceID;references:DeviceID"`
}

// SMSMessage represents SMS messages
type SMSMessage struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	DeviceID       string     `json:"device_id" gorm:"not null"`
	Type           string     `json:"type"` // "incoming", "outgoing", "delivery_report"
	Target         string     `json:"target"`
	From           string     `json:"from"`
	Message        string     `json:"message"`
	SimSlot        int        `json:"sim_slot"`
	InternalLogID  int        `json:"internal_log_id"`
	Status         string     `json:"status" gorm:"default:pending"` // "pending", "sent", "delivered", "failed"
	DeliveryReport string     `json:"delivery_report"`
	ErrorMessage   string     `json:"error_message"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	Priority       int        `json:"priority" gorm:"default:1"` // 1-5, higher is more priority
	Retries        int        `json:"retries" gorm:"default:0"`
	MaxRetries     int        `json:"max_retries" gorm:"default:3"`
	ScheduledAt    *time.Time `json:"scheduled_at"`                         // For scheduled SMS
	IsTestMessage  bool       `json:"is_test_message" gorm:"default:false"` // Admin test messages
	AdminUserID    *uint      `json:"admin_user_id"`                        // Who sent the test message
	Timestamp      int64      `json:"timestamp"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Device    *Device `json:"device" gorm:"foreignKey:DeviceID;references:DeviceID"`
	AdminUser *User   `json:"admin_user" gorm:"foreignKey:AdminUserID"`
}

// USSDCommand represents USSD commands
type USSDCommand struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	DeviceID      string    `json:"device_id" gorm:"not null"`
	USSDCode      string    `json:"ussd_code"`
	SimSlot       int       `json:"sim_slot"`
	InternalLogID int       `json:"internal_log_id"`
	Result        string    `json:"result"`
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message"`
	Status        string    `json:"status"` // "pending", "completed", "failed"
	Timestamp     int64     `json:"timestamp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relations
	Device *Device `json:"device" gorm:"foreignKey:DeviceID;references:DeviceID"`
}

// Alarm represents alarms from devices or server
type Alarm struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DeviceID  string    `json:"device_id"`
	Type      string    `json:"type"` // "client", "server"
	AlarmType string    `json:"alarm_type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"` // "low", "medium", "high", "critical"
	Resolved  bool      `json:"resolved" gorm:"default:false"`
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Device *Device `json:"device" gorm:"foreignKey:DeviceID;references:DeviceID"`
}

// User represents system users
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	UserRoles []UserRole `json:"user_roles" gorm:"foreignKey:UserID"`
	Sessions  []Session  `json:"sessions" gorm:"foreignKey:UserID"`
}

// HashPassword hashes the user password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Session represents user sessions
type Session struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	AccessToken      string    `json:"access_token" gorm:"uniqueIndex;not null"`
	RefreshToken     string    `json:"refresh_token" gorm:"uniqueIndex;not null"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	IPAddress        string    `json:"ip_address"`
	UserAgent        string    `json:"user_agent"`
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relations
	User *User `json:"user" gorm:"foreignKey:UserID"`
}

// Role represents user roles
type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	UserRoles       []UserRole       `json:"user_roles" gorm:"foreignKey:RoleID"`
	RolePermissions []RolePermission `json:"role_permissions" gorm:"foreignKey:RoleID"`
}

// Permission represents system permissions
type Permission struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	Resource    string         `json:"resource"`
	Action      string         `json:"action"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	RolePermissions []RolePermission `json:"role_permissions" gorm:"foreignKey:PermissionID"`
}

// UserRole represents user-role relationships
type UserRole struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	RoleID    uint      `json:"role_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User *User `json:"user" gorm:"foreignKey:UserID"`
	Role *Role `json:"role" gorm:"foreignKey:RoleID"`

	// Composite index
	gorm.Model `gorm:"-"`
}

// RolePermission represents role-permission relationships
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RoleID       uint      `json:"role_id" gorm:"not null"`
	PermissionID uint      `json:"permission_id" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relations
	Role       *Role       `json:"role" gorm:"foreignKey:RoleID"`
	Permission *Permission `json:"permission" gorm:"foreignKey:PermissionID"`

	// Composite index
	gorm.Model `gorm:"-"`
}

// Region represents a geographical region
type Region struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"type:varchar(100);not null"`
	Translations string     `json:"translations" gorm:"type:text"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP;not null"`
	Flag         int16      `json:"flag" gorm:"default:1;not null"`
	WikiDataID   *string    `json:"wikiDataId" gorm:"column:wikiDataId;type:varchar(255)"`

	// Relations
	Subregions []Subregion `json:"subregions" gorm:"foreignKey:RegionID"`
	Countries  []Country   `json:"countries" gorm:"foreignKey:RegionID"`
}

// Subregion represents a geographical subregion
type Subregion struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"type:varchar(100);not null"`
	Translations string     `json:"translations" gorm:"type:text"`
	RegionID     int64      `json:"region_id" gorm:"not null"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP;not null"`
	Flag         int16      `json:"flag" gorm:"default:1;not null"`
	WikiDataID   *string    `json:"wikiDataId" gorm:"column:wikiDataId;type:varchar(255)"`

	// Relations
	Region    *Region   `json:"region" gorm:"foreignKey:RegionID"`
	Countries []Country `json:"countries" gorm:"foreignKey:SubregionID"`
}

// Country represents a country
type Country struct {
	ID             int64      `json:"id" gorm:"primaryKey"`
	Name           string     `json:"name" gorm:"type:varchar(100);not null"`
	ISO3           *string    `json:"iso3" gorm:"type:char(3)"`
	NumericCode    *string    `json:"numeric_code" gorm:"type:char(3)"`
	ISO2           *string    `json:"iso2" gorm:"type:char(2)"`
	PhoneCode      *string    `json:"phonecode" gorm:"type:varchar(255)"`
	Capital        *string    `json:"capital" gorm:"type:varchar(255)"`
	Currency       *string    `json:"currency" gorm:"type:varchar(255)"`
	CurrencyName   *string    `json:"currency_name" gorm:"type:varchar(255)"`
	CurrencySymbol *string    `json:"currency_symbol" gorm:"type:varchar(255)"`
	TLD            *string    `json:"tld" gorm:"type:varchar(255)"`
	Native         *string    `json:"native" gorm:"type:varchar(255)"`
	Region         *string    `json:"region" gorm:"type:varchar(255)"`
	RegionID       *int64     `json:"region_id"`
	Subregion      *string    `json:"subregion" gorm:"type:varchar(255)"`
	SubregionID    *int64     `json:"subregion_id"`
	Nationality    *string    `json:"nationality" gorm:"type:varchar(255)"`
	Timezones      *string    `json:"timezones" gorm:"type:text"`
	Translations   *string    `json:"translations" gorm:"type:text"`
	Latitude       *float64   `json:"latitude" gorm:"type:numeric(10,8)"`
	Longitude      *float64   `json:"longitude" gorm:"type:numeric(11,8)"`
	Emoji          *string    `json:"emoji" gorm:"type:varchar(191)"`
	EmojiU         *string    `json:"emojiU" gorm:"type:varchar(191)"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP;not null"`
	Flag           int16      `json:"flag" gorm:"default:1;not null"`
	WikiDataID     *string    `json:"wikiDataId" gorm:"column:wikiDataId;type:varchar(255)"`

	// Relations
	RegionModel    *Region    `json:"region_model" gorm:"foreignKey:RegionID"`
	SubregionModel *Subregion `json:"subregion_model" gorm:"foreignKey:SubregionID"`
	States         []State    `json:"states" gorm:"foreignKey:CountryID"`
	Cities         []City     `json:"cities" gorm:"foreignKey:CountryID"`
}

// State represents a state/province
type State struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	CountryID   int64      `json:"country_id" gorm:"not null"`
	CountryCode string     `json:"country_code" gorm:"type:char(2);not null"`
	FipsCode    *string    `json:"fips_code" gorm:"type:varchar(255)"`
	ISO2        *string    `json:"iso2" gorm:"type:varchar(255)"`
	Type        *string    `json:"type" gorm:"type:varchar(191)"`
	Level       *int       `json:"level"`
	ParentID    *int       `json:"parent_id"`
	Native      *string    `json:"native" gorm:"type:varchar(255)"`
	Latitude    *float64   `json:"latitude" gorm:"type:numeric(10,8)"`
	Longitude   *float64   `json:"longitude" gorm:"type:numeric(11,8)"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP;not null"`
	Flag        int16      `json:"flag" gorm:"default:1;not null"`
	WikiDataID  *string    `json:"wikiDataId" gorm:"column:wikiDataId;type:varchar(255)"`

	// Relations
	Country *Country `json:"country" gorm:"foreignKey:CountryID"`
	Cities  []City   `json:"cities" gorm:"foreignKey:StateID"`
}

// City represents a city
type City struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	StateID     int64     `json:"state_id" gorm:"not null"`
	StateCode   string    `json:"state_code" gorm:"type:varchar(255);not null"`
	CountryID   int64     `json:"country_id" gorm:"not null"`
	CountryCode string    `json:"country_code" gorm:"type:char(2);not null"`
	Latitude    float64   `json:"latitude" gorm:"type:numeric(10,8);not null"`
	Longitude   float64   `json:"longitude" gorm:"type:numeric(11,8);not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:'2014-01-01 12:01:01';not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP;not null"`
	Flag        int16     `json:"flag" gorm:"default:1;not null"`
	WikiDataID  *string   `json:"wikiDataId" gorm:"column:wikiDataId;type:varchar(255)"`

	// Relations
	State   *State   `json:"state" gorm:"foreignKey:StateID"`
	Country *Country `json:"country" gorm:"foreignKey:CountryID"`
}
