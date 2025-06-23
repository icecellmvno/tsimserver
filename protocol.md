# TsimCloud Android Client WebSocket Protokolü

Bu doküman, TsimCloud Android istemcisi ile sunucu arasındaki WebSocket iletişim protokolünü tanımlar.

## 1. Bağlantı Bilgileri
- **Protokol**: WebSocket (`ws://` veya `wss://`)
- **Varsayılan Port**: `8080`
- **Endpoint**: `/ws`
- **URL Formatı**: `ws://sunucu_adresi:port/ws`

## 2. Kimlik Doğrulama ve Oturum
### 2.1. Client -> Server: Kimlik Doğrulama İsteği
İstemci, sunucuya bağlandıktan sonra kimliğini doğrulamak için bu mesajı gönderir.

```json
{
    "type": "auth",
    "connectkey": "CIHAZIN_BAGLANTI_ANAHTARI"
}
```

### 2.2. Server -> Client: Kimlik Doğrulama Yanıtı
Sunucu, kimlik doğrulama isteğinin sonucunu bu mesajla bildirir. Başarılı olursa, istemcinin arayüzde göstereceği ek bilgiler de gönderilir.

```json
{
    "type": "auth_response",
    "success": true|false,
    "sitename": "string",
    "groupname": "string",
    "devicename": "string"
}
```

## 3. Cihaz Yönetimi
### 3.1. Client -> Server: Cihaz Kaydı
İstemci, başarılı kimlik doğrulamanın ardından sunucuya cihaz bilgilerini kaydetmek için bu mesajı gönderir.

```json
{
    "type": "device_registration",
    "payload": {
        "device_id": "string",
        "device_name": "string",
        "model": "string",
        "android_version": "string",
        "app_version": "string",
        "batteryLevel": number,
        "batteryStatus": "string",
        "latitude": number,
        "longitude": number,
        "timestamp": number,
        "simCards": [
            {
                "identifier": "string",
                "imsi": "string",
                "imei": "string",
                "operator": "string",
                "phoneNumber": "string",
                "signalStrength": number,
                "networkType": "string",
                "mcc": "string",
                "mnc": "string",
                "isActive": boolean
            }
        ]
    }
}
```

### 3.2. Client -> Server: Cihaz Durum Güncellemesi
İstemci, periyodik olarak (örn. 30 saniyede bir) veya önemli bir durum değişikliği olduğunda (örn. SIM kart değişikliği) cihazın anlık durumunu sunucuya bu mesajla bildirir. Payload yapısı `device_registration` ile aynıdır.

```json
{
    "type": "device_status",
    "payload": {
        "device_id": "string",
        "batteryLevel": number,
        "batteryStatus": "string",
        "latitude": number,
        "longitude": number,
        "timestamp": number,
        "simCards": [
            {
                "identifier": "string",
                "imsi": "string",
                "imei": "string",
                "operator": "string",
                "phoneNumber": "string",
                "signalStrength": number,
                "networkType": "string",
                "mcc": "string",
                "mnc": "string",
                "isActive": boolean
            }
        ]
    }
}
```

## 4. SMS Yönetimi
### 4.1. Server -> Client: SMS Gönderme Komutu
Sunucu, istemciye SMS göndermesi için bu komutu gönderir.

```json
{
    "type": "send_sms",
    "target": "ALICI_NUMARASI",
    "simSlot": 0,
    "message": "SMS_METNI",
    "internalLogId": 12345
}
```
- `simSlot`: `0` (birinci SIM), `1` (ikinci SIM)

### 4.2. Client -> Server: Gelen SMS Bildirimi
İstemci, cihaza yeni bir SMS geldiğinde bu mesajla sunucuyu bilgilendirir.

```json
{
    "type": "incoming_sms",
    "from": "GONDEREN_NUMARA",
    "message": "GELEN_SMS_METNI",
    "timestamp": number
}
```

### 4.3. Client -> Server: SMS Teslimat Raporu (DLR)
İstemci, sunucudan gelen bir SMS gönderme komutunun nihai durumunu bu mesajla bildirir.

