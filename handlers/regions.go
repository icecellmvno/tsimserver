package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetRegions returns all regions with pagination
func GetRegions(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Region{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get regions with pagination
	var regions []models.Region
	result := query.Offset(offset).Limit(limit).
		Preload("Subregions").
		Preload("Countries").
		Find(&regions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch regions",
		})
	}

	return c.JSON(fiber.Map{
		"regions": regions,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetRegion returns a specific region by ID
func GetRegion(c *fiber.Ctx) error {
	regionIDStr := c.Params("id")
	regionID, err := strconv.ParseInt(regionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid region ID",
		})
	}

	var region models.Region
	result := database.DB.Preload("Subregions").
		Preload("Countries").
		Where("id = ?", regionID).
		First(&region)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Region not found",
		})
	}

	return c.JSON(region)
}

// CreateRegion creates a new region
func CreateRegion(c *fiber.Ctx) error {
	var region models.Region

	if err := c.BodyParser(&region); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if region.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Region name is required",
		})
	}

	// Create region
	if err := database.DB.Create(&region).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create region",
		})
	}

	return c.Status(201).JSON(region)
}

// UpdateRegion updates an existing region
func UpdateRegion(c *fiber.Ctx) error {
	regionIDStr := c.Params("id")
	regionID, err := strconv.ParseInt(regionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid region ID",
		})
	}

	var region models.Region
	if err := database.DB.Where("id = ?", regionID).First(&region).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Region not found",
		})
	}

	var updateData models.Region
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if updateData.Name != "" {
		region.Name = updateData.Name
	}
	if updateData.Translations != "" {
		region.Translations = updateData.Translations
	}
	if updateData.WikiDataID != nil {
		region.WikiDataID = updateData.WikiDataID
	}
	region.Flag = updateData.Flag

	if err := database.DB.Save(&region).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update region",
		})
	}

	return c.JSON(region)
}

// DeleteRegion deletes a region
func DeleteRegion(c *fiber.Ctx) error {
	regionIDStr := c.Params("id")
	regionID, err := strconv.ParseInt(regionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid region ID",
		})
	}

	// Check if region has subregions or countries
	var subregionCount, countryCount int64
	database.DB.Model(&models.Subregion{}).Where("region_id = ?", regionID).Count(&subregionCount)
	database.DB.Model(&models.Country{}).Where("region_id = ?", regionID).Count(&countryCount)

	if subregionCount > 0 || countryCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete region with associated subregions or countries",
		})
	}

	if err := database.DB.Delete(&models.Region{}, regionID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete region",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Region deleted successfully",
	})
}

// GetRegionSubregions returns subregions of a specific region
func GetRegionSubregions(c *fiber.Ctx) error {
	regionIDStr := c.Params("id")
	regionID, err := strconv.ParseInt(regionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid region ID",
		})
	}

	var subregions []models.Subregion
	result := database.DB.Where("region_id = ?", regionID).
		Preload("Countries").
		Find(&subregions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch subregions",
		})
	}

	return c.JSON(fiber.Map{
		"subregions": subregions,
		"count":      len(subregions),
	})
}

// GetRegionCountries returns countries of a specific region
func GetRegionCountries(c *fiber.Ctx) error {
	regionIDStr := c.Params("id")
	regionID, err := strconv.ParseInt(regionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid region ID",
		})
	}

	var countries []models.Country
	result := database.DB.Where("region_id = ?", regionID).
		Preload("States").
		Find(&countries)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch countries",
		})
	}

	return c.JSON(fiber.Map{
		"countries": countries,
		"count":     len(countries),
	})
}
