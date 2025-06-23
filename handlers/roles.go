package handlers

import (
	"strconv"
	"tsimserver/auth"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetRoles returns all roles with pagination
func GetRoles(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Role{})

	// Search functionality
	if search != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get roles with pagination
	var roles []models.Role
	result := query.Offset(offset).Limit(limit).
		Preload("UserRoles").
		Preload("RolePermissions").
		Find(&roles)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch roles",
		})
	}

	return c.JSON(fiber.Map{
		"roles": roles,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetRole returns a specific role by ID
func GetRole(c *fiber.Ctx) error {
	roleIDStr := c.Params("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	var role models.Role
	result := database.DB.Preload("UserRoles").
		Preload("RolePermissions.Permission").
		Where("id = ?", uint(roleID)).
		First(&role)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	return c.JSON(role)
}

// CreateRole creates a new role
func CreateRole(c *fiber.Ctx) error {
	var role models.Role

	if err := c.BodyParser(&role); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if role.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Role name is required",
		})
	}

	// Check if role name already exists
	var existingRole models.Role
	if result := database.DB.Where("name = ?", role.Name).First(&existingRole); result.Error == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Role name already exists",
		})
	}

	// Create role
	if err := database.DB.Create(&role).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create role",
		})
	}

	return c.Status(201).JSON(role)
}

// UpdateRole updates an existing role
func UpdateRole(c *fiber.Ctx) error {
	roleIDStr := c.Params("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	var role models.Role
	if err := database.DB.Where("id = ?", uint(roleID)).First(&role).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	var updateData models.Role
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if role name already exists (for other roles)
	if updateData.Name != "" && updateData.Name != role.Name {
		var existingRole models.Role
		if result := database.DB.Where("name = ? AND id != ?", updateData.Name, role.ID).First(&existingRole); result.Error == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Role name already exists",
			})
		}
		role.Name = updateData.Name
	}

	// Update fields
	if updateData.DisplayName != "" {
		role.DisplayName = updateData.DisplayName
	}
	if updateData.Description != "" {
		role.Description = updateData.Description
	}
	role.IsActive = updateData.IsActive

	if err := database.DB.Save(&role).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update role",
		})
	}

	return c.JSON(role)
}

// DeleteRole deletes a role
func DeleteRole(c *fiber.Ctx) error {
	roleIDStr := c.Params("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	// Check if role has users
	var userCount int64
	database.DB.Model(&models.UserRole{}).Where("role_id = ?", uint(roleID)).Count(&userCount)

	if userCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete role with assigned users",
		})
	}

	// Delete role permissions first
	database.DB.Where("role_id = ?", uint(roleID)).Delete(&models.RolePermission{})

	// Delete role
	if err := database.DB.Delete(&models.Role{}, uint(roleID)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete role",
		})
	}

	// Sync Casbin policies
	auth.SyncPoliciesFromDatabase()

	return c.JSON(fiber.Map{
		"message": "Role deleted successfully",
	})
}

// AssignPermissionToRole assigns permission to role
func AssignPermissionToRole(c *fiber.Ctx) error {
	roleIDStr := c.Params("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	var req struct {
		PermissionID uint `json:"permission_id" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if role exists
	var role models.Role
	if err := database.DB.Where("id = ?", uint(roleID)).First(&role).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Check if permission exists
	var permission models.Permission
	if err := database.DB.Where("id = ?", req.PermissionID).First(&permission).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Permission not found",
		})
	}

	// Check if assignment already exists
	var existingAssignment models.RolePermission
	result := database.DB.Where("role_id = ? AND permission_id = ?", uint(roleID), req.PermissionID).First(&existingAssignment)
	if result.Error == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Permission already assigned to role",
		})
	}

	// Create assignment
	assignment := models.RolePermission{
		RoleID:       uint(roleID),
		PermissionID: req.PermissionID,
	}

	if err := database.DB.Create(&assignment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to assign permission to role",
		})
	}

	// Sync Casbin policies
	auth.SyncPoliciesFromDatabase()

	return c.Status(201).JSON(fiber.Map{
		"message": "Permission assigned to role successfully",
	})
}

// RemovePermissionFromRole removes permission from role
func RemovePermissionFromRole(c *fiber.Ctx) error {
	roleIDStr := c.Params("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	permissionIDStr := c.Params("permission_id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid permission ID",
		})
	}

	// Remove assignment
	if err := database.DB.Where("role_id = ? AND permission_id = ?", uint(roleID), uint(permissionID)).Delete(&models.RolePermission{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to remove permission from role",
		})
	}

	// Sync Casbin policies
	auth.SyncPoliciesFromDatabase()

	return c.JSON(fiber.Map{
		"message": "Permission removed from role successfully",
	})
}

// GetRoleUsers returns users assigned to a role
func GetRoleUsers(c *fiber.Ctx) error {
	roleIDStr := c.Params("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	var users []models.User
	result := database.DB.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ? AND users.is_active = ?", uint(roleID), true).
		Find(&users)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch role users",
		})
	}

	// Don't return passwords
	for i := range users {
		users[i].Password = ""
	}

	return c.JSON(fiber.Map{
		"users": users,
		"count": len(users),
	})
}
