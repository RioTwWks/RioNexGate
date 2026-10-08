package db

import (
	"errors"
	"time"

	"rionexgate/internal/models"

	"gorm.io/gorm"
)

type CreateInviteInput struct {
	UserID    uint
	Label     string
	MaxUses   int
	ExpiresIn time.Duration // 0 = no expiry
}

func (d *DB) CreateInvite(in CreateInviteInput) (*models.Invite, error) {
	if in.UserID == 0 {
		return nil, errors.New("user_id is required")
	}
	if _, err := d.GetUser(in.UserID); err != nil {
		return nil, err
	}
	token, err := generateToken(32)
	if err != nil {
		return nil, err
	}
	maxUses := in.MaxUses
	if maxUses <= 0 {
		maxUses = 1
	}
	inv := &models.Invite{
		UserID:  in.UserID,
		Token:   token,
		Label:   in.Label,
		MaxUses: maxUses,
	}
	if in.ExpiresIn > 0 {
		t := time.Now().Add(in.ExpiresIn)
		inv.ExpiresAt = &t
	}
	if err := d.Create(inv).Error; err != nil {
		return nil, err
	}
	return inv, nil
}

func (d *DB) ListInvitesByUser(userID uint) ([]models.Invite, error) {
	var invites []models.Invite
	err := d.Where("user_id = ?", userID).Order("created_at desc").Find(&invites).Error
	return invites, err
}

func (d *DB) GetInviteByIDForUser(userID, inviteID uint) (*models.Invite, error) {
	var inv models.Invite
	err := d.Where("id = ? AND user_id = ?", inviteID, userID).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (d *DB) GetInviteByToken(token string) (*models.Invite, error) {
	var inv models.Invite
	err := d.Where("token = ?", token).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// ConsumeInvite atomically increments used_count if the invite is still usable.
func (d *DB) ConsumeInvite(token string) (*models.Invite, error) {
	now := time.Now()
	res := d.Model(&models.Invite{}).
		Where("token = ? AND revoked_at IS NULL AND used_count < max_uses AND (expires_at IS NULL OR expires_at > ?)",
			token, now).
		UpdateColumn("used_count", gorm.Expr("used_count + 1"))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("invite not usable")
	}
	return d.GetInviteByToken(token)
}

func (d *DB) RevokeInvite(userID, inviteID uint) error {
	now := time.Now()
	res := d.Model(&models.Invite{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", inviteID, userID).
		Update("revoked_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