```json
{
    "type": "sms_delivery_report",
    "id": 12345,
    "simSlot": 0,
    "sub": 1,
    "dlvrd": 1,
    "submit_date": "YYMMDDhhmm",
    "done_date": "YYMMDDhhmm",
    "stat": "DELIVRD",
    "err": "000",
    "text": "ORIJINAL_SMS_METNI"
}
```
- **id**: `send_sms` komutundaki `internalLogId` ile aynıdır.
- **stat**: Mesaj durumu (`ENROUTE`, `DELIVRD`, `EXPIRED`, `DELETED`, `UNDELIV`, `ACCEPTD`, `UNKNOWN`, `REJECTD`).
- **err**: SMPP standartlarına göre hata kodu.

## 5. USSD Yönetimi
### 5.1. Server -> Client: USSD Komutu Gönderme
Sunucu, istemciye bir USSD kodu çalıştırması için bu komutu gönderir.

```json
{
    "type": "ussd_command",
    "ussdCode": "*123#",
    "simSlot": 0,
    "internalLogId": 54321
}
```

### 5.2. Client -> Server: USSD Komut Sonucu
İstemci, çalıştırılan USSD komutunun sonucunu sunucuya bu mesajla bildirir.

```json
{
    "type": "ussd_result",
    "internalLogId": 54321,
    "success": true,
    "result": "USSD_YANIT_METNI",
    "errorMessage": "",
    "timestamp": number
}
```

## 6. Cihaz ve SIM Kontrolü
### 6.1. Server -> Client: Cihazı/SIM'i Devre Dışı Bırakma/Etkinleştirme
Sunucu, istemci cihazını veya içindeki bir SIM kartı uzaktan devre dışı bırakmak veya tekrar etkinleştirmek için bu komutları kullanır.

```json
{
    "type": "disable_device" | "enable_device",
    "deviceId": "string"
}
```
```json
{
    "type": "disable_sim" | "enable_sim",
    "deviceId": "string",
    "simSlot": 0
}
```

## 7. Alarm ve Bildirimler
### 7.1. Client -> Server: İstemci Kaynaklı Alarm
İstemci, kritik bir durum algıladığında (düşük pil, SIM kartın bloke olması vb.) sunucuya alarm gönderir.

```json
{
    "type": "alarm",
    "alarmType": "battery_low" | "sim_blocked" | "connection_lost" | "sim_card_removed" | "sim_card_inserted",
    "message": "Alarm açıklaması",
    "timestamp": number
}
```

### 7.2. Server -> Client: Sunucu Kaynaklı Alarm
Sunucu, istemcide sesli veya görsel bir alarm tetiklemek için bu mesajı gönderir.

```json
{
    "type": "alarm",
    "title": "Sunucu Alarmı",
    "message": "Bu bir test alarmıdır."
}
```
## 8. Hata Kodları ve Yönetimi

- **`PERMISSION_DENIED`**: İstemcinin istenen işlemi yapmak için gerekli Android iznine sahip olmaması.
- **`CONNECTION_FAILED`**: WebSocket bağlantısının kurulamaması.
- **`AUTHENTICATION_FAILED`**: Kimlik doğrulamanın başarısız olması.

## 9. Bakiye Sorgulama
### Server -> Client
```json
{
    "type": "check_balance",
    "simSlot": number,
    "ussdCode": "string",
    "internalLogId": number
}
```

## 10. Telefon Numarası Keşfi
### Server -> Client
```json
{
    "type": "discover_phone_number",
    "simSlot": number,
    "ussdCode": "string",
    "internalLogId": number
}
```
### Client -> Server (Yanıt)
```json
{
    "type": "phone_number_result",
    "internalLogId": number,
    "success": boolean,
    "phoneNumber": "string",
    "errorMessage": "string",
    "timestamp": number
}
```

## 11. Performans Optimizasyonları
- USSD monitoring arka planda çalışır
- Logcat filtreleme ile CPU kullanımı optimize edilir
- Gereksiz mesajlar filtrelenir
- Bellek kullanımı minimize edilir

## 12. Güvenlik Önlemleri
- Root erişimi sadece gerekli işlemler için kullanılır
- Sistem dosyaları değiştirilmez
- Sadece USSD mesajları yakalanır
- Kişisel veriler korunur 
