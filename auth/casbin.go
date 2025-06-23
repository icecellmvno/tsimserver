package auth

import (
	"log"
	"tsimserver/config"
	"tsimserver/database"
	"tsimserver/models"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

var Enforcer *casbin.Enforcer

// InitCasbin initializes Casbin enforcer
func InitCasbin() error {
	// Initialize Gorm adapter
	adapter, err := gormadapter.NewAdapterByDB(database.DB)
	if err != nil {
		return err
	}

	// Create enforcer
	Enforcer, err = casbin.NewEnforcer(config.AppConfig.Casbin.ModelPath, adapter)
	if err != nil {
		return err
	}

	// Load policies from database
	err = Enforcer.LoadPolicy()
	if err != nil {
		return err
	}

	log.Println("Casbin enforcer initialized successfully")
	return nil
}

// CheckPermission checks if user has permission for resource and action
func CheckPermission(userID uint, resource, action string) (bool, error) {
	// Get user roles
	var userRoles []models.UserRole
	if err := database.DB.Preload("Role").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return false, err
	}

	// Check permission for each role
	for _, userRole := range userRoles {
		if userRole.Role != nil && userRole.Role.IsActive {
			allowed, err := Enforcer.Enforce(userRole.Role.Name, resource, action)
			if err != nil {
				return false, err
			}
			if allowed {
				return true, nil
			}
		}
	}

	return false, nil
}

// AddRoleForUser assigns a role to user
func AddRoleForUser(userID uint, roleID uint) error {
	// Check if role assignment already exists
	var existingRole models.UserRole
	result := database.DB.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existingRole)
	if result.Error == nil {
		return nil // Already exists
	}

	// Create new role assignment
	userRole := models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := database.DB.Create(&userRole).Error; err != nil {
		return err
	}

	// Update Casbin policies
	return Enforcer.LoadPolicy()
}

// RemoveRoleForUser removes a role from user
func RemoveRoleForUser(userID uint, roleID uint) error {
	// Remove role assignment
	if err := database.DB.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error; err != nil {
		return err
	}

	// Update Casbin policies
	return Enforcer.LoadPolicy()
}

// GetUserRoles returns all roles for a user
func GetUserRoles(userID uint) ([]models.Role, error) {
	var roles []models.Role
	err := database.DB.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.is_active = ?", userID, true).
		Find(&roles).Error

	return roles, err
}

// GetRolePermissions returns all permissions for a role
func GetRolePermissions(roleID uint) ([]models.Permission, error) {
	var permissions []models.Permission
	err := database.DB.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.is_active = ?", roleID, true).
		Find(&permissions).Error

	return permissions, err
}

// SyncPoliciesFromDatabase syncs Casbin policies with database
func SyncPoliciesFromDatabase() error {
	// Clear existing policies
	Enforcer.ClearPolicy()

	// Load role permissions from database
	var rolePermissions []models.RolePermission
	err := database.DB.Preload("Role").Preload("Permission").Find(&rolePermissions).Error
	if err != nil {
		return err
	}

	// Add policies to Casbin
	for _, rp := range rolePermissions {
		if rp.Role != nil && rp.Permission != nil && rp.Role.IsActive && rp.Permission.IsActive {
			_, err := Enforcer.AddPolicy(rp.Role.Name, rp.Permission.Resource, rp.Permission.Action)
			if err != nil {
				log.Printf("Error adding policy: %v", err)
			}
		}
	}

	// Save policies
	return Enforcer.SavePolicy()
}
