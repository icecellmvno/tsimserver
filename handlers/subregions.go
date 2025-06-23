package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetSubregions returns all subregions with pagination
func GetSubregions(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	regionID := c.Query("region_id", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Subregion{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Filter by region
	if regionID != "" {
		query = query.Where("region_id = ?", regionID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get subregions with pagination
	var subregions []models.Subregion
	result := query.Offset(offset).Limit(limit).
		Preload("Region").
		Preload("Countries").
		Find(&subregions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch subregions",
		})
	}

	return c.JSON(fiber.Map{
		"subregions": subregions,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSubregion returns a specific subregion by ID
func GetSubregion(c *fiber.Ctx) error {
	subregionIDStr := c.Params("id")
	subregionID, err := strconv.ParseInt(subregionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid subregion ID",
		})
	}

	var subregion models.Subregion
	result := database.DB.Preload("Region").
		Preload("Countries").
		Where("id = ?", subregionID).
		First(&subregion)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Subregion not found",
		})
	}

	return c.JSON(subregion)
}

// CreateSubregion creates a new subregion
func CreateSubregion(c *fiber.Ctx) error {
	var subregion models.Subregion

	if err := c.BodyParser(&subregion); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if subregion.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Subregion name is required",
		})
	}
	if subregion.RegionID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Region ID is required",
		})
	}

	// Create subregion
	if err := database.DB.Create(&subregion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create subregion",
		})
	}

	return c.Status(201).JSON(subregion)
}

// UpdateSubregion updates an existing subregion
func UpdateSubregion(c *fiber.Ctx) error {
	subregionIDStr := c.Params("id")
	subregionID, err := strconv.ParseInt(subregionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid subregion ID",
		})
	}

	var subregion models.Subregion
	if err := database.DB.Where("id = ?", subregionID).First(&subregion).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Subregion not found",
		})
	}

	var updateData models.Subregion
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if updateData.Name != "" {
		subregion.Name = updateData.Name
	}
	if updateData.RegionID != 0 {
		subregion.RegionID = updateData.RegionID
	}
	if updateData.Translations != "" {
		subregion.Translations = updateData.Translations
	}
	if updateData.WikiDataID != nil {
		subregion.WikiDataID = updateData.WikiDataID
	}
	subregion.Flag = updateData.Flag

	if err := database.DB.Save(&subregion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update subregion",
		})
	}

	return c.JSON(subregion)
}

// DeleteSubregion deletes a subregion
func DeleteSubregion(c *fiber.Ctx) error {
	subregionIDStr := c.Params("id")
	subregionID, err := strconv.ParseInt(subregionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid subregion ID",
		})
	}

	// Check if subregion has countries
	var countryCount int64
	database.DB.Model(&models.Country{}).Where("subregion_id = ?", subregionID).Count(&countryCount)

	if countryCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete subregion with associated countries",
		})
	}

	if err := database.DB.Delete(&models.Subregion{}, subregionID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete subregion",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Subregion deleted successfully",
	})
}

// GetSubregionCountries returns countries of a specific subregion
func GetSubregionCountries(c *fiber.Ctx) error {
	subregionIDStr := c.Params("id")
	subregionID, err := strconv.ParseInt(subregionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid subregion ID",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 100)
	offset := (page - 1) * limit

	var countries []models.Country
	var total int64

	query := database.DB.Where("subregion_id = ?", subregionID)
	query.Model(&models.Country{}).Count(&total)

	result := query.Offset(offset).Limit(limit).
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
