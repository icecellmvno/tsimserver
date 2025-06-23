# World Data API Modülleri

Bu dokümantasyon, TsimServer projesi içerisinde oluşturulan World veritabanı (Regions, Countries, States, Cities) modüllerini açıklamaktadır.

## Modeller

### Region (Bölge)
```go
type Region struct {
    ID           int64     `json:"id"`
    Name         string    `json:"name"`
    Translations string    `json:"translations"`
    CreatedAt    *time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    Flag         int16     `json:"flag"`
    WikiDataID   *string   `json:"wikiDataId"`
    
    // İlişkiler
    Subregions []Subregion `json:"subregions"`
    Countries  []Country   `json:"countries"`
}
```

### Subregion (Alt Bölge)
```go
type Subregion struct {
    ID           int64     `json:"id"`
    Name         string    `json:"name"`
    Translations string    `json:"translations"`
    RegionID     int64     `json:"region_id"`
    CreatedAt    *time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    Flag         int16     `json:"flag"`
    WikiDataID   *string   `json:"wikiDataId"`
    
    // İlişkiler
    Region    *Region   `json:"region"`
    Countries []Country `json:"countries"`
}
```

### Country (Ülke)
```go
type Country struct {
    ID               int64      `json:"id"`
    Name             string     `json:"name"`
    ISO3             *string    `json:"iso3"`
    NumericCode      *string    `json:"numeric_code"`
    ISO2             *string    `json:"iso2"`
    PhoneCode        *string    `json:"phonecode"`
    Capital          *string    `json:"capital"`
    Currency         *string    `json:"currency"`
    CurrencyName     *string    `json:"currency_name"`
    CurrencySymbol   *string    `json:"currency_symbol"`
    // ... diğer alanlar
    
    // İlişkiler
    RegionModel    *Region     `json:"region_model"`
    SubregionModel *Subregion  `json:"subregion_model"`
    States         []State     `json:"states"`
    Cities         []City      `json:"cities"`
}
```

