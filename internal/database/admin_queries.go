package database

import (
	"errors"

	"moebot-next/internal/models"

	"gorm.io/gorm"
)

// --- Admin User Queries ---

// CountAdminUsers returns the number of registered console accounts.
func (d *DB) CountAdminUsers() (int64, error) {
	var n int64
	err := d.Model(&models.AdminUser{}).Count(&n).Error
	return n, err
}

// GetAdminUser returns the (single) admin account, or gorm.ErrRecordNotFound.
//
// 当前设计仅支持单一管理员账号，若未来扩展为多账号，调用方应改用按 username 查询。
func (d *DB) GetAdminUser() (*models.AdminUser, error) {
	var u models.AdminUser
	err := d.Order("id ASC").First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAdminUserByUsername finds an admin account by login name.
func (d *DB) GetAdminUserByUsername(username string) (*models.AdminUser, error) {
	var u models.AdminUser
	err := d.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateAdminUser inserts a new admin account. Caller must pre-hash the password.
func (d *DB) CreateAdminUser(u *models.AdminUser) error {
	return d.Create(u).Error
}

// UpdateAdminPassword overwrites the bcrypt hash for an existing account.
func (d *DB) UpdateAdminPassword(id uint, passwordHash string) error {
	return d.Model(&models.AdminUser{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

// --- App Meta (key/value) Queries ---

// GetAppMeta returns the value for a meta key, or empty string when missing.
func (d *DB) GetAppMeta(key string) (string, error) {
	var m models.AppMeta
	err := d.Where("key = ?", key).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return m.Value, nil
}

// SetAppMeta upserts a meta key/value pair.
func (d *DB) SetAppMeta(key, value string) error {
	var m models.AppMeta
	err := d.Where("key = ?", key).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return d.Create(&models.AppMeta{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	m.Value = value
	return d.Save(&m).Error
}
