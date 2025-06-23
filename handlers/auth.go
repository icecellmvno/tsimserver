package handlers

import (
	"time"
	"tsimserver/database"
	"tsimserver/models"
	"tsimserver/utils"

	"github.com/gofiber/fiber/v2"
)

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents register request body
type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

// Login authenticates a user
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Find user by username or email
	var user models.User
	result := database.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user)
	if result.Error != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Check if user is active
	if !user.IsActive {
		return c.Status(401).JSON(fiber.Map{
			"error": "Account is deactivated",
		})
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error generating access token",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error generating refresh token",
		})
	}

	// Create session
	session := models.Session{
		UserID:           user.ID,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        time.Now().Add(15 * time.Minute),   // 15 minutes
		RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		IPAddress:        c.IP(),
		UserAgent:        c.Get("User-Agent"),
		IsActive:         true,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error creating session",
		})
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	database.DB.Save(&user)

	// Prepare response
	user.Password = "" // Don't return password
	response := AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
	}

	return c.JSON(response)
}

// Register creates a new user account
func Register(c *fiber.Ctx) error {
	var req RegisterRequest
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
		IsActive:  true,
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

	// Assign default role (viewer)
	var defaultRole models.Role
	if err := database.DB.Where("name = ?", "viewer").First(&defaultRole).Error; err == nil {
		userRole := models.UserRole{
			UserID: user.ID,
			RoleID: defaultRole.ID,
		}
		database.DB.Create(&userRole)
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error generating access token",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error generating refresh token",
		})
	}

	// Create session
	session := models.Session{
		UserID:           user.ID,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        time.Now().Add(15 * time.Minute),
		RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IPAddress:        c.IP(),
		UserAgent:        c.Get("User-Agent"),
		IsActive:         true,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error creating session",
		})
	}

	// Prepare response
	user.Password = "" // Don't return password
	response := AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
	}

	return c.Status(201).JSON(response)
}

// RefreshToken refreshes access token using refresh token
func RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate refresh token
	_, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
		})
	}

	// Find session
	var session models.Session
	if err := database.DB.Where("refresh_token = ? AND is_active = ?", req.RefreshToken, true).First(&session).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Session not found or expired",
		})
	}

	// Check if refresh token is expired
	if time.Now().After(session.RefreshExpiresAt) {
		// Deactivate session
		session.IsActive = false
		database.DB.Save(&session)

		return c.Status(401).JSON(fiber.Map{
			"error": "Refresh token expired",
		})
	}

	// Get user
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", session.UserID, true).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not found or inactive",
		})
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error generating access token",
		})
	}

	// Update session
	session.AccessToken = newAccessToken
	session.ExpiresAt = time.Now().Add(15 * time.Minute)
	database.DB.Save(&session)

	// Prepare response
	user.Password = "" // Don't return password
	response := AuthResponse{
		User:         &user,
		AccessToken:  newAccessToken,
		RefreshToken: req.RefreshToken,
		ExpiresAt:    session.ExpiresAt,
	}

	return c.JSON(response)
}

// Logout invalidates current session
func Logout(c *fiber.Ctx) error {
	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.JSON(fiber.Map{
			"message": "Already logged out",
		})
	}

	token := utils.ExtractTokenFromHeader(authHeader)
	if token == "" {
		return c.JSON(fiber.Map{
			"message": "Already logged out",
		})
	}

	// Deactivate session
	database.DB.Model(&models.Session{}).
		Where("access_token = ?", token).
		Update("is_active", false)

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// GetProfile returns current user profile
func GetProfile(c *fiber.Ctx) error {
	// Get user from context
	user := c.Locals("user").(*models.User)

	// Load user roles
	var userRoles []models.UserRole
	database.DB.Preload("Role").Where("user_id = ?", user.ID).Find(&userRoles)

	// Extract role names
	var roles []string
	for _, userRole := range userRoles {
		if userRole.Role != nil && userRole.Role.IsActive {
			roles = append(roles, userRole.Role.Name)
		}
	}

	return c.JSON(fiber.Map{
		"user":  user,
		"roles": roles,
	})
}

// UpdateProfile updates current user profile
func UpdateProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var updateData struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
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

	// Save user
	if err := database.DB.Save(user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error updating profile",
		})
	}

	// Don't return password
	user.Password = ""

	return c.JSON(fiber.Map{
		"user":    user,
		"message": "Profile updated successfully",
	})
}

// ChangePassword changes user password
func ChangePassword(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,min=6"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Current password is incorrect",
		})
	}

	// Update password
	user.Password = req.NewPassword
	if err := user.HashPassword(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	// Save user
	if err := database.DB.Save(user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error updating password",
		})
	}

	// Invalidate all sessions except current one
	currentToken := utils.ExtractTokenFromHeader(c.Get("Authorization"))
	database.DB.Model(&models.Session{}).
		Where("user_id = ? AND access_token != ?", user.ID, currentToken).
		Update("is_active", false)

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}
