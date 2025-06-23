package handlers

import (
	"strconv"
	"tsimserver/auth"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/gofiber/fiber/v2"
)

// GetUsers returns all users with pagination and roles
func GetUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	search := c.Query("search", "")
	role := c.Query("role", "")
	isActive := c.Query("is_active", "")

	offset := (page - 1) * limit

	query := database.DB.Model(&models.User{})

	// Search functionality
	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Filter by active status
	if isActive != "" {
		if isActive == "true" {
			query = query.Where("is_active = ?", true)
		} else if isActive == "false" {
			query = query.Where("is_active = ?", false)
		}
	}

	// Filter by role
	if role != "" {
		query = query.Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Joins("JOIN roles ON user_roles.role_id = roles.id").
			Where("roles.name = ?", role)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get users with pagination
	var users []models.User
	result := query.Offset(offset).Limit(limit).
		Preload("UserRoles.Role").
		Find(&users)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	// Don't return passwords
	for i := range users {
		users[i].Password = ""
	}

	return c.JSON(fiber.Map{
		"users": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetUser returns a specific user by ID
func GetUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var user models.User
	result := database.DB.Preload("UserRoles.Role").
		Where("id = ?", uint(userID)).
		First(&user)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Don't return password
	user.Password = ""

	return c.JSON(user)
}

// CreateUser creates a new user (admin only)
func CreateUser(c *fiber.Ctx) error {
	var req struct {
		Username  string `json:"username" validate:"required,min=3,max=50"`
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=6"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		IsActive  bool   `json:"is_active"`
		RoleIDs   []uint `json:"role_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if username already exists
	var existingUser models.User
	if result := database.DB.Where("username = ?", req.Username).First(&existingUser); result.Error == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	// Check if email already exists
	if result := database.DB.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	// Create new user
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  req.IsActive,
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	// Save user
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error creating user",
		})
	}

	// Assign roles
	for _, roleID := range req.RoleIDs {
		userRole := models.UserRole{
			UserID: user.ID,
			RoleID: roleID,
		}
		database.DB.Create(&userRole)
	}

	// Don't return password
	user.Password = ""

	return c.Status(201).JSON(user)
}

// UpdateUser updates an existing user
func UpdateUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var user models.User
	if err := database.DB.Where("id = ?", uint(userID)).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var updateData struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		IsActive  *bool  `json:"is_active"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if email already exists (for other users)
	if updateData.Email != "" && updateData.Email != user.Email {
		var existingUser models.User
		if result := database.DB.Where("email = ? AND id != ?", updateData.Email, user.ID).First(&existingUser); result.Error == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Email already exists",
			})
		}
		user.Email = updateData.Email
	}

	// Update fields
	if updateData.FirstName != "" {
		user.FirstName = updateData.FirstName
	}
	if updateData.LastName != "" {
		user.LastName = updateData.LastName
	}
	if updateData.IsActive != nil {
		user.IsActive = *updateData.IsActive
	}

	// Save user
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error updating user",
		})
	}

	// Don't return password
	user.Password = ""

	return c.JSON(user)
}

// DeleteUser deletes a user
func DeleteUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Check if it's the current user
	currentUserID := c.Locals("user_id").(uint)
	if uint(userID) == currentUserID {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete your own account",
		})
	}

	// Soft delete user roles first
	database.DB.Where("user_id = ?", uint(userID)).Delete(&models.UserRole{})

	// Deactivate sessions
	database.DB.Model(&models.Session{}).Where("user_id = ?", uint(userID)).Update("is_active", false)

	// Soft delete user
	if err := database.DB.Delete(&models.User{}, uint(userID)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// AssignRoleToUser assigns a role to user
func AssignRoleToUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		RoleID uint `json:"role_id" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("id = ?", uint(userID)).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Check if role exists
	var role models.Role
	if err := database.DB.Where("id = ?", req.RoleID).First(&role).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Use auth package to assign role
	if err := auth.AddRoleForUser(uint(userID), req.RoleID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to assign role to user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role assigned to user successfully",
	})
}

// RemoveRoleFromUser removes a role from user
func RemoveRoleFromUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	roleIDStr := c.Params("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	// Use auth package to remove role
	if err := auth.RemoveRoleForUser(uint(userID), uint(roleID)); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to remove role from user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role removed from user successfully",
	})
}

// GetUserRoles returns all roles for a user
func GetUserRoles(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	roles, err := auth.GetUserRoles(uint(userID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch user roles",
		})
	}

	return c.JSON(fiber.Map{
		"roles": roles,
		"count": len(roles),
	})
}

// GetUserSessions returns all active sessions for a user
func GetUserSessions(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var sessions []models.Session
	result := database.DB.Where("user_id = ? AND is_active = ?", uint(userID), true).
		Order("created_at DESC").
		Find(&sessions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch user sessions",
		})
	}

	// Don't return tokens
	for i := range sessions {
		sessions[i].AccessToken = ""
		sessions[i].RefreshToken = ""
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// RevokeUserSession revokes a specific user session
func RevokeUserSession(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	sessionIDStr := c.Params("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid session ID",
		})
	}

	// Deactivate session
	result := database.DB.Model(&models.Session{}).
		Where("id = ? AND user_id = ?", uint(sessionID), uint(userID)).
		Update("is_active", false)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to revoke session",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Session not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Session revoked successfully",
	})
}
