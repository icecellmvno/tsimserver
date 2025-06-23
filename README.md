# TsimServer - Android SMS Gateway System

A complete Android SMS Gateway system built with Go, featuring intelligent routing, site management, and real-time monitoring. Communicates with Android devices via WebSocket and manages SMS, USSD operations with advanced routing capabilities.

## Features

- **Smart SMS Routing**: Intelligent device selection based on country, operator, battery, and signal strength
- **Site Management**: Multi-location SMS hub management with device groups
- **Real-time Communication**: WebSocket communication with Android devices
- **SMS Management**: Send, receive SMS and delivery reports with retry logic and priority queuing
- **USSD Commands**: Send USSD commands and receive responses
- **Device Management**: Device registration, status monitoring, and remote control
- **SIM Card Management**: Multi-SIM support with operator-based routing
- **Admin Test System**: Special admin-only test SMS and commands
- **Alarm System**: Comprehensive monitoring with automated alerts
- **Statistics & Reporting**: Advanced analytics and dashboard
- **Authentication & Authorization**: JWT-based auth with RBAC using Casbin
- **Scalable Architecture**: Microservices with separate binaries for different functions
- **RESTful API**: Comprehensive web API with full CRUD operations

## Technologies

- **Go Fiber**: High-performance web framework
- **GORM**: Feature-rich ORM library
- **PostgreSQL**: Primary database with advanced features
- **Redis**: Caching and session management
- **RabbitMQ**: Message queue for delivery reports
- **WebSocket**: Real-time bidirectional communication
- **Casbin**: RBAC authorization
- **JWT**: Secure authentication tokens
- **Docker**: Containerization support

## Architecture

### Modular CMD Structure
- **cmd/server**: Main API server (port 8080)
- **cmd/websocket**: Dedicated WebSocket server (port 8081)
- **cmd/migrate**: Database migration tool
- **cmd/seed**: Data seeding utility

### Smart SMS Routing System
```
SMS Request → Country Detection → Operator Matching → Device Selection → WebSocket Delivery
     ↓               ↓                    ↓                   ↓              ↓
Phone Number → Extract Code → Find Groups → Best Device → Send to Android
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- RabbitMQ 3.8+ (optional)
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/icecellmvno/tsimserver
cd tsimserver
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure the system**
```bash
cp config.yaml.example config.yaml
nano config.yaml
```

4. **Setup database**
```sql
CREATE DATABASE tsimserver;
CREATE USER tsimserver WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE tsimserver TO tsimserver;
```

5. **Start the system**
```bash
# Using Makefile (recommended)
make setup          # Development environment setup
make migrate        # Run database migrations
make seed           # Seed initial data
make run-server     # Start main server

# Or manually
go run cmd/server/main.go
```

### Docker Setup
```bash
# Start all services
docker-compose up -d

# Run migrations and seeding
make migrate
make seed

