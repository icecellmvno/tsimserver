# TsimServer - Workflow Documentation

## 📋 Overview

TsimServer features complex workflows designed for Android SMS Gateway systems. This documentation explains the main workflows step by step.

## 🔄 Main Workflows

1. [SMS Sending Workflow](#1-sms-sending-workflow)
2. [Device Management Workflow](#2-device-management-workflow)
3. [Site Management Workflow](#3-site-management-workflow)
4. [Authentication & Authorization Workflow](#4-authentication--authorization-workflow)
5. [Admin Test SMS Workflow](#5-admin-test-sms-workflow)
6. [Device Status Monitoring Workflow](#6-device-status-monitoring-workflow)
7. [USSD Command Workflow](#7-ussd-command-workflow)
8. [Alarm Management Workflow](#8-alarm-management-workflow)
9. [Database Migration & Seeding Workflow](#9-database-migration--seeding-workflow)
10. [System Startup Workflow](#10-system-startup-workflow)

---

## 1. 📧 SMS Sending Workflow

### Smart SMS Routing System

#### 1.1 Normal SMS Sending
```
[API Request] → [Validation] → [Device Selection] → [WebSocket Send] → [Response] → [DLR Tracking]
```

**Step-by-Step Process:**

1. **API Request Received:**
   ```http
   POST /api/v1/sms-gateway/send
   Content-Type: application/json
   Authorization: Bearer <token>
   
   {
     "target": "+905551234567",
     "message": "Test message",
     "priority": 1
   }
   ```

2. **Validation:**
   - Phone number format validation
   - Message length validation
   - User permission check
   - Rate limiting check

3. **Smart Device Selection:**
   ```go
   // Extract country code (+90 → TR)
   country := extractCountryFromPhone("+905551234567") // TR
   
   // Find available devices
   devices := findAvailableDevices(country, operator)
   
   // Filtering criteria:
   // - is_active = true
   // - is_available = true  
   // - battery_level >= 10%
   // - operator_status = "online"
   // - signal_strength > 0
   
   // Sort by: battery_level DESC, signal_strength DESC, last_seen DESC
   selectedDevice := selectBestDevice(devices)
   ```

4. **Create SMS Record:**
   ```sql
   INSERT INTO sms_messages (
     device_id, type, target, message, status, priority
   ) VALUES (
     'device123', 'outgoing', '+905551234567', 'Test message', 'pending', 1
   );
   ```

5. **Send via WebSocket:**
   ```json
   {
     "action": "send_sms",
     "data": {
       "target": "+905551234567",
       "message": "Test message",
       "sim_slot": 0,
       "internal_log_id": 12345
     }
   }
   ```

6. **Device Response:**
   ```json
   {
     "action": "sms_result",
     "data": {
       "internal_log_id": 12345,
       "success": true,
       "message": "SMS sent successfully"
     }
   }
   ```

7. **Status Update:**
   ```sql
   UPDATE sms_messages 
   SET status = 'sent', updated_at = NOW() 
   WHERE internal_log_id = 12345;
   ```

#### 1.2 Delivery Report (DLR) Processing
```
[Device DLR] → [Parse DLR] → [Update Status] → [RabbitMQ Publish] → [Notification]
```

**DLR Workflow:**

1. **DLR Received from Device:**
   ```json
   {
     "action": "delivery_report",
     "data": {
       "internal_log_id": 12345,
       "status": "delivered",
       "timestamp": 1640995200
     }
   }
   ```

2. **Update SMS Status:**
   ```sql
   UPDATE sms_messages 
   SET status = 'delivered', 
       delivered_at = FROM_UNIXTIME(1640995200),
       delivery_report = 'Message delivered successfully'
   WHERE internal_log_id = 12345;
   ```

3. **Publish to RabbitMQ:**
   ```go
   publishToQueue("delivery_reports", DeliveryReport{
     MessageID: 12345,
     Status: "delivered",
     DeliveredAt: time.Unix(1640995200, 0),
   })
   ```

#### 1.3 Error Handling and Retry Mechanism
```
[Send Failed] → [Check Retry Count] → [Wait Interval] → [Retry or Mark Failed]
```

**Retry Logic:**

1. **On Error:**
   ```sql
   UPDATE sms_messages 
   SET retries = retries + 1,
       error_message = 'Device timeout',
       status = CASE 
         WHEN retries < max_retries THEN 'pending'
         ELSE 'failed'
       END
   WHERE id = 12345;
   ```

2. **Exponential Backoff:**
   - 1st attempt: Immediate
   - 2nd attempt: 30 seconds later
   - 3rd attempt: 2 minutes later
   - 4th attempt: 5 minutes later (failed if max_retries = 3)

---

## 2. 📱 Device Management Workflow

### 2.1 Device Registration and Connection
```
[Device Connect] → [Authentication] → [Registration] → [Status Update] → [Group Assignment]
```

**Device Connection Process:**

1. **WebSocket Connection:**
   ```
   ws://localhost:8081/ws?connect_key=unique_device_key_123
   ```

2. **Device Authentication:**
   ```go
   func authenticateDevice(connectKey string) (*models.Device, error) {
       var device models.Device
       err := db.Where("connect_key = ? AND is_active = ?", 
                      connectKey, true).First(&device).Error
       return &device, err
   }
   ```

3. **Update Device Status:**
   ```sql
   UPDATE devices 
   SET operator_status = 'online',
       last_seen = NOW(),
       ip_address = ?
   WHERE device_id = ?;
   ```

4. **SIM Card Information Sync:**
   ```json
   {
     "action": "get_sim_info",
     "data": {}
   }
   ```

### 2.2 Periodic Status Updates
```
[Heartbeat] → [Battery Check] → [Signal Check] → [Location Update] → [Database Update]
```

**5-Minute Heartbeat:**

1. **Device Status Information:**
   ```json
   {
     "action": "device_status",
     "data": {
       "battery_level": 85,
       "battery_status": "charging",
       "signal_strength": 4,
       "latitude": 41.0082,
       "longitude": 28.9784
     }
   }
   ```

2. **Database Update:**
   ```sql
   INSERT INTO device_statuses (
     device_id, battery_level, battery_status, 
     latitude, longitude, timestamp
   ) VALUES (?, ?, ?, ?, ?, ?);
   
   UPDATE devices 
   SET battery_level = ?,
       signal_strength = ?,
       latitude = ?,
       longitude = ?,
       last_seen = NOW()
   WHERE device_id = ?;
   ```

### 2.3 Device Availability Check
```
[Availability Check] → [Battery Validation] → [Signal Validation] → [Status Update]
```

**IsReadyForSMS Check:**
```go
func (d *Device) IsReadyForSMS() bool {
    return d.IsActive &&
           d.IsAvailable &&
           d.OperatorStatus == "online" &&
           d.BatteryLevel >= 10 &&
           d.SignalStrength > 0 &&
           time.Since(d.LastSeen) < 10*time.Minute
}
```

---

## 3. 🏢 Site Management Workflow

### 3.1 Site Creation Workflow
```
[Create Site] → [Validate Data] → [Save to DB] → [Create Default Groups] → [Assign Devices]
```

**Site Creation Steps:**

1. **Site Information Input:**
   ```json
   {
     "name": "Izmir Hub",
     "country": "TR",
     "phone_code": "+90",
     "address": "Izmir Technology Park",
     "latitude": 38.4192,
     "longitude": 27.1287,
     "manager_name": "Ali Veli",
     "contact_info": "ali@company.com"
   }
   ```

2. **Validation:**
   - Site name uniqueness check
   - Country code validation
   - GPS coordinate validation
   - Contact information format

3. **Database Save:**
   ```sql
   INSERT INTO sites (name, country, phone_code, address, ...) 
   VALUES (?, ?, ?, ?, ...);
   ```

4. **Create Default Device Groups:**
   ```go
   if site.Country == "TR" {
       createDeviceGroup(site.ID, "Turkcell Group", "Turkcell")
       createDeviceGroup(site.ID, "Vodafone Group", "Vodafone") 
       createDeviceGroup(site.ID, "Turk Telekom Group", "Turk Telekom")
   }
   ```

### 3.2 Device Group Management
```
[Group Creation] → [Operator Assignment] → [Device Migration] → [Statistics Update]
```

**Device Group Operations:**

1. **Group Creation:**
   ```http
   POST /api/v1/device-groups
   {
     "site_id": 1,
     "name": "VIP Turkcell Group",
     "group_type": "operator",
     "operator": "Turkcell"
   }
   ```

2. **Move Devices to Group:**
   ```sql
   UPDATE devices 
   SET device_group_id = ?
   WHERE device_id IN (?, ?, ?) AND is_active = true;
   ```

3. **Group Statistics:**
   ```sql
   SELECT 
     dg.name,
     COUNT(d.id) as total_devices,
     COUNT(CASE WHEN d.is_available = true THEN 1 END) as available_devices,
     AVG(d.battery_level) as avg_battery,
     AVG(d.signal_strength) as avg_signal
   FROM device_groups dg
   LEFT JOIN devices d ON dg.id = d.device_group_id
   WHERE dg.site_id = ?
   GROUP BY dg.id;
   ```

---

## 4. 🔐 Authentication & Authorization Workflow

### 4.1 User Login Workflow
```
[Login Request] → [Credential Validation] → [JWT Generation] → [Session Creation] → [Response]
```

**Login Process:**

1. **Login Request:**
   ```http
   POST /api/v1/auth/login
   {
     "username": "admin",
     "password": "admin123"
   }
   ```

2. **Authentication:**
   ```go
   func authenticateUser(username, password string) (*models.User, error) {
       var user models.User
       err := db.Where("username = ? AND is_active = ?", username, true).First(&user).Error
       if err != nil {
           return nil, err
       }
       
       if !user.CheckPassword(password) {
           return nil, errors.New("invalid password")
       }
       
       return &user, nil
   }
   ```

3. **JWT Token Creation:**
   ```go
   accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
       "user_id": user.ID,
       "username": user.Username,
       "exp": time.Now().Add(24 * time.Hour).Unix(),
   })
   
   refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
       "user_id": user.ID,
       "type": "refresh",
       "exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
   })
   ```

4. **Session Save:**
   ```sql
   INSERT INTO sessions (
     user_id, access_token, refresh_token, 
     expires_at, refresh_expires_at, ip_address, user_agent
   ) VALUES (?, ?, ?, ?, ?, ?, ?);
   ```

### 4.2 Authorization Check
```
[API Request] → [Token Validation] → [Permission Check] → [Resource Access]
```

**Permission Check Process:**

1. **Token Validation:**
   ```go
   func validateToken(tokenString string) (*jwt.Token, error) {
       return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
           return []byte(os.Getenv("JWT_SECRET")), nil
       })
   }
   ```

2. **Get User Permissions:**
   ```sql
   SELECT p.name, p.resource, p.action 
   FROM permissions p 
   JOIN role_permissions rp ON p.id = rp.permission_id 
   JOIN roles r ON rp.role_id = r.id 
   JOIN user_roles ur ON r.id = ur.role_id 
   WHERE ur.user_id = ? AND p.is_active = true;
   ```

3. **Casbin Authorization:**
   ```go
   func checkPermission(userID uint, resource, action string) bool {
       enforcer := casbin.GetEnforcer()
       return enforcer.Enforce(fmt.Sprintf("user:%d", userID), resource, action)
   }
   ```

---

## 5. 🧪 Admin Test SMS Workflow

### 5.1 Test SMS Sending
```
[Admin Request] → [Permission Check] → [Device Selection] → [Test Flag Set] → [Send SMS] → [Track Results]
```

**Admin Test Process:**

1. **Test SMS Request:**
   ```http
   POST /api/v1/sms-gateway/test
   {
     "target": "+905551234567",
     "message": "Test message - Admin",
     "device_id": "specific_device_123" // optional
   }
   ```

2. **Admin Permission Check:**
   ```go
   hasPermission := checkPermission(userID, "sms", "admin")
   if !hasPermission {
       return errors.New("insufficient permissions")
   }
   ```

3. **Create Test SMS:**
   ```sql
   INSERT INTO sms_messages (
     device_id, type, target, message, 
     is_test_message, admin_user_id, priority
   ) VALUES (?, 'outgoing', ?, ?, true, ?, 5);
   ```

4. **Test Results Reporting:**
   ```json
   {
     "test_id": 12345,
     "status": "sent",
     "device_used": "device123",
     "send_time": "2024-01-01T10:00:00Z",
     "delivery_status": "pending"
   }
   ```

### 5.2 USSD Test Commands
```
[USSD Request] → [Device Check] → [Command Send] → [Result Wait] → [Response Parse]
```

**USSD Test Workflow:**

1. **Send USSD Command:**
   ```http
   POST /api/v1/sms-gateway/test-command
   {
     "device_id": "device123",
     "command_type": "ussd",
     "ussd_code": "*100#",
     "sim_slot": 0
   }
   ```

2. **Send Command to Device:**
   ```json
   {
     "action": "ussd_command",
     "data": {
       "ussd_code": "*100#",
       "sim_slot": 0,
       "internal_log_id": 98765
     }
   }
   ```

3. **Wait for Result and Process:**
   ```json
   {
     "action": "ussd_result",
     "data": {
       "internal_log_id": 98765,
       "success": true,
       "result": "Your balance: 15.50 TL"
     }
   }
   ```

---

## 6. 📊 Device Status Monitoring Workflow

### 6.1 Real-time Monitoring
```
[Status Collection] → [Anomaly Detection] → [Alert Generation] → [Notification] → [Auto-healing]
```

**Monitoring Process:**

1. **Status Collection (Every 5 minutes):**
   ```go
   func collectDeviceStatuses() {
       for _, device := range activeDevices {
           status := getDeviceStatus(device.ID)
           
           // Critical checks
           if status.BatteryLevel < 10 {
               generateAlert("LOW_BATTERY", device.ID)
           }
           
           if time.Since(status.LastSeen) > 10*time.Minute {
               generateAlert("DEVICE_OFFLINE", device.ID)
           }
           
           if status.SignalStrength == 0 {
               generateAlert("NO_SIGNAL", device.ID)
           }
       }
   }
   ```

2. **Anomaly Detection:**
   ```go
   func detectAnomalies(device *models.Device) {
       // Battery drain check
       if device.BatteryLevel < previousLevel-20 {
           generateAlert("RAPID_BATTERY_DRAIN", device.ID)
       }
       
       // Signal loss check
       if device.SignalStrength == 0 && previousSignal > 2 {
           generateAlert("SIGNAL_LOST", device.ID)
       }
       
       // Offline check
       if time.Since(device.LastSeen) > 5*time.Minute {
           generateAlert("DEVICE_OFFLINE", device.ID)
       }
   }
   ```

3. **Alert Creation:**
   ```sql
   INSERT INTO alarms (
     device_id, type, alarm_type, title, message, severity
   ) VALUES (
     ?, 'server', 'LOW_BATTERY', 
     'Low Battery Warning', 
     'Device battery level is below 10%', 
     'high'
   );
   ```

### 6.2 Health Check Workflow
```
[Scheduled Check] → [Device Ping] → [Response Validation] → [Status Update] → [Alert if Needed]
```

**Health Check (Every minute):**

1. **Send Ping:**
   ```json
   {
     "action": "ping",
     "data": {
       "timestamp": 1640995200
     }
   }
   ```

2. **Wait for Pong:**
   ```json
   {
     "action": "pong", 
     "data": {
       "timestamp": 1640995200,
       "device_status": "healthy"
     }
   }
   ```

3. **Response Time Calculation:**
   ```go
   responseTime := time.Now().Sub(pingTime)
   if responseTime > 5*time.Second {
       generateAlert("HIGH_LATENCY", device.ID)
   }
   ```

---

## 7. 📞 USSD Command Workflow

### 7.1 USSD Command Processing
```
[USSD Request] → [Validation] → [Device Send] → [Result Wait] → [Response Process] → [Database Save]
```

**USSD Processing Steps:**

1. **USSD Request:**
   ```http
   POST /api/v1/devices/{device_id}/ussd
   {
     "ussd_code": "*100#",
     "sim_slot": 0
   }
   ```

2. **Send to Device:**
   ```json
   {
     "action": "ussd_command",
     "data": {
       "ussd_code": "*100#",
       "sim_slot": 0,
       "internal_log_id": 54321
     }
   }
   ```

3. **Database Record:**
   ```sql
   INSERT INTO ussd_commands (
     device_id, ussd_code, sim_slot, 
     internal_log_id, status
   ) VALUES (?, ?, ?, ?, 'pending');
   ```

4. **Result Processing:**
   ```go
   func processUSSDResult(result USSDResult) {
       command := getUSSDCommand(result.InternalLogID)
       
       command.Result = result.Response
       command.Success = result.Success
       command.Status = "completed"
       
       if !result.Success {
           command.ErrorMessage = result.Error
           command.Status = "failed"
       }
       
       saveUSSDCommand(command)
   }
   ```

---

## 8. 🚨 Alarm Management Workflow

### 8.1 Alarm Creation and Processing
```
[Event Trigger] → [Severity Assessment] → [Alarm Creation] → [Notification] → [Escalation] → [Resolution]
```

**Alarm Lifecycle:**

1. **Trigger Alert:**
   ```go
   func generateAlert(alertType string, deviceID string) {
       severity := getSeverityLevel(alertType)
       
       alarm := models.Alarm{
           DeviceID:  deviceID,
           Type:      "server",
           AlarmType: alertType,
           Title:     getAlarmTitle(alertType),
           Message:   getAlarmMessage(alertType, deviceID),
           Severity:  severity,
           Timestamp: time.Now().Unix(),
       }
       
       database.DB.Create(&alarm)
       notifyAlarm(alarm)
   }
   ```

2. **Send Notification:**
   ```go
   func notifyAlarm(alarm models.Alarm) {
       switch alarm.Severity {
       case "critical":
           sendSMSAlert(alarm)
           sendEmailAlert(alarm)
           sendSlackAlert(alarm)
       case "high":
           sendEmailAlert(alarm)
           sendSlackAlert(alarm)
       case "medium":
           sendSlackAlert(alarm)
       case "low":
           logAlert(alarm)
       }
   }
   ```

3. **Auto-Resolution Check:**
   ```go
   func checkAlarmResolution() {
       openAlarms := getOpenAlarms()
       
       for _, alarm := range openAlarms {
           if shouldAutoResolve(alarm) {
               alarm.Resolved = true
               updateAlarm(alarm)
               
               notifyResolution(alarm)
           }
       }
   }
   ```

---

## 9. 🗄️ Database Migration & Seeding Workflow

### 9.1 Migration Process
```
[Migration Check] → [Backup Creation] → [Schema Update] → [Data Migration] → [Validation] → [Rollback if Failed]
```

**Migration Commands:**
```bash
# Normal migration
make migrate

# Reset and re-migrate
make migrate-reset

# Rollback
make migrate-rollback

# Migration with specific config
./bin/migrate --config=config.yaml --reset
```

**Migration Workflow:**

1. **Create Backup:**
   ```bash
   pg_dump -h localhost -U postgres -d tsimserver > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Run Migration Files:**
   ```go
   func runMigrations() error {
       err := database.DB.AutoMigrate(
           &models.User{},
           &models.Role{},
           &models.Permission{},
           &models.Site{},
           &models.DeviceGroup{},
           &models.Device{},
           // ... other models
       )
       return err
   }
   ```

3. **Migration Validation:**
   ```go
   func validateMigration() error {
       // Table existence check
       if !database.DB.Migrator().HasTable("users") {
           return errors.New("users table not found")
       }
       
       // Constraint check
       if !database.DB.Migrator().HasConstraint(&models.User{}, "fk_user_roles") {
           return errors.New("foreign key constraint missing")
       }
       
       return nil
   }
   ```

### 9.2 Seeding Workflow
```
[Seed Check] → [World Data] → [Auth Data] → [Site Data] → [Verification] → [Report]
```

**Seeding Commands:**
```bash
# All seeders
make seed

# Specific seeders
./bin/seed --world --auth --site

# Verification
./bin/seed --verify
```

**Seeding Order:**
1. **World Data** (regions, countries, states, cities)
2. **Auth Data** (roles, permissions, admin user)
3. **Site Data** (sites, device groups)

---

## 10. 🚀 System Startup Workflow

### 10.1 System Startup Sequence
```
[Config Load] → [Database Connect] → [Migration Check] → [Service Init] → [WebSocket Start] → [API Server Start]
```

**Startup Sequence:**

1. **Configuration Loading:**
   ```go
   func loadConfig() error {
       config.LoadConfig("config.yaml")
       
       // Environment variable overrides
       if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
           config.AppConfig.Database.Host = dbHost
       }
       
       return validateConfig()
   }
   ```

2. **Database Connection:**
   ```go
   func initDatabase() error {
       err := database.InitDatabase()
       if err != nil {
           return fmt.Errorf("database connection failed: %v", err)
       }
       
       // Health check
       return database.DB.Exec("SELECT 1").Error
   }
   ```

3. **Service Initialization:**
   ```go
   func initServices() error {
       // Redis connection
       if err := cache.InitRedis(); err != nil {
           return err
       }
       
       // RabbitMQ connection
       if err := queue.InitRabbitMQ(); err != nil {
           return err
       }
       
       // Casbin enforcer
       if err := auth.InitCasbin(); err != nil {
           return err
       }
       
       return nil
   }
   ```

4. **Server Startup:**
   ```go
   func startServers() {
       // WebSocket server (port 8081)
       go func() {
           log.Println("Starting WebSocket server on :8081")
           websocket.StartWebSocketServer()
       }()
       
       // API server (port 8080)
       log.Println("Starting API server on :8080")
       api.StartAPIServer()
   }
   ```

### 10.2 Graceful Shutdown
```
[Signal Catch] → [New Connections Stop] → [Active Connections Drain] → [Services Cleanup] → [Database Close]
```

**Shutdown Sequence:**
```go
func gracefulShutdown() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    <-c
    log.Println("Graceful shutdown initiated...")
    
    // Stop accepting new connections
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Shutdown API server
    if err := apiServer.Shutdown(ctx); err != nil {
        log.Printf("API server shutdown error: %v", err)
    }
    
    // Close WebSocket connections
    websocket.CloseAllConnections()
    
    // Close database connections
    database.CloseDB()
    
    // Close Redis connection
    cache.CloseRedis()
    
    // Close RabbitMQ connection
    queue.CloseRabbitMQ()
    
    log.Println("Shutdown completed")
}
```

---

## 📊 Workflow Metrics and Monitoring

### Performance KPIs:
- **SMS Sending Success Rate:** >99%
- **API Response Time:** <200ms
- **WebSocket Latency:** <50ms
- **Device Availability:** >95%
- **Database Query Time:** <100ms

### Log Formats:
```json
{
  "timestamp": "2024-01-01T10:00:00Z",
  "level": "INFO",
  "workflow": "sms_send",
  "step": "device_selection",
  "device_id": "device123",
  "target": "+905551234567",
  "duration_ms": 45,
  "success": true
}
```

### Error Handling Patterns:
- **Immediate Retry:** Network timeouts
- **Exponential Backoff:** Device unavailability  
- **Circuit Breaker:** Service failures
- **Dead Letter Queue:** Unprocessable messages

This documentation covers all critical workflows of TsimServer and enables system administrators to understand operational processes. 