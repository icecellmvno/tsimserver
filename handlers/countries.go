package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetCountries returns all countries with pagination
func GetCountries(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	regionID := c.Query("region_id", "")
	subregionID := c.Query("subregion_id", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Country{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ? OR iso2 ILIKE ? OR iso3 ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Filter by region
	if regionID != "" {
		query = query.Where("region_id = ?", regionID)
	}

	// Filter by subregion
	if subregionID != "" {
		query = query.Where("subregion_id = ?", subregionID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get countries with pagination
	var countries []models.Country
	result := query.Offset(offset).Limit(limit).
		Preload("RegionModel").
		Preload("SubregionModel").
		Preload("States").
		Find(&countries)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch countries",
		})
	}

	return c.JSON(fiber.Map{
		"countries": countries,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetCountry returns a specific country by ID
func GetCountry(c *fiber.Ctx) error {
	countryIDStr := c.Params("id")
	countryID, err := strconv.ParseInt(countryIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid country ID",
		})
	}

	var country models.Country
	result := database.DB.Preload("RegionModel").
		Preload("SubregionModel").
		Preload("States").
		Where("id = ?", countryID).
		First(&country)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Country not found",
		})
	}

	return c.JSON(country)
}

// GetCountryByISO returns a country by ISO2 or ISO3 code
func GetCountryByISO(c *fiber.Ctx) error {
	isoCode := c.Params("iso")

	var country models.Country
	result := database.DB.Preload("RegionModel").
		Preload("SubregionModel").
		Preload("States").
		Where("iso2 = ? OR iso3 = ?", isoCode, isoCode).
		First(&country)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Country not found",
		})
	}

	return c.JSON(country)
}

// CreateCountry creates a new country
func CreateCountry(c *fiber.Ctx) error {
	var country models.Country

	if err := c.BodyParser(&country); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if country.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Country name is required",
		})
	}

	// Create country
	if err := database.DB.Create(&country).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create country",
		})
	}

	return c.Status(201).JSON(country)
}

// UpdateCountry updates an existing country
func UpdateCountry(c *fiber.Ctx) error {
	countryIDStr := c.Params("id")
	countryID, err := strconv.ParseInt(countryIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid country ID",
		})
	}

	var country models.Country
	if err := database.DB.Where("id = ?", countryID).First(&country).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Country not found",
		})
	}

	var updateData models.Country
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if updateData.Name != "" {
		country.Name = updateData.Name
	}
	if updateData.ISO2 != nil {
		country.ISO2 = updateData.ISO2
	}
	if updateData.ISO3 != nil {
		country.ISO3 = updateData.ISO3
	}
	if updateData.Capital != nil {
		country.Capital = updateData.Capital
	}
	if updateData.Currency != nil {
		country.Currency = updateData.Currency
	}
	if updateData.CurrencyName != nil {
		country.CurrencyName = updateData.CurrencyName
	}
	if updateData.PhoneCode != nil {
		country.PhoneCode = updateData.PhoneCode
	}
	if updateData.RegionID != nil {
		country.RegionID = updateData.RegionID
	}
	if updateData.SubregionID != nil {
		country.SubregionID = updateData.SubregionID
	}
	country.Flag = updateData.Flag

	if err := database.DB.Save(&country).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update country",
		})
	}

	return c.JSON(country)
}

// DeleteCountry deletes a country
func DeleteCountry(c *fiber.Ctx) error {
	countryIDStr := c.Params("id")
	countryID, err := strconv.ParseInt(countryIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid country ID",
		})
	}

	// Check if country has states or cities
	var stateCount, cityCount int64
	database.DB.Model(&models.State{}).Where("country_id = ?", countryID).Count(&stateCount)
	database.DB.Model(&models.City{}).Where("country_id = ?", countryID).Count(&cityCount)

	if stateCount > 0 || cityCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete country with associated states or cities",
		})
	}

	if err := database.DB.Delete(&models.Country{}, countryID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete country",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Country deleted successfully",
	})
}

// GetCountryStates returns states of a specific country
func GetCountryStates(c *fiber.Ctx) error {
	countryIDStr := c.Params("id")
	countryID, err := strconv.ParseInt(countryIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid country ID",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 100)
	offset := (page - 1) * limit

	var states []models.State
	var total int64

	query := database.DB.Where("country_id = ?", countryID)
	query.Model(&models.State{}).Count(&total)

	result := query.Offset(offset).Limit(limit).
		Preload("Cities").
		Find(&states)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch states",
		})
	}

	return c.JSON(fiber.Map{
		"states": states,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetCountryCities returns cities of a specific country
func GetCountryCities(c *fiber.Ctx) error {
	countryIDStr := c.Params("id")
	countryID, err := strconv.ParseInt(countryIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid country ID",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 100)
	offset := (page - 1) * limit

	var cities []models.City
	var total int64

	query := database.DB.Where("country_id = ?", countryID)
	query.Model(&models.City{}).Count(&total)

	result := query.Offset(offset).Limit(limit).
		Preload("State").
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