# Start the application
make run-server
```

## Configuration

Edit the `config.yaml` file according to your needs:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "tsimserver"
  sslmode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"

jwt:
  secret: "your_jwt_secret_key_here"
  expire_hours: 24

websocket:
  port: 8081
  endpoint: "/ws"
  read_buffer_size: 1024
  write_buffer_size: 1024

logging:
  level: "info"
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - User logout

### Site Management
- `GET /api/v1/sites` - List all sites
- `POST /api/v1/sites` - Create new site
- `GET /api/v1/sites/:id` - Get site details
- `PUT /api/v1/sites/:id` - Update site
- `DELETE /api/v1/sites/:id` - Delete site
- `GET /api/v1/sites/:id/stats` - Site statistics

### Device Groups
- `GET /api/v1/device-groups` - List device groups
- `POST /api/v1/device-groups` - Create device group
- `GET /api/v1/device-groups/:id` - Get group details
- `PUT /api/v1/device-groups/:id` - Update group
- `DELETE /api/v1/device-groups/:id` - Delete group

### Device Management
- `GET /api/v1/devices` - List all devices
- `POST /api/v1/devices` - Create new device
- `GET /api/v1/devices/:id` - Device details
- `PUT /api/v1/devices/:id` - Update device
- `DELETE /api/v1/devices/:id` - Delete device
- `POST /api/v1/devices/:id/disable` - Disable device
- `POST /api/v1/devices/:id/enable` - Enable device

### Smart SMS Gateway
- `POST /api/v1/sms-gateway/send` - Send SMS with intelligent routing
- `POST /api/v1/sms-gateway/test` - Admin test SMS
- `POST /api/v1/sms-gateway/test-command` - Send test commands to devices
- `POST /api/v1/sms-gateway/delivery-report` - Process delivery reports

### SMS Management
- `GET /api/v1/sms/incoming` - List incoming SMS
- `GET /api/v1/sms/outgoing` - List outgoing SMS
- `GET /api/v1/sms/stats` - SMS statistics
- `GET /api/v1/sms/device/:deviceId` - Device-specific SMS

### USSD Management
- `POST /api/v1/ussd/send` - Send USSD command
- `GET /api/v1/ussd/device/:deviceId` - Device USSD commands

### User Management
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user details
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Alarm Management
- `GET /api/v1/alarms` - List alarms
- `GET /api/v1/alarms/:id` - Alarm details
- `POST /api/v1/alarms/:id/resolve` - Resolve alarm
- `DELETE /api/v1/alarms/:id` - Delete alarm

### Statistics
- `GET /api/v1/stats/dashboard` - Dashboard statistics
- `GET /api/v1/stats/devices` - Device statistics

## WebSocket Protocol

The server accepts WebSocket connections at `/ws` endpoint. See `protocol.md` for detailed protocol specifications.

### Authentication
```json
{
    "action": "auth",
    "data": {
        "connect_key": "DEVICE_CONNECTION_KEY"
    }
}
```

### Device Registration
```json
{
    "action": "device_registration",
    "data": {
        "device_id": "string",
        "device_name": "string",
        "model": "string",
        "android_version": "string",
        "app_version": "string",
        "battery_level": 85,
        "battery_status": "charging",
        "latitude": 41.0082,
        "longitude": 28.9784,
        "timestamp": 1640995200,
        "sim_cards": [...]
    }
}
```

### SMS Sending
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

## Development

### Project Structure
```
tsimserver/
├── auth/               # Casbin authorization
├── cache/              # Redis cache management
├── cmd/                # Command line applications
│   ├── server/         # Main API server
│   ├── migrate/        # Database migration tool
│   ├── seed/           # Data seeding utility
│   └── websocket/      # WebSocket server
├── config/             # Configuration management
├── database/           # Database connection and models
├── handlers/           # HTTP and WebSocket handlers
├── middleware/         # Authentication middleware
├── models/             # Database models (GORM)
├── queue/              # RabbitMQ message queue
├── seeders/            # Data seeding functions
├── types/              # WebSocket message types
├── utils/              # JWT and utility functions
├── websocket/          # WebSocket connection management
├── Makefile            # Build and run commands
├── DATABASE_SCHEMA.md  # Comprehensive database documentation
├── WORKFLOW_DOCUMENTATION.md # System workflows
└── config.yaml         # Configuration file
```

### Makefile Commands

Use Makefile for project build and management:

```bash
# Build commands
make build              # Build all binaries
make build-server       # Build server only
make build-migrate      # Build migrate only
make build-seed         # Build seed only
make build-websocket    # Build websocket only

# Database commands
make migrate            # Run migrations
make migrate-reset      # Reset database
make migrate-rollback   # Rollback migrations
make seed               # Seed all data
make seed-world         # Seed world data only
make seed-auth          # Seed auth data only
make seed-site          # Seed site data only

# Run commands
make run-server         # Run main server
make run-websocket      # Run WebSocket server

# Setup commands
make setup              # Development environment setup
make setup-prod         # Production setup

# Utility commands
make help               # List all commands
make clean              # Clean build files
make deps               # Update dependencies
```

### Testing

```bash
# Start server in development mode
make run-server

