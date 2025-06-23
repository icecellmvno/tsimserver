package cache

import (
	"context"
	"fmt"
	"log"
	"time"
	"tsimserver/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

// Connect establishes Redis connection
func Connect() error {
	cfg := config.AppConfig.Redis

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established successfully")
	return nil
}

// Close closes Redis connection
func Close() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// Set stores a key-value pair with expiration
func Set(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// Get retrieves value by key
func Get(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

// Delete removes a key
func Delete(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// Exists checks if key exists
func Exists(key string) (bool, error) {
	count, err := RedisClient.Exists(ctx, key).Result()
	return count > 0, err
}

// SetDeviceStatus stores device connection status
func SetDeviceStatus(deviceID string, status string) error {
	key := fmt.Sprintf("device:status:%s", deviceID)
	return Set(key, status, 5*time.Minute)
}

// GetDeviceStatus retrieves device connection status
func GetDeviceStatus(deviceID string) (string, error) {
	key := fmt.Sprintf("device:status:%s", deviceID)
	return Get(key)
}

// SetDeviceConnection stores device WebSocket connection info
func SetDeviceConnection(deviceID string, connectionID string) error {
	key := fmt.Sprintf("device:connection:%s", deviceID)
	return Set(key, connectionID, 24*time.Hour)
}

// GetDeviceConnection retrieves device WebSocket connection info
func GetDeviceConnection(deviceID string) (string, error) {
	key := fmt.Sprintf("device:connection:%s", deviceID)
	return Get(key)
}

// RemoveDeviceConnection removes device WebSocket connection info
func RemoveDeviceConnection(deviceID string) error {
	key := fmt.Sprintf("device:connection:%s", deviceID)
	return Delete(key)
}

// SetSession stores user session
func SetSession(token string, userID uint, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", token)
	return Set(key, userID, expiration)
}

// GetSession retrieves user session
func GetSession(token string) (string, error) {
	key := fmt.Sprintf("session:%s", token)
	return Get(key)
}

// RemoveSession removes user session
func RemoveSession(token string) error {
	key := fmt.Sprintf("session:%s", token)
	return Delete(key)
}

// GetClient returns Redis client instance
func GetClient() *redis.Client {
	return RedisClient
}
