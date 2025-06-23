# TsimServer - Database Schema Documentation

## Overview

TsimServer features a comprehensive database schema designed for Android SMS Gateway systems. It is managed using GORM ORM with PostgreSQL.

## Table Categories

### 1. 🌍 Geographic Data Tables (World Data)
### 2. 🏢 Site Management Tables 
### 3. 📱 Device Management Tables
### 4. 📧 SMS and Messaging Tables
### 5. 👤 User and Authorization Tables
### 6. 🔔 Alarm and Status Tables

---

## 📋 Detailed Table Descriptions

## 1. 🌍 Geographic Data Tables

### `regions` - World Regions
Defines world regions (Asia, Europe, etc.)

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | **PK** - Unique region ID |
| `name` | `varchar(100)` | **NN** - Region name |
| `translations` | `text` | Translations in JSON format |
| `created_at` | `timestamp` | Creation time |
| `updated_at` | `timestamp` | **NN** - Update time |
| `flag` | `int16` | **Default: 1** - Status flag |
| `wikiDataId` | `varchar(255)` | WikiData reference ID |

**Relationships:**
- `subregions[]` - Subregions (1:N)
- `countries[]` - Countries (1:N)

### `subregions` - Subregions
Defines subcategories of regions

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | **PK** - Unique subregion ID |
| `name` | `varchar(100)` | **NN** - Subregion name |
| `translations` | `text` | Translations in JSON format |
| `region_id` | `int64` | **NN, FK** - Parent region |
| `created_at` | `timestamp` | Creation time |
| `updated_at` | `timestamp` | **NN** - Update time |
| `flag` | `int16` | **Default: 1** - Status flag |
| `wikiDataId` | `varchar(255)` | WikiData reference ID |

**Relationships:**
- `region` - Parent region (N:1)
- `countries[]` - Countries (1:N)

### `countries` - Countries
Detailed information about world countries

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | **PK** - Unique country ID |
| `name` | `varchar(100)` | **NN** - Country name |
| `iso3` | `char(3)` | ISO 3166-1 alpha-3 code |
| `numeric_code` | `char(3)` | Numeric country code |
| `iso2` | `char(2)` | ISO 3166-1 alpha-2 code |
| `phonecode` | `varchar(255)` | Phone country code (+90, +1, etc.) |
| `capital` | `varchar(255)` | Capital city |
| `currency` | `varchar(255)` | Currency code |
| `currency_name` | `varchar(255)` | Currency name |
| `currency_symbol` | `varchar(255)` | Currency symbol |
| `tld` | `varchar(255)` | Top level domain |
| `native` | `varchar(255)` | Native language name |
| `region` | `varchar(255)` | Region name |
| `region_id` | `int64` | **FK** - Region ID |
| `subregion` | `varchar(255)` | Subregion name |
| `subregion_id` | `int64` | **FK** - Subregion ID |
| `nationality` | `varchar(255)` | Nationality |
| `timezones` | `text` | Timezones in JSON format |
| `translations` | `text` | Translations in JSON format |
| `latitude` | `numeric(10,8)` | Latitude |
| `longitude` | `numeric(11,8)` | Longitude |
| `emoji` | `varchar(191)` | Country flag emoji |
| `emojiU` | `varchar(191)` | Unicode emoji |
| `created_at` | `timestamp` | Creation time |
| `updated_at` | `timestamp` | **NN** - Update time |
| `flag` | `int16` | **Default: 1** - Status flag |
| `wikiDataId` | `varchar(255)` | WikiData reference ID |

**Relationships:**
- `region_model` - Parent region (N:1)
- `subregion_model` - Subregion (N:1)
- `states[]` - States/provinces (1:N)
- `cities[]` - Cities (1:N)

