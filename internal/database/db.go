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

	// One-shot migration of legacy gateway default rules → default template (must run
	// BEFORE AutoMigrate drops the columns from the live schema). We snapshot the two
	// legacy columns from raw SQL so it works regardless of model state.
	legacyUserRules, legacyGroupRules := readLegacyGatewayDefaults(db)

	// Auto-migrate all models (this also creates filter_templates, adds template_id,
	// and removes default_*_rules columns from filter_gateways).
	if err := db.AutoMigrate(
		&models.User{},
		&models.SuiteSetting{},
		&models.Group{},
		&models.CommandStat{},
		&models.ImageCache{},
		&models.FilterGateway{},
		&models.FilterTemplate{},
		&models.FilterApp{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}
	for _, col := range []string{"default_user_id_rules", "default_group_id_rules"} {
		if db.Migrator().HasColumn(&models.FilterGateway{}, col) {
			if err := db.Migrator().DropColumn(&models.FilterGateway{}, col); err != nil {
				log.Warn().Err(err).Str("column", col).Msg("Failed to drop legacy filter gateway column")
			}
		}
	}
	if db.Migrator().HasIndex(&models.User{}, "idx_platform_user") {
		if err := db.Migrator().DropIndex(&models.User{}, "idx_platform_user"); err != nil {
			log.Warn().Err(err).Msg("Failed to drop legacy single-server user index")
		}
	}
	// After AutoMigrate, ensure the built-in default template exists and seed it
	// with the legacy gateway defaults if applicable (one-time migration).
	if err := seedDefaultFilterTemplate(db, legacyUserRules, legacyGroupRules); err != nil {
		log.Warn().Err(err).Msg("Failed to seed default filter template")
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

// readLegacyGatewayDefaults reads default_user_id_rules and default_group_id_rules
// from filter_gateways via raw SQL, before the AutoMigrate step removes those
// columns. Returns empty strings when the columns or table don't exist.
func readLegacyGatewayDefaults(db *gorm.DB) (user, group string) {
	if !db.Migrator().HasTable("filter_gateways") {
		return "", ""
	}
	row := struct {
		DefaultUserIDRules  string
		DefaultGroupIDRules string
	}{}
	// Use Raw + Scan; ignore errors (missing columns are expected on new installs).
	_ = db.Raw(`SELECT default_user_id_rules, default_group_id_rules FROM filter_gateways WHERE id = 1`).Row().Scan(&row.DefaultUserIDRules, &row.DefaultGroupIDRules)
	return row.DefaultUserIDRules, row.DefaultGroupIDRules
}

// seedDefaultFilterTemplate ensures the built-in default template exists. When
// legacyUser/legacyGroup are non-empty (i.e. a migration from a pre-template
// install), they are written to the new default template's rule fields.
func seedDefaultFilterTemplate(db *gorm.DB, legacyUser, legacyGroup string) error {
	var t models.FilterTemplate
	err := db.Where("name = ?", "default").First(&t).Error
	if err == nil {
		// If the template already exists with empty rules and we have legacy
		// values to migrate, fill them in once.
		dirty := false
		if legacyUser != "" && legacyUser != "{}" && (t.UserIDRules == "" || t.UserIDRules == "{}" || t.UserIDRules == `{"mode":"on","ids":[]}`) {
			t.UserIDRules = legacyUser
			dirty = true
		}
		if legacyGroup != "" && legacyGroup != "{}" && (t.GroupIDRules == "" || t.GroupIDRules == "{}" || t.GroupIDRules == `{"mode":"on","ids":[]}`) {
			t.GroupIDRules = legacyGroup
			dirty = true
		}
		if dirty {
			return db.Save(&t).Error
		}
		return nil
	}

	user := legacyUser
	if user == "" {
		user = `{"mode":"on","ids":[]}`
	}
	group := legacyGroup
	if group == "" {
		group = `{"mode":"on","ids":[]}`
	}
	t = models.FilterTemplate{
		Name:                "default",
		Description:         "默认模板。当下游应用规则的 mode=default 时，回退到此模板的规则。",
		Builtin:             true,
		UserIDRules:         user,
		GroupIDRules:        group,
		MessageRules:        `{"mode":"on"}`,
		PrivateMessageRules: `{"mode":"default"}`,
		GroupMessageRules:   `{"mode":"default"}`,
	}
	return db.Create(&t).Error
}
