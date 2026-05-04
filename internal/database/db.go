package database

import (
	"fmt"
	"os"
	"path/filepath"

	"moebot-next/internal/config"
	"moebot-next/internal/models"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps the GORM database connection.
type DB struct {
	*gorm.DB
}

// New creates a new database connection and runs auto-migration.
func New(cfg config.DatabaseConfig) (*DB, error) {
	// Ensure the data directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Warn().Err(err).Msg("Failed to enable WAL mode")
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.SuiteSetting{},
		&models.Group{},
		&models.CommandStat{},
		&models.ImageCache{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}
	if db.Migrator().HasIndex(&models.User{}, "idx_platform_user") {
		if err := db.Migrator().DropIndex(&models.User{}, "idx_platform_user"); err != nil {
			log.Warn().Err(err).Msg("Failed to drop legacy single-server user index")
		}
	}
	if err := db.Model(&models.User{}).Where("server_region = '' OR server_region IS NULL").Update("server_region", config.RegionJP).Error; err != nil {
		log.Warn().Err(err).Msg("Failed to backfill user server region")
	}

	log.Info().Str("path", cfg.Path).Msg("Database initialized")
	return &DB{db}, nil
}

// Close gracefully shuts down the database connection.
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
