package middleware

import (
	"strings"
	"tsimserver/auth"
	"tsimserver/database"
	"tsimserver/models"
	"tsimserver/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired middleware checks if user is authenticated
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Bearer token required",
			})
		}

		// Validate token
		claims, err := utils.ValidateAccessToken(token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Check if user exists and is active
		var user models.User
		if err := database.DB.Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "User not found or inactive",
			})
		}

		// Check if session exists and is active
		var session models.Session
		if err := database.DB.Where("access_token = ? AND is_active = ? AND user_id = ?", token, true, claims.UserID).First(&session).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Session not found or expired",
			})
		}

		// Store user info in context
		c.Locals("user", &user)
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)

		return c.Next()
	}
}

// RequirePermission middleware checks if user has specific permission
func RequirePermission(resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check permission
		allowed, err := auth.CheckPermission(userID.(uint), resource, action)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Error checking permissions",
			})
		}

		if !allowed {
			return c.Status(403).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireRole middleware checks if user has specific role
func RequireRole(roleName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Get user roles
		roles, err := auth.GetUserRoles(userID.(uint))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Error checking roles",
			})
		}

		// Check if user has required role
		hasRole := false
		for _, role := range roles {
			if strings.EqualFold(role.Name, roleName) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(403).JSON(fiber.Map{
				"error": "Required role not found",
			})
		}

		return c.Next()
	}
}

// OptionalAuth middleware for optional authentication
func OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			return c.Next()
		}

		// Validate token
		claims, err := utils.ValidateAccessToken(token)
		if err != nil {
			return c.Next()
		}

		// Check if user exists and is active
		var user models.User
		if err := database.DB.Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
			return c.Next()
		}

		// Store user info in context
		c.Locals("user", &user)
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)

		return c.Next()
	}
}
