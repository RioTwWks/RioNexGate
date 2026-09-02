package db

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"rionexgate/internal/models"

	"gorm.io/gorm"
)

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (d *DB) EnsureSubscriptionToken(userID uint) (string, error) {
	user, err := d.GetUser(userID)
	if err != nil {
		return "", err
	}
	if user.SubscriptionToken != "" {
		return user.SubscriptionToken, nil
	}
	token, err := generateToken(32)
	if err != nil {
		return "", err
	}
	if err := d.Model(user).Update("subscription_token", token).Error; err != nil {
		return "", err
	}
	return token, nil
}

func (d *DB) BackfillSubscriptionTokens() error {
	var users []models.User
	if err := d.Where("subscription_token = ? OR subscription_token IS NULL", "").Find(&users).Error; err != nil {
		return err
	}
	for _, user := range users {
		if user.SubscriptionToken != "" {
			continue
		}
		token, err := generateToken(32)
		if err != nil {
			return err
		}
		if err := d.Model(&user).Update("subscription_token", token).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) GetUserBySubscriptionToken(token string) (*models.User, error) {
	var user models.User
	err := d.Where("subscription_token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *DB) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := d.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *DB) CreateDevice(userID uint, label string) (*models.Device, error) {
	token, err := generateToken(32)
	if err != nil {
		return nil, err
	}
	if label == "" {
		label = "device"
	}
	device := &models.Device{
		UserID: userID,
		Token:  token,
		Label:  label,
	}
	if err := d.Create(device).Error; err != nil {
		return nil, err
	}
	return device, nil
}

func (d *DB) GetDeviceByToken(token string) (*models.Device, error) {
	var device models.Device
	err := d.Where("token = ?", token).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *DB) ListDevicesByUser(userID uint) ([]models.Device, error) {
	var devices []models.Device
	err := d.Where("user_id = ?", userID).Order("created_at desc").Find(&devices).Error
	return devices, err
}

func (d *DB) UpdateDeviceLastSeen(token string) error {
	now := time.Now()
	return d.Model(&models.Device{}).Where("token = ?", token).Update("last_seen_at", now).Error
}

func (d *DB) UpdateDeviceConfigCache(deviceID uint, configJSON, configHash string) error {
	return d.Model(&models.Device{}).Where("id = ?", deviceID).Updates(map[string]interface{}{
		"cached_config":      configJSON,
		"cached_config_hash": configHash,
	}).Error
}

func (d *DB) DeleteDevice(token string) error {
	return d.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("device_token = ?", token).Delete(&models.ClientStatsReport{}).Error; err != nil {
			return err
		}
		res := tx.Where("token = ?", token).Delete(&models.Device{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

type ClientStatsInput struct {
	DeviceToken string
	SessionID   string
	BytesIn     int64
	BytesOut    int64
}

func (d *DB) UpsertClientStats(in ClientStatsInput) error {
	if in.SessionID == "" {
		return errors.New("session_id is required")
	}
	now := time.Now()
	var existing models.ClientStatsReport
	err := d.Where("device_token = ? AND session_id = ?", in.DeviceToken, in.SessionID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return d.Create(&models.ClientStatsReport{
			DeviceToken: in.DeviceToken,
			SessionID:   in.SessionID,
			BytesIn:     in.BytesIn,
			BytesOut:    in.BytesOut,
			ReportedAt:  now,
		}).Error
	}
	if err != nil {
		return err
	}
	return d.Model(&existing).Updates(map[string]interface{}{
		"bytes_in":    in.BytesIn,
		"bytes_out":   in.BytesOut,
		"reported_at": now,
	}).Error
}

func (d *DB) ClientReportedBytesForUser(userID uint) (int64, error) {
	var devices []models.Device
	if err := d.Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		return 0, err
	}
	if len(devices) == 0 {
		return 0, nil
	}
	tokens := make([]string, len(devices))
	for i, dev := range devices {
		tokens[i] = dev.Token
	}
	var total int64
	err := d.Model(&models.ClientStatsReport{}).
		Select("COALESCE(SUM(bytes_in + bytes_out), 0)").
		Where("device_token IN ?", tokens).
		Scan(&total).Error
	return total, err
}

func (d *DB) CountActiveDevices(since time.Time) (int64, error) {
	var count int64
	err := d.Model(&models.Device{}).Where("last_seen_at >= ?", since).Count(&count).Error
	return count, err
}
