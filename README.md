# TsimServer - TsimCloud SIM Server

TsimCloud protokolüne uygun, Go dilinde yazılmış multithread SIM sunucusu. WebSocket üzerinden Android cihazlarla iletişim kurar ve SMS, USSD yönetimi yapar.

## Özellikler

- **WebSocket İletişimi**: Android cihazlarla gerçek zamanlı iletişim
- **SMS Yönetimi**: SMS gönderme, alma ve teslimat raporları
- **USSD Komutları**: USSD komutları gönderme ve sonuçları alma
- **Cihaz Yönetimi**: Cihaz kaydı, durum takibi ve uzaktan kontrol
- **SIM Kart Yönetimi**: SIM kart bilgileri ve aktivasyon durumu
- **Alarm Sistemi**: Cihaz alarmları ve sunucu alarmları
- **İstatistikler**: Kapsamlı raporlama ve dashboard
- **Multithread**: Eşzamanlı çoklu bağlantı desteği
- **RESTful API**: Kapsamlı web API

## Teknolojiler

- **Go Fiber**: Web framework
- **GORM**: ORM library
- **PostgreSQL**: Veritabanı
- **Redis**: Cache ve session yönetimi
- **RabbitMQ**: Mesaj kuyruğu
- **WebSocket**: Gerçek zamanlı iletişim
- **Viper**: Konfigürasyon yönetimi

## Kurulum

### Gereksinimler

- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- RabbitMQ 3.8+

### Adımlar

1. **Depoyu klonlayın**
```bash
git clone <repo-url>
cd tsimserver
```

2. **Bağımlılıkları yükleyin**
```bash
go mod download
```

3. **Konfigürasyonu düzenleyin**
```bash
cp config.yaml.example config.yaml
nano config.yaml
```

4. **Veritabanını hazırlayın**
```sql
CREATE DATABASE tsimserver;
CREATE USER tsimserver WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE tsimserver TO tsimserver;
```

5. **Sunucuyu başlatın**
```bash
# Makefile kullanarak (önerilen)
make setup          # Geliştirme ortamı kurulumu
make run-server     # Ana sunucuyu başlat

# Manuel olarak
go run cmd/server/main.go
```

## Konfigürasyon

`config.yaml` dosyasındaki ayarları ihtiyaçlarınıza göre düzenleyin:

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
  endpoint: "/ws"
  read_buffer_size: 1024
  write_buffer_size: 1024

logging:
  level: "info"
```

## API Endpoints

### Cihaz Yönetimi
- `GET /api/v1/devices` - Tüm cihazları listele
- `POST /api/v1/devices` - Yeni cihaz oluştur
- `GET /api/v1/devices/:id` - Cihaz detayları
- `PUT /api/v1/devices/:id` - Cihaz güncelle
- `DELETE /api/v1/devices/:id` - Cihaz sil
- `POST /api/v1/devices/:id/disable` - Cihazı devre dışı bırak
- `POST /api/v1/devices/:id/enable` - Cihazı etkinleştir
- `POST /api/v1/devices/:id/alarm` - Cihaza alarm gönder

### SMS Yönetimi
- `POST /api/v1/sms/send` - SMS gönder
- `GET /api/v1/sms/incoming` - Gelen SMS'ler
- `GET /api/v1/sms/outgoing` - Giden SMS'ler
- `GET /api/v1/sms/stats` - SMS istatistikleri
- `GET /api/v1/sms/device/:deviceId` - Cihaza özel SMS'ler

### USSD Yönetimi
- `POST /api/v1/ussd/send` - USSD komutu gönder
- `GET /api/v1/ussd/device/:deviceId` - Cihaza özel USSD komutları

### Alarm Yönetimi
- `GET /api/v1/alarms` - Alarmları listele
- `GET /api/v1/alarms/:id` - Alarm detayları
- `POST /api/v1/alarms/:id/resolve` - Alarmı çözümle
- `DELETE /api/v1/alarms/:id` - Alarm sil

### İstatistikler
- `GET /api/v1/stats/dashboard` - Dashboard istatistikleri
- `GET /api/v1/stats/devices` - Cihaz istatistikleri

## WebSocket Protokolü

Sunucu `/ws` endpoint'inde WebSocket bağlantılarını kabul eder. Protokol detayları için `protocol.md` dosyasına bakın.

### Kimlik Doğrulama
```json
{
    "type": "auth",
    "connectkey": "CIHAZIN_BAGLANTI_ANAHTARI"
}
```

### Cihaz Kaydı
```json
{
    "type": "device_registration",
    "payload": {
        "device_id": "string",
        "device_name": "string",
        "model": "string",
        "android_version": "string",
        "app_version": "string",
        "batteryLevel": 85,
        "batteryStatus": "charging",
        "latitude": 41.0082,
        "longitude": 28.9784,
        "timestamp": 1640995200,
        "simCards": [...]
    }
}
```

## Geliştirme

### Proje Yapısı
```
tsimserver/
├── auth/           # Casbin yetkilendirme
├── cache/          # Redis cache yönetimi
├── cmd/            # Komut satırı uygulamaları
│   ├── server/     # Ana API sunucusu
│   ├── migrate/    # Veritabanı migration
│   ├── seed/       # Veri seeding
│   └── websocket/  # WebSocket sunucusu
├── config/         # Konfigürasyon
├── database/       # Veritabanı bağlantısı
├── handlers/       # HTTP ve WebSocket handler'lar
├── middleware/     # Auth middleware'ler
├── models/         # Veritabanı modelleri
├── queue/          # RabbitMQ mesaj kuyruğu
├── seeders/        # Veri seeding fonksiyonları
├── types/          # WebSocket mesaj tipleri
├── utils/          # JWT ve yardımcı fonksiyonlar
├── websocket/      # WebSocket yönetimi
├── Makefile        # Build ve çalıştırma komutları
└── config.yaml     # Konfigürasyon dosyası
```

### Makefile Komutları

Proje build ve yönetim işlemleri için Makefile kullanın:

```bash
# Build komutları
make build              # Tüm binary'leri build et
make build-server       # Sadece server build et
make build-migrate      # Sadece migrate build et
make build-seed         # Sadece seed build et
make build-websocket    # Sadece websocket build et

