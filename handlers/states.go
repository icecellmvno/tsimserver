package handlers

import (
	"strconv"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetStates returns all states with pagination
func GetStates(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	countryID := c.Query("country_id", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.State{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ? OR iso2 ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Filter by country
	if countryID != "" {
		query = query.Where("country_id = ?", countryID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get states with pagination
	var states []models.State
	result := query.Offset(offset).Limit(limit).
		Preload("Country").
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

// GetState returns a specific state by ID
func GetState(c *fiber.Ctx) error {
	stateIDStr := c.Params("id")
	stateID, err := strconv.ParseInt(stateIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid state ID",
		})
	}

	var state models.State
	result := database.DB.Preload("Country").
		Preload("Cities").
		Where("id = ?", stateID).
		First(&state)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "State not found",
		})
	}

	return c.JSON(state)
}

// CreateState creates a new state
func CreateState(c *fiber.Ctx) error {
	var state models.State

	if err := c.BodyParser(&state); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if state.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "State name is required",
		})
	}
	if state.CountryID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Country ID is required",
		})
	}
	if state.CountryCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Country code is required",
		})
	}

	// Create state
	if err := database.DB.Create(&state).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create state",
		})
	}

	return c.Status(201).JSON(state)
}

// UpdateState updates an existing state
func UpdateState(c *fiber.Ctx) error {
	stateIDStr := c.Params("id")
	stateID, err := strconv.ParseInt(stateIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid state ID",
		})
	}

	var state models.State
	if err := database.DB.Where("id = ?", stateID).First(&state).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "State not found",
		})
	}

	var updateData models.State
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if updateData.Name != "" {
		state.Name = updateData.Name
	}
	if updateData.CountryID != 0 {
		state.CountryID = updateData.CountryID
	}
	if updateData.CountryCode != "" {
		state.CountryCode = updateData.CountryCode
	}
	if updateData.FipsCode != nil {
		state.FipsCode = updateData.FipsCode
	}
	if updateData.ISO2 != nil {
		state.ISO2 = updateData.ISO2
	}
	if updateData.Type != nil {
		state.Type = updateData.Type
	}
	if updateData.Level != nil {
		state.Level = updateData.Level
	}
	if updateData.ParentID != nil {
		state.ParentID = updateData.ParentID
	}
	if updateData.Native != nil {
		state.Native = updateData.Native
	}
	if updateData.Latitude != nil {
		state.Latitude = updateData.Latitude
	}
	if updateData.Longitude != nil {
		state.Longitude = updateData.Longitude
	}
	state.Flag = updateData.Flag

	if err := database.DB.Save(&state).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update state",
		})
	}

	return c.JSON(state)
}

// DeleteState deletes a state
func DeleteState(c *fiber.Ctx) error {
	stateIDStr := c.Params("id")
	stateID, err := strconv.ParseInt(stateIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid state ID",
		})
	}

	// Check if state has cities
	var cityCount int64
	database.DB.Model(&models.City{}).Where("state_id = ?", stateID).Count(&cityCount)

	if cityCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete state with associated cities",
		})
	}

	if err := database.DB.Delete(&models.State{}, stateID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete state",
		})
	}

	return c.JSON(fiber.Map{
		"message": "State deleted successfully",
	})
}

// GetStateCities returns cities of a specific state
func GetStateCities(c *fiber.Ctx) error {
	stateIDStr := c.Params("id")
	stateID, err := strconv.ParseInt(stateIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid state ID",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 100)
	search := c.Query("search", "")
	offset := (page - 1) * limit

	var cities []models.City
	var total int64

	query := database.DB.Where("state_id = ?", stateID)

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	query.Model(&models.City{}).Count(&total)

	result := query.Offset(offset).Limit(limit).
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