### `states` - States/Provinces
State/province information for countries

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | **PK** - Unique state ID |
| `name` | `varchar(255)` | **NN** - State name |
| `country_id` | `int64` | **NN, FK** - Country ID |
| `country_code` | `char(2)` | **NN** - Country code |
| `fips_code` | `varchar(255)` | FIPS code |
| `iso2` | `varchar(255)` | ISO code |
| `type` | `varchar(191)` | State type |
| `level` | `int` | Hierarchy level |
| `parent_id` | `int` | Parent state ID |
| `native` | `varchar(255)` | Native language name |
| `latitude` | `numeric(10,8)` | Latitude |
| `longitude` | `numeric(11,8)` | Longitude |
| `created_at` | `timestamp` | Creation time |
| `updated_at` | `timestamp` | **NN** - Update time |
| `flag` | `int16` | **Default: 1** - Status flag |
| `wikiDataId` | `varchar(255)` | WikiData reference ID |

**Relationships:**
- `country` - Country (N:1)
- `cities[]` - Cities (1:N)

### `cities` - Cities
City information

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | **PK** - Unique city ID |
| `name` | `varchar(255)` | **NN** - City name |
| `state_id` | `int64` | **NN, FK** - State ID |
| `state_code` | `varchar(255)` | **NN** - State code |
| `country_id` | `int64` | **NN, FK** - Country ID |
| `country_code` | `char(2)` | **NN** - Country code |
| `latitude` | `numeric(10,8)` | **NN** - Latitude |
| `longitude` | `numeric(11,8)` | **NN** - Longitude |
| `created_at` | `timestamp` | **NN, Default: 2014-01-01 12:01:01** |
| `updated_at` | `timestamp` | **NN** - Update time |
| `flag` | `int16` | **Default: 1** - Status flag |
| `wikiDataId` | `varchar(255)` | WikiData reference ID |

**Relationships:**
- `state` - State (N:1)
- `country` - Country (N:1)

---

## 2. 🏢 Site Management Tables

### `sites` - SMS Hub Locations
Defines physical SMS hub locations

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique site ID |
| `name` | `string` | **NN** - Site name |
| `description` | `string` | Site description |
| `country` | `string` | **NN** - Country code (TR, US, UK) |
| `country_name` | `string` | Country name (Turkey, United States) |
| `phone_code` | `string` | **NN** - Phone code (+90, +1, +44) |
| `address` | `string` | Physical address |
| `latitude` | `float64` | Latitude coordinate |
| `longitude` | `float64` | Longitude coordinate |
| `manager_name` | `string` | Site manager name |
| `contact_info` | `string` | Contact information |
| `is_active` | `bool` | **Default: true** - Active status |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `device_groups[]` - Device groups (1:N)

**Sample Data:**
- Istanbul Hub (TR)
- Ankara Backup (TR)
- London Hub (UK)

### `device_groups` - Device Groups
Groups devices by operator, floor, department, etc.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique group ID |
| `site_id` | `uint` | **NN, FK** - Parent site |
| `name` | `string` | **NN** - Group name |
| `description` | `string` | Group description |
| `group_type` | `string` | Group type (operator, floor, department, area) |
| `operator` | `string` | Operator name (Turkcell, Vodafone, Türk Telekom) |
| `is_active` | `bool` | **Default: true** - Active status |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `site` - Parent site (N:1)
- `devices[]` - Devices (1:N)

**Sample Data:**
- Turkcell Group, Vodafone Group (for TR sites)
- EE Group, Vodafone UK Group (for UK sites)

---

## 3. 📱 Device Management Tables

### `devices` - Android SMS Gateway Devices
Android devices used for SMS sending

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique device ID |
| `device_group_id` | `uint` | **FK** - Parent device group |
| `device_id` | `string` | **NN, Unique** - Unique device ID |
| `device_name` | `string` | **NN** - Device name |
| `model` | `string` | Device model |
| `android_version` | `string` | Android version |
| `app_version` | `string` | SMS Gateway app version |
| `connect_key` | `string` | **NN, Unique** - Connection key |
| `battery_level` | `int` | **Default: 0** - Battery level (0-100) |
| `battery_status` | `string` | Battery status (charging, discharging, full) |
| `operator_status` | `string` | **Default: offline** - Operator status |
| `signal_strength` | `int` | **Default: 0** - Signal strength (0-4 or 0-100) |
| `is_active` | `bool` | **Default: true** - Active status |
| `is_available` | `bool` | **Default: false** - Ready for SMS sending |
| `last_seen` | `time.Time` | Last seen time |
| `ip_address` | `string` | IP address |
| `location` | `string` | Location description |
| `latitude` | `float64` | Latitude |
| `longitude` | `float64` | Longitude |
| `site_name` | `string` | **Legacy** - Site name |
| `group_name` | `string` | **Legacy** - Group name |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |
| `deleted_at` | `gorm.DeletedAt` | **Soft Delete** - Deletion time |

