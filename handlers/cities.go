package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetCities returns all cities with pagination
func GetCities(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	countryID := c.Query("country_id", "")
	stateID := c.Query("state_id", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.City{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Filter by country
	if countryID != "" {
		query = query.Where("country_id = ?", countryID)
	}

	// Filter by state
	if stateID != "" {
		query = query.Where("state_id = ?", stateID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get cities with pagination
	var cities []models.City
	result := query.Offset(offset).Limit(limit).
		Preload("State").
		Preload("Country").
		Find(&cities)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch cities",
		})
	}

	return c.JSON(fiber.Map{
		"cities": cities,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetCity returns a specific city by ID
func GetCity(c *fiber.Ctx) error {
	cityIDStr := c.Params("id")
	cityID, err := strconv.ParseInt(cityIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid city ID",
		})
	}

	var city models.City
	result := database.DB.Preload("State").
		Preload("Country").
		Where("id = ?", cityID).
		First(&city)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "City not found",
		})
	}

	return c.JSON(city)
}

// SearchCitiesByCoordinates finds cities near given coordinates
func SearchCitiesByCoordinates(c *fiber.Ctx) error {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusStr := c.Query("radius", "10") // Default 10km radius
	limit := c.QueryInt("limit", 20)

	if latStr == "" || lonStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Latitude and longitude are required",
		})
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid latitude",
		})
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid longitude",
		})
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid radius",
		})
	}

	var cities []models.City

	// Use Haversine formula to find nearby cities
	query := `
		SELECT *, 
		(6371 * acos(cos(radians(?)) * cos(radians(latitude)) * 
		cos(radians(longitude) - radians(?)) + sin(radians(?)) * 
		sin(radians(latitude)))) AS distance 
		FROM cities 
		WHERE (6371 * acos(cos(radians(?)) * cos(radians(latitude)) * 
		cos(radians(longitude) - radians(?)) + sin(radians(?)) * 
		sin(radians(latitude)))) <= ? 
		ORDER BY distance 
		LIMIT ?`

	result := database.DB.Raw(query, lat, lon, lat, lat, lon, lat, radius, limit).
		Preload("State").
		Preload("Country").
		Find(&cities)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to search cities",
		})
	}

	return c.JSON(fiber.Map{
		"cities":    cities,
		"count":     len(cities),
		"latitude":  lat,
		"longitude": lon,
		"radius":    radius,
	})
}

// CreateCity creates a new city
func CreateCity(c *fiber.Ctx) error {
	var city models.City

	if err := c.BodyParser(&city); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if city.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "City name is required",
		})
	}
	if city.StateID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "State ID is required",
		})
	}
	if city.CountryID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Country ID is required",
		})
	}
	if city.StateCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "State code is required",
		})
	}
	if city.CountryCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Country code is required",
		})
	}

	// Create city
	if err := database.DB.Create(&city).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create city",
		})
	}

	return c.Status(201).JSON(city)
}

// UpdateCity updates an existing city
func UpdateCity(c *fiber.Ctx) error {
	cityIDStr := c.Params("id")
	cityID, err := strconv.ParseInt(cityIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid city ID",
		})
	}

	var city models.City
	if err := database.DB.Where("id = ?", cityID).First(&city).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "City not found",
		})
	}

	var updateData models.City
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if updateData.Name != "" {
		city.Name = updateData.Name
	}
	if updateData.StateID != 0 {
		city.StateID = updateData.StateID
	}
	if updateData.CountryID != 0 {
		city.CountryID = updateData.CountryID
	}
	if updateData.StateCode != "" {
		city.StateCode = updateData.StateCode
	}
	if updateData.CountryCode != "" {
		city.CountryCode = updateData.CountryCode
	}
	if updateData.Latitude != 0 {
		city.Latitude = updateData.Latitude
	}
	if updateData.Longitude != 0 {
		city.Longitude = updateData.Longitude
	}
	if updateData.WikiDataID != nil {
		city.WikiDataID = updateData.WikiDataID
	}
	city.Flag = updateData.Flag

	if err := database.DB.Save(&city).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update city",
		})
	}

	return c.JSON(city)
}

// DeleteCity deletes a city
func DeleteCity(c *fiber.Ctx) error {
	cityIDStr := c.Params("id")
	cityID, err := strconv.ParseInt(cityIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid city ID",
		})
	}

	if err := database.DB.Delete(&models.City{}, cityID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete city",
		})
	}

	return c.JSON(fiber.Map{
		"message": "City deleted successfully",
	})
}

// GetCitiesStats returns statistics about cities
func GetCitiesStats(c *fiber.Ctx) error {
	var stats struct {
		TotalCities      int64            `json:"total_cities"`
		TotalCountries   int64            `json:"total_countries"`
		TotalStates      int64            `json:"total_states"`
		CitiesPerCountry map[string]int64 `json:"cities_per_country"`
		CitiesPerState   map[string]int64 `json:"cities_per_state"`
	}

	// Get total counts
	database.DB.Model(&models.City{}).Count(&stats.TotalCities)
	database.DB.Model(&models.Country{}).Count(&stats.TotalCountries)
	database.DB.Model(&models.State{}).Count(&stats.TotalStates)

	// Get cities per country (top 10)
	type CountryStats struct {
		CountryName string `json:"country_name"`
		CityCount   int64  `json:"city_count"`
	}

	var countryStats []CountryStats
	database.DB.Table("cities").
		Select("countries.name as country_name, COUNT(cities.id) as city_count").
		Joins("JOIN countries ON cities.country_id = countries.id").
		Group("countries.id, countries.name").
		Order("city_count DESC").
		Limit(10).
		Find(&countryStats)

	stats.CitiesPerCountry = make(map[string]int64)
	for _, cs := range countryStats {
		stats.CitiesPerCountry[cs.CountryName] = cs.CityCount
	}

	// Get cities per state (top 10)
	type StateStats struct {
		StateName string `json:"state_name"`
		CityCount int64  `json:"city_count"`
	}

	var stateStats []StateStats
	database.DB.Table("cities").
		Select("states.name as state_name, COUNT(cities.id) as city_count").
		Joins("JOIN states ON cities.state_id = states.id").
		Group("states.id, states.name").
		Order("city_count DESC").
		Limit(10).
		Find(&stateStats)

	stats.CitiesPerState = make(map[string]int64)
	for _, ss := range stateStats {
		stats.CitiesPerState[ss.StateName] = ss.CityCount
	}

	return c.JSON(stats)
}