# Veritabanı komutları
make migrate            # Migration'ları çalıştır
make migrate-reset      # Veritabanını sıfırla
make migrate-rollback   # Migration'ları geri al
make seed               # Tüm verileri seed et
make seed-world         # Sadece world verilerini seed et
make seed-auth          # Sadece auth verilerini seed et

# Çalıştırma komutları
make run-server         # Ana sunucuyu çalıştır
make run-websocket      # WebSocket sunucusunu çalıştır

# Kurulum komutları
make setup              # Geliştirme ortamı kurulumu
make setup-prod         # Production kurulumu

# Yardımcı komutları
make help               # Tüm komutları listele
make clean              # Build dosyalarını temizle
make deps               # Bağımlılıkları güncelle
```

### Test Etme
```bash
# Sunucuyu geliştirme modunda çalıştır
make run-server

# Health check
curl http://localhost:8080/api/v1/health

# Authentication endpoint test
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# WebSocket bağlantısı test et
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" \
     --header "Sec-WebSocket-Version: 13" \
     http://localhost:8080/ws

# WebSocket sunucusunu ayrı çalıştır
make run-websocket    # Port 8081'de çalışır
```

### Kimlik Doğrulama ve Yetkilendirme

Sistem JWT tabanlı kimlik doğrulama ve Casbin RBAC yetkilendirme kullanır:

#### Varsayılan Kullanıcı
- **Kullanıcı adı**: admin
- **Şifre**: admin123
- **Rol**: administrator (tam erişim)

#### Roller ve İzinler
- **admin**: Tüm kaynaklara tam erişim
- **manager**: Cihaz ve SMS yönetimi
- **operator**: Sınırlı operasyon erişimi
- **viewer**: Sadece okuma erişimi

#### API Kullanımı
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# JWT token ile protected endpoint
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/v1/devices
```
     --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     --header "Sec-WebSocket-Version: 13" \
     http://localhost:8080/ws
```

## Deployment

### Docker ile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o tsimserver main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/tsimserver .
COPY --from=builder /app/config.yaml .
CMD ["./tsimserver"]
```

### Systemd ile
```ini
[Unit]
Description=TsimServer
After=network.target

[Service]
Type=simple
User=tsimserver
WorkingDirectory=/opt/tsimserver
ExecStart=/opt/tsimserver/tsimserver
Restart=always

[Install]
WantedBy=multi-user.target
```

## Lisans

Bu proje MIT lisansı altında lisanslanmıştır.

## Katkıda Bulunma

1. Fork yapın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit edin (`git commit -m 'Add amazing feature'`)
4. Branch'inizi push edin (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## Destek

Herhangi bir sorun için issue açabilir veya [email] üzerinden iletişime geçebilirsiniz. 