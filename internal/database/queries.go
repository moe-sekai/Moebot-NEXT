package database

import (
	"fmt"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/models"
)

// --- User Queries ---

// GetUserByPlatform finds the first user binding for their platform ID.
// If the user has explicitly set a default region via UserDefaultRegion, that
// region's binding is returned first; otherwise it falls back to JP-first
// then most-recently-updated ordering.
func (d *DB) GetUserByPlatform(platform, platformID string) (*models.User, error) {
	if region, _ := d.GetUserDefaultRegion(platform, platformID); region != "" {
		if user, err := d.GetUserByPlatformRegion(platform, platformID, region); err == nil {
			return user, nil
		}
	}
	var user models.User
	err := d.Where("platform = ? AND platform_id = ?", platform, platformID).Order("server_region = 'jp' DESC, updated_at DESC").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserDefaultRegion returns the user's explicitly set default region, or
// an empty string when none is configured.
func (d *DB) GetUserDefaultRegion(platform, platformID string) (string, error) {
	var row models.UserDefaultRegion
	err := d.Where("platform = ? AND platform_id = ?", platform, platformID).First(&row).Error
	if err != nil {
		return "", err
	}
	return config.NormalizeRegion(row.ServerRegion), nil
}

// SetUserDefaultRegion upserts the default region for a platform user.
// The region is normalized; an empty / invalid region returns an error.
func (d *DB) SetUserDefaultRegion(platform, platformID, region string) error {
	region = config.NormalizeRegion(region)
	if region == "" || !config.IsValidRegion(region) {
		return fmt.Errorf("invalid region: %s", region)
	}
	var row models.UserDefaultRegion
	err := d.Where("platform = ? AND platform_id = ?", platform, platformID).First(&row).Error
	if err != nil {
		row = models.UserDefaultRegion{Platform: platform, PlatformID: platformID, ServerRegion: region}
		return d.Create(&row).Error
	}
	row.ServerRegion = region
	return d.Save(&row).Error
}

// GetUserByPlatformRegion finds a user binding for one game server.
func (d *DB) GetUserByPlatformRegion(platform, platformID, region string) (*models.User, error) {
	var user models.User
	region = config.NormalizeRegion(region)
	if region == "" {
		region = config.RegionJP
	}
	err := d.Where("platform = ? AND platform_id = ? AND server_region = ?", platform, platformID, region).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpsertUser creates or updates a user binding.
func (d *DB) UpsertUser(user *models.User) error {
	if user.ServerRegion == "" {
		user.ServerRegion = config.RegionJP
	}
	return d.Save(user).Error
}

// DeleteUser removes a user by ID.
func (d *DB) DeleteUser(id uint) error {
	return d.Delete(&models.User{}, id).Error
}

// ListUsers returns all users with pagination.
func (d *DB) ListUsers(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	d.Model(&models.User{}).Count(&total)
	err := d.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

// --- Suite Setting Queries ---

func (d *DB) GetSuiteSetting(platform, platformID, region string) (*models.SuiteSetting, error) {
	var setting models.SuiteSetting
	region = config.NormalizeRegion(region)
	if region == "" {
		region = config.RegionJP
	}
	err := d.Where("platform = ? AND platform_id = ? AND server_region = ?", platform, platformID, region).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (d *DB) UpsertSuiteSetting(setting *models.SuiteSetting) error {
	if setting.ServerRegion == "" {
		setting.ServerRegion = config.RegionJP
	}
	setting.ServerRegion = config.NormalizeRegion(setting.ServerRegion)
	setting.Mode = config.NormalizeSuiteMode(setting.Mode)
	return d.Save(setting).Error
}

// --- Group Queries ---

// GetGroup finds a group by platform and group ID.
func (d *DB) GetGroup(platform, groupID string) (*models.Group, error) {
	var group models.Group
	err := d.Where("platform = ? AND group_id = ?", platform, groupID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// UpsertGroup creates or updates a group configuration.
func (d *DB) UpsertGroup(group *models.Group) error {
	return d.Save(group).Error
}

// ListGroups returns all groups with pagination.
func (d *DB) ListGroups(offset, limit int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64
	d.Model(&models.Group{}).Count(&total)
	err := d.Offset(offset).Limit(limit).Order("created_at DESC").Find(&groups).Error
	return groups, total, err
}

// GetGroupByID finds a group by primary key.
func (d *DB) GetGroupByID(id uint) (*models.Group, error) {
	var group models.Group
	if err := d.First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// DeleteGroup removes a group by primary key.
func (d *DB) DeleteGroup(id uint) error {
	return d.Delete(&models.Group{}, id).Error
}

// GroupCommandStat aggregates command usage for a single group within a window.
type GroupCommandStat struct {
	Platform string    `json:"platform"`
	GroupID  string    `json:"group_id"`
	Count    int64     `json:"count"`
	LastUsed time.Time `json:"last_used"`
	AvgMs    float64   `json:"avg_ms"`
}

// GetGroupCommandStats returns per-group aggregates within the time window.
func (d *DB) GetGroupCommandStats(since time.Time) ([]GroupCommandStat, error) {
	var results []GroupCommandStat
	err := d.Model(&models.CommandStat{}).
		Select("platform, group_id, COUNT(*) as count, MAX(created_at) as last_used, COALESCE(AVG(response_ms), 0) as avg_ms").
		Where("created_at > ? AND group_id <> ''", since).
		Group("platform, group_id").
		Find(&results).Error
	return results, err
}

// ListGroupRecentCommands returns recent command invocations from one group.
func (d *DB) ListGroupRecentCommands(platform, groupID string, limit int) ([]models.CommandStat, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []models.CommandStat
	err := d.Where("platform = ? AND group_id = ?", platform, groupID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// --- Command Stats Queries ---

// RecordCommandStat inserts a new command usage record.
func (d *DB) RecordCommandStat(stat *models.CommandStat) error {
	return d.Create(stat).Error
}

// CommandStatsSummary holds aggregated command usage data.
type CommandStatsSummary struct {
	Command string  `json:"command"`
	Count   int64   `json:"count"`
	AvgMs   float64 `json:"avg_ms"`
}

// GetCommandStats returns aggregated command usage statistics.
func (d *DB) GetCommandStats(since time.Time) ([]CommandStatsSummary, error) {
	var results []CommandStatsSummary
	err := d.Model(&models.CommandStat{}).
		Select("command, COUNT(*) as count, AVG(response_ms) as avg_ms").
		Where("created_at > ?", since).
		Group("command").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// CommandStatsTrendPoint is one data point in the daily trend.
type CommandStatsTrendPoint struct {
	Date  string  `json:"date"`
	Count int64   `json:"count"`
	AvgMs float64 `json:"avg_ms"`
}

// GetCommandStatsTrend returns daily call counts since a cutoff time.
func (d *DB) GetCommandStatsTrend(since time.Time) ([]CommandStatsTrendPoint, error) {
	var results []CommandStatsTrendPoint
	err := d.Model(&models.CommandStat{}).
		Select("strftime('%Y-%m-%d', created_at) as date, COUNT(*) as count, AVG(response_ms) as avg_ms").
		Where("created_at > ?", since).
		Group("date").
		Order("date ASC").
		Find(&results).Error
	return results, err
}

// CommandStatsTotals summarises usage during a window.
type CommandStatsTotals struct {
	Calls          int64   `json:"calls"`
	DistinctUsers  int64   `json:"users"`
	DistinctGroups int64   `json:"groups"`
	AvgMs          float64 `json:"avg_ms"`
}

// GetCommandStatsTotals returns aggregate counters for the time window.
func (d *DB) GetCommandStatsTotals(since time.Time) (CommandStatsTotals, error) {
	var totals CommandStatsTotals
	err := d.Model(&models.CommandStat{}).
		Select("COUNT(*) as calls, COUNT(DISTINCT user_id) as distinct_users, COUNT(DISTINCT group_id) as distinct_groups, COALESCE(AVG(response_ms), 0) as avg_ms").
		Where("created_at > ?", since).
		Scan(&totals).Error
	return totals, err
}

// CommandStatsPlatformPoint counts calls per platform.
type CommandStatsPlatformPoint struct {
	Platform string `json:"platform"`
	Count    int64  `json:"count"`
}

// GetCommandStatsByPlatform returns per-platform call counts since a cutoff.
func (d *DB) GetCommandStatsByPlatform(since time.Time) ([]CommandStatsPlatformPoint, error) {
	var results []CommandStatsPlatformPoint
	err := d.Model(&models.CommandStat{}).
		Select("COALESCE(NULLIF(platform, ''), 'unknown') as platform, COUNT(*) as count").
		Where("created_at > ?", since).
		Group("platform").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// ListRecentCommands returns the latest command invocation records.
func (d *DB) ListRecentCommands(limit int) ([]models.CommandStat, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var commands []models.CommandStat
	err := d.Order("created_at DESC").Limit(limit).Find(&commands).Error
	return commands, err
}

// GetTotalStats returns total command count and unique user count.
func (d *DB) GetTotalStats() (commandCount int64, userCount int64, groupCount int64) {
	d.Model(&models.CommandStat{}).Count(&commandCount)
	d.Model(&models.User{}).Count(&userCount)
	d.Model(&models.Group{}).Count(&groupCount)
	return
}

// Ping checks whether the underlying database connection is alive.
func (d *DB) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}
