package db

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"rionexgate/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	gdb, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &DB{gdb}, nil
}

func (d *DB) AutoMigrate() error {
	return d.DB.AutoMigrate(&models.User{}, &models.Traffic{}, &models.Node{})
}

func (d *DB) SeedDefaultNode() error {
	var count int64
	d.Model(&models.Node{}).Count(&count)
	if count > 0 {
		return nil
	}
	return d.Create(&models.Node{Name: "default", Address: "127.0.0.1", Port: 443, Active: true}).Error
}

func (d *DB) ListUsers() ([]models.User, error) {
	var users []models.User
	err := d.Order("created_at desc").Find(&users).Error
	return users, err
}

func (d *DB) GetUser(id uint) (*models.User, error) {
	var user models.User
	err := d.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *DB) GetUserByUUID(uuidStr string) (*models.User, error) {
	var user models.User
	err := d.Where("uuid = ?", uuidStr).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type CreateUserInput struct {
	Email      string
	TrafficGB  int64
	ExpireDays int
}

func (d *DB) CreateUser(in CreateUserInput) (*models.User, error) {
	if in.Email == "" {
		return nil, errors.New("email is required")
	}
	expires := time.Now().AddDate(0, 0, in.ExpireDays)
	if in.ExpireDays <= 0 {
		expires = time.Now().AddDate(0, 0, 30)
	}
	user := &models.User{
		UUID:      uuid.New().String(),
		Email:     in.Email,
		TrafficGB: in.TrafficGB,
		ExpiresAt: expires,
		Active:    true,
	}
	if err := d.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

type UpdateUserInput struct {
	Email     *string
	TrafficGB *int64
	ExpiresAt *time.Time
	Active    *bool
}

func (d *DB) UpdateUser(id uint, in UpdateUserInput) (*models.User, error) {
	user, err := d.GetUser(id)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if in.Email != nil {
		updates["email"] = *in.Email
	}
	if in.TrafficGB != nil {
		updates["traffic_gb"] = *in.TrafficGB
	}
	if in.ExpiresAt != nil {
		updates["expires_at"] = *in.ExpiresAt
	}
	if in.Active != nil {
		updates["active"] = *in.Active
	}
	if len(updates) > 0 {
		if err := d.Model(user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return d.GetUser(id)
}

func (d *DB) DeleteUser(id uint) error {
	return d.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&models.Traffic{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.User{}, id).Error
	})
}

func (d *DB) ListActiveUsers() ([]models.User, error) {
	var users []models.User
	err := d.Where("active = ?", true).Find(&users).Error
	return users, err
}

func (d *DB) UpdateUserUsedBytes(id uint, usedBytes int64) error {
	return d.Model(&models.User{}).Where("id = ?", id).Update("used_bytes", usedBytes).Error
}

func (d *DB) RecordTraffic(userID uint, up, down int64) error {
	return d.Create(&models.Traffic{
		UserID:     userID,
		BytesUp:    up,
		BytesDown:  down,
		RecordedAt: time.Now(),
	}).Error
}

func (d *DB) TotalUsedBytes() (int64, error) {
	var total int64
	err := d.Model(&models.User{}).Select("COALESCE(SUM(used_bytes), 0)").Scan(&total).Error
	return total, err
}

func (d *DB) TrafficHistory(since time.Time) ([]models.Traffic, error) {
	var records []models.Traffic
	err := d.Where("recorded_at >= ?", since).Order("recorded_at asc").Find(&records).Error
	return records, err
}

func (d *DB) UserTrafficHistory(userID uint, since time.Time) ([]models.Traffic, error) {
	var records []models.Traffic
	err := d.Where("user_id = ? AND recorded_at >= ?", userID, since).Order("recorded_at asc").Find(&records).Error
	return records, err
}