### State (Eyalet/İl)
```go
type State struct {
    ID          int64      `json:"id"`
    Name        string     `json:"name"`
    CountryID   int64      `json:"country_id"`
    CountryCode string     `json:"country_code"`
    FipsCode    *string    `json:"fips_code"`
    ISO2        *string    `json:"iso2"`
    Type        *string    `json:"type"`
    Level       *int       `json:"level"`
    ParentID    *int       `json:"parent_id"`
    Native      *string    `json:"native"`
    Latitude    *float64   `json:"latitude"`
    Longitude   *float64   `json:"longitude"`
    CreatedAt   *time.Time `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    Flag        int16      `json:"flag"`
    WikiDataID  *string    `json:"wikiDataId"`
    
    // İlişkiler
    Country *Country `json:"country"`
    Cities  []City   `json:"cities"`
}
```

### City (Şehir)
```go
type City struct {
    ID          int64      `json:"id"`
    Name        string     `json:"name"`
    StateID     int64      `json:"state_id"`
    StateCode   string     `json:"state_code"`
    CountryID   int64      `json:"country_id"`
    CountryCode string     `json:"country_code"`
    Latitude    float64    `json:"latitude"`
    Longitude   float64    `json:"longitude"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    Flag        int16      `json:"flag"`
    WikiDataID  *string    `json:"wikiDataId"`
    
    // İlişkiler
    State   *State   `json:"state"`
    Country *Country `json:"country"`
}
```

## API Endpoints

### Regions (Bölgeler)
- `GET /api/v1/regions` - Tüm bölgeleri listele (sayfalama ile)
- `GET /api/v1/regions/:id` - Belirli bir bölgeyi getir
- `POST /api/v1/regions` - Yeni bölge oluştur
- `PUT /api/v1/regions/:id` - Bölgeyi güncelle
- `DELETE /api/v1/regions/:id` - Bölgeyi sil
- `GET /api/v1/regions/:id/subregions` - Bölgenin alt bölgelerini getir
- `GET /api/v1/regions/:id/countries` - Bölgenin ülkelerini getir

### Subregions (Alt Bölgeler)
- `GET /api/v1/subregions` - Tüm alt bölgeleri listele
- `GET /api/v1/subregions/:id` - Belirli bir alt bölgeyi getir
- `POST /api/v1/subregions` - Yeni alt bölge oluştur
- `PUT /api/v1/subregions/:id` - Alt bölgeyi güncelle
- `DELETE /api/v1/subregions/:id` - Alt bölgeyi sil
- `GET /api/v1/subregions/:id/countries` - Alt bölgenin ülkelerini getir

### Countries (Ülkeler)
- `GET /api/v1/countries` - Tüm ülkeleri listele
- `GET /api/v1/countries/:id` - Belirli bir ülkeyi getir
- `GET /api/v1/countries/iso/:iso` - ISO2 veya ISO3 koduna göre ülke getir
- `POST /api/v1/countries` - Yeni ülke oluştur
- `PUT /api/v1/countries/:id` - Ülkeyi güncelle
- `DELETE /api/v1/countries/:id` - Ülkeyi sil
- `GET /api/v1/countries/:id/states` - Ülkenin eyaletlerini getir
- `GET /api/v1/countries/:id/cities` - Ülkenin şehirlerini getir

### States (Eyaletler)
- `GET /api/v1/states` - Tüm eyaletleri listele
- `GET /api/v1/states/:id` - Belirli bir eyaleti getir
- `POST /api/v1/states` - Yeni eyalet oluştur
- `PUT /api/v1/states/:id` - Eyaleti güncelle
- `DELETE /api/v1/states/:id` - Eyaleti sil
- `GET /api/v1/states/:id/cities` - Eyaletin şehirlerini getir

### Cities (Şehirler)
- `GET /api/v1/cities` - Tüm şehirleri listele
- `GET /api/v1/cities/:id` - Belirli bir şehri getir
- `GET /api/v1/cities/search/coordinates` - Koordinatlara göre yakın şehirleri ara
- `POST /api/v1/cities` - Yeni şehir oluştur
- `PUT /api/v1/cities/:id` - Şehri güncelle
- `DELETE /api/v1/cities/:id` - Şehri sil
- `GET /api/v1/cities/stats` - Şehir istatistikleri

## Query Parameters

### Sayfalama
- `page` - Sayfa numarası (varsayılan: 1)
- `limit` - Sayfa başına kayıt sayısı (varsayılan: 50)

### Arama ve Filtreleme
- `search` - İsim bazlı arama
- `region_id` - Bölge ID'sine göre filtreleme
- `subregion_id` - Alt bölge ID'sine göre filtreleme
- `country_id` - Ülke ID'sine göre filtreleme
- `state_id` - Eyalet ID'sine göre filtreleme

### Koordinat Arama (Cities)
- `lat` - Enlem
- `lon` - Boylam
- `radius` - Arama yarıçapı (km, varsayılan: 10)

## Veri Seeding

### Manuel Seeding
World veritabanını seed etmek için:

```bash
go run cmd/seed-world/main.go
```

### Otomatik Seeding
Sunucu başlatıldığında otomatik olarak world verileri kontrol edilir ve gerekirse yüklenir.

## Örnek Kullanım

### Türkiye'deki Şehirleri Getirme
```bash
curl "http://localhost:8080/api/v1/countries/iso/TR" | jq .id
curl "http://localhost:8080/api/v1/countries/223/cities?page=1&limit=100"
```

### İstanbul Koordinatlarına Yakın Şehirler
```bash
curl "http://localhost:8080/api/v1/cities/search/coordinates?lat=41.0082&lon=28.9784&radius=50"
```

### Avrupa Bölgesindeki Ülkeler
```bash
curl "http://localhost:8080/api/v1/regions/3/countries"
```

## Özellikler

- **Sayfalama**: Tüm listeleme endpoint'lerinde sayfalama desteği
- **Arama**: İsim bazlı arama desteği
- **Filtreleme**: Hierarchical filtreleme (bölge > alt bölge > ülke > eyalet > şehir)
- **Koordinat Arama**: Haversine formülü kullanarak konum bazlı arama
- **İlişkisel Veri**: GORM ile tanımlanmış ilişkiler
- **İstatistikler**: Şehir sayıları ve dağılım istatistikleri
- **Güvenli Silme**: Bağımlı kayıtlar varsa silme işlemini engelleme
- **Doğrulama**: Gerekli alanların kontrolü

## Performans

- Veritabanı indeksleri kullanılarak hızlı sorgulama
- Preload ile ilişkisel verilerin optimize edilmiş yüklenmesi
- Coordinate-based search için Haversine formülü optimizasyonu
- Sayfalama ile büyük veri setlerinin verimli işlenmesi

## Veri Kaynağı

World veritabanı PostgreSQL dump formatında `dbsource/world.sql` dosyasında bulunmaktadır. Bu veri seti şunları içerir:

- 7 Bölge
- 22 Alt Bölge
- 250 Ülke
- 5000+ Eyalet/İl
- 150,000+ Şehir

Tüm veriler coordinate bilgileri, ISO kodları, para birimleri ve diğer metadata ile birlikte gelir. 