# Health check
curl http://localhost:8080/api/v1/health

# Authentication test
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Test SMS sending with intelligent routing
curl -X POST http://localhost:8080/api/v1/sms-gateway/send \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"target":"+905551234567","message":"Test message"}'

# WebSocket connection test
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" \
     --header "Sec-WebSocket-Version: 13" \
     http://localhost:8081/ws

# Run separate WebSocket server
make run-websocket    # Runs on port 8081
```

### Authentication & Authorization

The system uses JWT-based authentication and Casbin RBAC authorization:

#### Default User
- **Username**: admin
- **Password**: admin123
- **Role**: administrator (full access)

#### Roles and Permissions
- **admin**: Full access to all resources
- **manager**: Device and SMS management
- **operator**: Limited operation access
- **viewer**: Read-only access

#### API Usage
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Use JWT token for protected endpoints
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/v1/devices
```

### Smart SMS Routing

The system implements intelligent SMS routing based on:

1. **Country Detection**: Extract country code from phone number
2. **Site Selection**: Find appropriate site for the country
3. **Device Group Matching**: Match operator (Turkcell, Vodafone, etc.)
4. **Device Selection**: Choose best available device based on:
   - Battery level (≥10%)
   - Signal strength
   - Last seen time
   - Availability status

```go
// Example: Send SMS to Turkey (+90) number
// System automatically:
// 1. Detects country: TR
// 2. Finds Turkey sites
// 3. Matches Turkcell operator from number
// 4. Selects best Turkcell device
// 5. Sends SMS via WebSocket
```

## Deployment

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o tsimserver cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/tsimserver .
COPY --from=builder /app/config.yaml .
CMD ["./tsimserver"]
```

### Production Deployment with Docker Compose
```yaml
version: '3.8'
services:
  tsimserver:
    build: .
    ports:
      - "8080:8080"
      - "8081:8081"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
      
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: tsimserver
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      
  redis:
    image: redis:6-alpine
    
  rabbitmq:
    image: rabbitmq:3.8-management
```

### Systemd Service
```ini
[Unit]
Description=TsimServer Android SMS Gateway
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=tsimserver
WorkingDirectory=/opt/tsimserver
ExecStart=/opt/tsimserver/bin/tsimserver
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Database Schema

The system uses a comprehensive database schema with 20+ tables:

- **Geographic Data**: regions, countries, states, cities (world data)
- **Site Management**: sites, device_groups
- **Device Management**: devices, sim_cards, device_statuses
- **SMS & Messaging**: sms_messages, ussd_commands
- **User & Auth**: users, roles, permissions, sessions
- **Monitoring**: alarms

See `DATABASE_SCHEMA.md` for detailed documentation.

## Workflows

The system implements complex workflows for:

- Smart SMS routing and delivery
- Device management and monitoring
- Site and device group management
- Authentication and authorization
- Admin testing capabilities
- Real-time alarm management

See `WORKFLOW_DOCUMENTATION.md` for detailed workflow documentation.

## Performance & Monitoring

### Key Performance Indicators
- **SMS Sending Success Rate**: >99%
- **API Response Time**: <200ms
- **WebSocket Latency**: <50ms
- **Device Availability**: >95%
- **Database Query Time**: <100ms

### Monitoring Features
- Real-time device status monitoring
- Battery level and signal strength tracking
- Automatic alarm generation
- Performance metrics and analytics
- Delivery report tracking

## License

This project is licensed under the MIT License.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

For any issues or questions:
- Open an issue on GitHub
- Check the documentation in `DATABASE_SCHEMA.md` and `WORKFLOW_DOCUMENTATION.md`
- Review the API endpoints and examples above

## Roadmap

- [ ] Web dashboard UI
- [ ] SMS campaign management
- [ ] Advanced analytics and reporting
- [ ] Multi-language support
- [ ] Mobile app for monitoring
- [ ] API rate limiting enhancements
- [ ] Kubernetes deployment manifests 