**Relationships:**
- `device_group` - Device group (N:1)
- `sim_cards[]` - SIM cards (1:N)
- `device_statuses[]` - Status updates (1:N)
- `sms_messages[]` - SMS messages (1:N)
- `ussd_commands[]` - USSD commands (1:N)
- `alarms[]` - Alarms (1:N)

**Important Methods:**
- `IsReadyForSMS()` - Checks if ready for SMS sending

### `sim_cards` - SIM Cards
SIM card information for devices

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique SIM ID |
| `device_id` | `string` | **NN, FK** - Parent device ID |
| `identifier` | `string` | SIM identifier |
| `imsi` | `string` | IMSI number |
| `imei` | `string` | IMEI number |
| `operator` | `string` | Operator name |
| `phone_number` | `string` | Phone number |
| `signal_strength` | `int` | Signal strength |
| `network_type` | `string` | Network type (2G, 3G, 4G, 5G) |
| `mcc` | `string` | Mobile Country Code |
| `mnc` | `string` | Mobile Network Code |
| `is_active` | `bool` | Active status |
| `is_enabled` | `bool` | **Default: true** - Enabled status |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |
| `deleted_at` | `gorm.DeletedAt` | **Soft Delete** - Deletion time |

**Relationships:**
- `device` - Parent device (N:1)

### `device_statuses` - Device Status Updates
Real-time device status information

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique status ID |
| `device_id` | `string` | **NN, FK** - Device ID |
| `battery_level` | `int` | Battery level |
| `battery_status` | `string` | Battery status |
| `latitude` | `float64` | Latitude |
| `longitude` | `float64` | Longitude |
| `timestamp` | `int64` | Unix timestamp |
| `created_at` | `time.Time` | Creation time |

**Relationships:**
- `device` - Parent device (N:1)

---

## 4. 📧 SMS and Messaging Tables

### `sms_messages` - SMS Messages
Incoming, outgoing and delivery reports

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique message ID |
| `device_id` | `string` | **NN, FK** - Sender/receiver device ID |
| `type` | `string` | Message type (incoming, outgoing, delivery_report) |
| `target` | `string` | Target phone number |
| `from` | `string` | Sender phone number |
| `message` | `string` | Message content |
| `sim_slot` | `int` | SIM slot number |
| `internal_log_id` | `int` | Internal log ID |
| `status` | `string` | **Default: pending** - Status (pending, sent, delivered, failed) |
| `delivery_report` | `string` | Delivery report |
| `error_message` | `string` | Error message |
| `delivered_at` | `time.Time` | Delivery time |
| `priority` | `int` | **Default: 1** - Priority (1-5, higher = more priority) |
| `retries` | `int` | **Default: 0** - Retry count |
| `max_retries` | `int` | **Default: 3** - Maximum retry count |
| `scheduled_at` | `time.Time` | Scheduled send time |
| `is_test_message` | `bool` | **Default: false** - Admin test message |
| `admin_user_id` | `uint` | **FK** - Admin user who sent test message |
| `timestamp` | `int64` | Unix timestamp |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `device` - Parent device (N:1)
- `admin_user` - Admin user (N:1)

