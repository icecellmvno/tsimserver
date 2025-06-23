package handlers

import (
	"strconv"
	"tsimserver/auth"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetPermissions returns all permissions with pagination
func GetPermissions(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	resource := c.Query("resource", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Permission{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ? OR description ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Filter by resource
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get permissions with pagination
	var permissions []models.Permission
	result := query.Offset(offset).Limit(limit).
		Preload("RolePermissions").
		Find(&permissions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch permissions",
		})
	}

	return c.JSON(fiber.Map{
		"permissions": permissions,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetPermission returns a specific permission by ID
func GetPermission(c *fiber.Ctx) error {
	permissionIDStr := c.Params("id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid permission ID",
		})
	}

	var permission models.Permission
	result := database.DB.Preload("RolePermissions.Role").
		Where("id = ?", uint(permissionID)).
		First(&permission)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Permission not found",
		})
	}

	return c.JSON(permission)
}

// CreatePermission creates a new permission
func CreatePermission(c *fiber.Ctx) error {
	var permission models.Permission

	if err := c.BodyParser(&permission); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if permission.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Permission name is required",
		})
	}

	if permission.Resource == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Resource is required",
		})
	}

	if permission.Action == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Action is required",
		})
	}

	// Check if permission name already exists
	var existingPermission models.Permission
	if result := database.DB.Where("name = ?", permission.Name).First(&existingPermission); result.Error == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Permission name already exists",
		})
	}

	// Create permission
	if err := database.DB.Create(&permission).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create permission",
		})
	}

	return c.Status(201).JSON(permission)
}

// UpdatePermission updates an existing permission
func UpdatePermission(c *fiber.Ctx) error {
	permissionIDStr := c.Params("id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid permission ID",
		})
	}

	var permission models.Permission
	if err := database.DB.Where("id = ?", uint(permissionID)).First(&permission).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Permission not found",
		})
	}

	var updateData models.Permission
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if permission name already exists (for other permissions)
	if updateData.Name != "" && updateData.Name != permission.Name {
		var existingPermission models.Permission
		if result := database.DB.Where("name = ? AND id != ?", updateData.Name, permission.ID).First(&existingPermission); result.Error == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Permission name already exists",
			})
		}
		permission.Name = updateData.Name
	}

	// Update fields
	if updateData.DisplayName != "" {
		permission.DisplayName = updateData.DisplayName
	}
	if updateData.Description != "" {
		permission.Description = updateData.Description
	}
	if updateData.Resource != "" {
		permission.Resource = updateData.Resource
	}
	if updateData.Action != "" {
		permission.Action = updateData.Action
	}
	permission.IsActive = updateData.IsActive

	if err := database.DB.Save(&permission).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update permission",
		})
	}

	// Sync Casbin policies
	auth.SyncPoliciesFromDatabase()

	return c.JSON(permission)
}

// DeletePermission deletes a permission
func DeletePermission(c *fiber.Ctx) error {
	permissionIDStr := c.Params("id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid permission ID",
		})
	}

	// Check if permission is assigned to roles
	var roleCount int64
	database.DB.Model(&models.RolePermission{}).Where("permission_id = ?", uint(permissionID)).Count(&roleCount)

	if roleCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete permission assigned to roles",
		})
	}

	// Delete permission
	if err := database.DB.Delete(&models.Permission{}, uint(permissionID)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete permission",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permission deleted successfully",
	})
}

// GetPermissionResources returns all unique resources
func GetPermissionResources(c *fiber.Ctx) error {
	var resources []string

	result := database.DB.Model(&models.Permission{}).
		Distinct("resource").
		Where("is_active = ?", true).
		Pluck("resource", &resources)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch resources",
		})
	}

	return c.JSON(fiber.Map{
		"resources": resources,
	})
}

// GetPermissionActions returns all unique actions
func GetPermissionActions(c *fiber.Ctx) error {
	var actions []string

	result := database.DB.Model(&models.Permission{}).
		Distinct("action").
		Where("is_active = ?", true).
		Pluck("action", &actions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch actions",
		})
	}

	return c.JSON(fiber.Map{
		"actions": actions,
	})
}

// BulkCreatePermissions creates multiple permissions at once
func BulkCreatePermissions(c *fiber.Ctx) error {
	var permissionsData struct {
		Permissions []models.Permission `json:"permissions" validate:"required"`
	}

	if err := c.BodyParser(&permissionsData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(permissionsData.Permissions) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "No permissions provided",
		})
	}

	// Validate and check for duplicates
	for i, permission := range permissionsData.Permissions {
		if permission.Name == "" || permission.Resource == "" || permission.Action == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "All permissions must have name, resource, and action",
				"index": i,
			})
		}

		// Check if permission name already exists
		var existingPermission models.Permission
		if result := database.DB.Where("name = ?", permission.Name).First(&existingPermission); result.Error == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Permission name already exists: " + permission.Name,
				"index": i,
			})
		}
	}

	// Create permissions
	if err := database.DB.Create(&permissionsData.Permissions).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create permissions",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":     "Permissions created successfully",
		"count":       len(permissionsData.Permissions),
		"permissions": permissionsData.Permissions,
	})
}