### `ussd_commands` - USSD Commands
USSD codes and responses

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique command ID |
| `device_id` | `string` | **NN, FK** - Device ID |
| `ussd_code` | `string` | USSD code (*100#, *123#, etc.) |
| `sim_slot` | `int` | SIM slot number |
| `internal_log_id` | `int` | Internal log ID |
| `result` | `string` | USSD response |
| `success` | `bool` | Success status |
| `error_message` | `string` | Error message |
| `status` | `string` | Status (pending, completed, failed) |
| `timestamp` | `int64` | Unix timestamp |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `device` - Parent device (N:1)

---

## 5. 👤 User and Authorization Tables

### `users` - System Users
Users with system access rights

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique user ID |
| `username` | `string` | **NN, Unique** - Username |
| `email` | `string` | **NN, Unique** - Email address |
| `password` | `string` | **NN** - Encrypted password |
| `first_name` | `string` | First name |
| `last_name` | `string` | Last name |
| `is_active` | `bool` | **Default: true** - Active status |
| `last_login` | `time.Time` | Last login time |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |
| `deleted_at` | `gorm.DeletedAt` | **Soft Delete** - Deletion time |

**Relationships:**
- `user_roles[]` - User roles (1:N)
- `sessions[]` - Sessions (1:N)

**Important Methods:**
- `HashPassword()` - Encrypts password
- `CheckPassword(password)` - Password verification

**Default Admin:**
- Username: `admin`
- Password: `admin123`
- Email: `admin@tsimserver.com`

### `sessions` - User Sessions
JWT token-based session management

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique session ID |
| `user_id` | `uint` | **NN, FK** - User ID |
| `access_token` | `string` | **NN, Unique** - Access token |
| `refresh_token` | `string` | **NN, Unique** - Refresh token |
| `expires_at` | `time.Time` | Token expiration time |
| `refresh_expires_at` | `time.Time` | Refresh token expiration time |
| `ip_address` | `string` | IP address |
| `user_agent` | `string` | Browser information |
| `is_active` | `bool` | **Default: true** - Active status |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `user` - Parent user (N:1)

### `roles` - System Roles
User role definitions

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique role ID |
| `name` | `string` | **NN, Unique** - Role name |
| `display_name` | `string` | Display name |
| `description` | `string` | Role description |
| `is_active` | `bool` | **Default: true** - Active status |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |
| `deleted_at` | `gorm.DeletedAt` | **Soft Delete** - Deletion time |

**Relationships:**
- `user_roles[]` - User roles (1:N)
- `role_permissions[]` - Role permissions (1:N)

**Default Roles:**
- `admin` - System Administrator (full access)
- `manager` - Manager (device and SMS management)
- `operator` - Operator (limited operations)
- `viewer` - Viewer (read-only access)

### `permissions` - System Permissions
Granular permission control system

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique permission ID |
| `name` | `string` | **NN, Unique** - Permission name (users.read, sms.write) |
| `display_name` | `string` | Display name |
| `description` | `string` | Permission description |
| `resource` | `string` | Resource name (users, devices, sms) |
| `action` | `string` | Action (read, write, delete, admin) |
| `is_active` | `bool` | **Default: true** - Active status |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |
| `deleted_at` | `gorm.DeletedAt` | **Soft Delete** - Deletion time |

**Relationships:**
- `role_permissions[]` - Role permissions (1:N)

**Permission Categories:**
- **User Management:** users.read, users.write, users.delete
- **Role Management:** roles.read, roles.write, roles.delete
- **Permission Management:** permissions.read, permissions.write, permissions.delete
- **Device Management:** devices.read, devices.write, devices.delete, devices.admin
- **SMS Management:** sms.read, sms.write, sms.delete, sms.admin
- **USSD Management:** ussd.read, ussd.write, ussd.delete
- **Alarm Management:** alarms.read, alarms.write, alarms.delete
- **Statistics:** stats.read
- **Geographic Data:** regions/countries/states/cities/subregions (read, write, delete)
- **Site Management:** sites.read, sites.write, sites.delete
- **Device Groups:** device_groups.read, device_groups.write, device_groups.delete

### `user_roles` - User-Role Relationship
Many-to-many relationship table

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique relationship ID |
| `user_id` | `uint` | **NN, FK** - User ID |
| `role_id` | `uint` | **NN, FK** - Role ID |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `user` - User (N:1)
- `role` - Role (N:1)

### `role_permissions` - Role-Permission Relationship
Many-to-many relationship table

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique relationship ID |
| `role_id` | `uint` | **NN, FK** - Role ID |
| `permission_id` | `uint` | **NN, FK** - Permission ID |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `role` - Role (N:1)
- `permission` - Permission (N:1)

---

## 6. 🔔 Alarm and Status Tables

### `alarms` - System Alarms
Device and server alarms

| Field | Type | Description |
|-------|------|-------------|
| `id` | `uint` | **PK** - Unique alarm ID |
| `device_id` | `string` | **FK** - Related device ID |
| `type` | `string` | Alarm type (client, server) |
| `alarm_type` | `string` | Alarm category |
| `title` | `string` | Alarm title |
| `message` | `string` | Alarm message |
| `severity` | `string` | Severity level (low, medium, high, critical) |
| `resolved` | `bool` | **Default: false** - Resolution status |
| `timestamp` | `int64` | Unix timestamp |
| `created_at` | `time.Time` | Creation time |
| `updated_at` | `time.Time` | Update time |

**Relationships:**
- `device` - Related device (N:1)

---

## 🔗 Table Relationships Summary

### Hierarchical Relationships:
```
Region (1:N) → Subregion (1:N) → Country (1:N) → State (1:N) → City
Site (1:N) → DeviceGroup (1:N) → Device (1:N) → SIMCard
Device (1:N) → SMS/USSD/DeviceStatus/Alarm
User (M:N) → Role (M:N) → Permission
```

### Main Workflows:
1. **Site Management**: Site → DeviceGroup → Device → SIMCard
2. **SMS Routing**: Country/Operator → DeviceGroup → Available Device → SMS Send
3. **Authentication**: User → Role → Permission → Resource Access
4. **Monitoring**: Device → DeviceStatus/Alarm → Alert System

---

## 📊 Data Model Features

### Optimizations:
- **Soft Delete**: For critical data (users, devices, sim_cards)
- **Indexing**: Unique constraints and foreign keys
- **Timestamping**: created_at/updated_at in all tables
- **Composite Indexes**: For user_roles and role_permissions

### Data Integrity:
- **Foreign Key Constraints**: Relational integrity
- **Not Null Constraints**: For critical fields
- **Unique Constraints**: Uniqueness requirements
- **Default Values**: Appropriate default values

### SMS Gateway Special Features:
- **Smart Device Selection**: Operator, battery, signal optimization
- **Admin Test System**: Special flags for test messages
- **Delivery Tracking**: DLR (Delivery Report) tracking
- **Priority & Retry Logic**: Message prioritization and retry mechanism
- **Scheduled SMS**: Scheduled message sending

---

## 🚀 Usage Scenarios

### 1. SMS Sending Workflow:
```sql
-- 1. Find suitable device by country and operator
SELECT d.* FROM devices d 
JOIN device_groups dg ON d.device_group_id = dg.id 
JOIN sites s ON dg.site_id = s.id 
WHERE s.country = 'TR' AND dg.operator = 'Turkcell' 
AND d.is_active = true AND d.is_available = true 
AND d.battery_level >= 10 ORDER BY d.signal_strength DESC;

-- 2. Create SMS message
INSERT INTO sms_messages (device_id, type, target, message, priority) 
VALUES ('device123', 'outgoing', '+905551234567', 'Test message', 1);
```

### 2. Site and Device Management:
```sql
-- Site statistics
SELECT s.name, s.country, 
       COUNT(dg.id) as device_groups,
       COUNT(d.id) as devices,
       COUNT(CASE WHEN d.is_available = true THEN 1 END) as available_devices
FROM sites s 
LEFT JOIN device_groups dg ON s.id = dg.site_id 
LEFT JOIN devices d ON dg.id = d.device_group_id 
GROUP BY s.id;
```

### 3. User Authorization:
```sql
-- Get user permissions
SELECT p.name, p.resource, p.action 
FROM permissions p 
JOIN role_permissions rp ON p.id = rp.permission_id 
JOIN roles r ON rp.role_id = r.id 
JOIN user_roles ur ON r.id = ur.role_id 
WHERE ur.user_id = ? AND p.is_active = true AND r.is_active = true;
```

This schema provides a comprehensive and scalable database structure designed to meet all requirements of the TsimServer project. 