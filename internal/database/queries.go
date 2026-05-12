package database

import (
	"fmt"
	"strings"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// GetGroup finds a group by platform, OneBot client and group ID.
func (d *DB) GetGroup(platform, clientID, groupID string) (*models.Group, error) {
	var group models.Group
	err := d.Where("platform = ? AND client_id = ? AND group_id = ?", normalizePlatform(platform), normalizeClientID(clientID), normalizeGroupID(groupID)).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// EnsureGroup creates the group row if it does not already exist. It is called
// from command recording so /groups can be populated from real group traffic.
func (d *DB) EnsureGroup(platform, clientID, groupID, name string) error {
	platform = normalizePlatform(platform)
	clientID = normalizeClientID(clientID)
	groupID = normalizeGroupID(groupID)
	name = strings.TrimSpace(name)
	if groupID == "" || groupID == "0" {
		return nil
	}

	group := models.Group{
		Platform:  platform,
		ClientID:  clientID,
		GroupID:   groupID,
		Name:      name,
		Enabled:   true,
		Config:    "{}",
		CreatedAt: time.Now(),
	}
	if err := d.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "platform"},
			{Name: "client_id"},
			{Name: "group_id"},
		},
		DoNothing: true,
	}).Create(&group).Error; err != nil {
		return err
	}
	if name == "" {
		return nil
	}
	return d.Model(&models.Group{}).
		Where("platform = ? AND client_id = ? AND group_id = ? AND (name = '' OR name IS NULL)", platform, clientID, groupID).
		Update("name", name).Error
}

// UpsertGroup creates or updates a group configuration.
func (d *DB) UpsertGroup(group *models.Group) error {
	if group == nil {
		return fmt.Errorf("group is nil")
	}
	group.Platform = normalizePlatform(group.Platform)
	group.ClientID = normalizeClientID(group.ClientID)
	group.GroupID = normalizeGroupID(group.GroupID)
	if strings.TrimSpace(group.Config) == "" {
		group.Config = "{}"
	}
	return d.Save(group).Error
}

// ListGroups returns all groups with pagination. If clientID is non-nil, it
// filters by that OneBot self_id; an empty string means the legacy/unknown
// client bucket.
func (d *DB) ListGroups(offset, limit int, clientID *string) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64
	q := applyClientFilter(d.Model(&models.Group{}), clientID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Offset(offset).Limit(limit).Order("created_at DESC, id DESC").Find(&groups).Error
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

// GroupClientSummary describes one client bucket in the groups table.
type GroupClientSummary struct {
	ClientID string `json:"client_id"`
	Count    int64  `json:"count"`
}

// ListGroupClients returns known OneBot client buckets from registered groups.
func (d *DB) ListGroupClients() ([]GroupClientSummary, error) {
	var results []GroupClientSummary
	err := d.Model(&models.Group{}).
		Select("COALESCE(client_id, '') as client_id, COUNT(*) as count").
		Group("client_id").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// GroupCommandStat aggregates command usage for a single group within a window.
type GroupCommandStat struct {
	Platform string    `json:"platform"`
	ClientID string    `json:"client_id"`
	GroupID  string    `json:"group_id"`
	Count    int64     `json:"count"`
	LastUsed time.Time `json:"last_used"`
	AvgMs    float64   `json:"avg_ms"`
}

// GetGroupCommandStats returns per-group aggregates within the time window.
func (d *DB) GetGroupCommandStats(since time.Time, clientID *string) ([]GroupCommandStat, error) {
	var results []GroupCommandStat
	q := applyClientFilter(d.Model(&models.CommandStat{}), clientID).
		Select("COALESCE(NULLIF(platform, ''), 'unknown') as platform, COALESCE(client_id, '') as client_id, group_id, COUNT(*) as count, MAX(created_at) as last_used, COALESCE(AVG(response_ms), 0) as avg_ms").
		Where("created_at > ? AND group_id <> '' AND group_id <> '0'", since).
		Group("platform, client_id, group_id")
	err := q.Find(&results).Error
	return results, err
}

// ListGroupRecentCommands returns recent command invocations from one group.
func (d *DB) ListGroupRecentCommands(platform, clientID, groupID string, limit int) ([]models.CommandStat, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []models.CommandStat
	err := d.Where("platform = ? AND client_id = ? AND group_id = ?", normalizePlatform(platform), normalizeClientID(clientID), normalizeGroupID(groupID)).
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
func (d *DB) GetCommandStats(since time.Time, clientID *string) ([]CommandStatsSummary, error) {
	var results []CommandStatsSummary
	err := commandStatsBase(d.DB, since, clientID).
		Select("command, COUNT(*) as count, AVG(response_ms) as avg_ms").
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
func (d *DB) GetCommandStatsTrend(since time.Time, clientID *string) ([]CommandStatsTrendPoint, error) {
	var results []CommandStatsTrendPoint
	err := commandStatsBase(d.DB, since, clientID).
		Select("strftime('%Y-%m-%d', created_at) as date, COUNT(*) as count, AVG(response_ms) as avg_ms").
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
func (d *DB) GetCommandStatsTotals(since time.Time, clientID *string) (CommandStatsTotals, error) {
	var totals CommandStatsTotals
	err := commandStatsBase(d.DB, since, clientID).
		Select("COUNT(*) as calls, COUNT(DISTINCT COALESCE(client_id, '') || '|' || COALESCE(NULLIF(user_id, ''), 'unknown')) as distinct_users, COUNT(DISTINCT COALESCE(client_id, '') || '|' || COALESCE(NULLIF(group_id, ''), 'private')) as distinct_groups, COALESCE(AVG(response_ms), 0) as avg_ms").
		Scan(&totals).Error
	return totals, err
}

// CommandStatsPlatformPoint counts calls per platform.
type CommandStatsPlatformPoint struct {
	Platform string `json:"platform"`
	Count    int64  `json:"count"`
}

// GetCommandStatsByPlatform returns per-platform call counts since a cutoff.
func (d *DB) GetCommandStatsByPlatform(since time.Time, clientID *string) ([]CommandStatsPlatformPoint, error) {
	var results []CommandStatsPlatformPoint
	err := commandStatsBase(d.DB, since, clientID).
		Select("COALESCE(NULLIF(platform, ''), 'unknown') as platform, COUNT(*) as count").
		Group("platform").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// CommandStatsClientPoint counts calls per OneBot client account.
type CommandStatsClientPoint struct {
	ClientID string `json:"client_id"`
	Count    int64  `json:"count"`
}

// GetCommandStatsByClient returns per-client call counts since a cutoff.
func (d *DB) GetCommandStatsByClient(since time.Time, clientID *string) ([]CommandStatsClientPoint, error) {
	var results []CommandStatsClientPoint
	err := commandStatsBase(d.DB, since, clientID).
		Select("COALESCE(client_id, '') as client_id, COUNT(*) as count").
		Group("client_id").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// CommandStatsGroupPoint counts calls per group.
type CommandStatsGroupPoint struct {
	Platform string `json:"platform"`
	ClientID string `json:"client_id"`
	GroupID  string `json:"group_id"`
	Count    int64  `json:"count"`
}

// GetCommandStatsByGroup returns top groups by call count since a cutoff.
// Empty group IDs (private chat) are aggregated under the "private" bucket.
func (d *DB) GetCommandStatsByGroup(since time.Time, clientID *string, limit int) ([]CommandStatsGroupPoint, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var results []CommandStatsGroupPoint
	err := commandStatsBase(d.DB, since, clientID).
		Select("COALESCE(NULLIF(platform, ''), 'unknown') as platform, COALESCE(client_id, '') as client_id, COALESCE(NULLIF(group_id, ''), 'private') as group_id, COUNT(*) as count").
		Group("platform, client_id, group_id").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// CommandStatsUserPoint counts calls per user.
type CommandStatsUserPoint struct {
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id"`
	Count    int64  `json:"count"`
}

// GetCommandStatsByUser returns top users by call count since a cutoff.
func (d *DB) GetCommandStatsByUser(since time.Time, clientID *string, limit int) ([]CommandStatsUserPoint, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var results []CommandStatsUserPoint
	err := commandStatsBase(d.DB, since, clientID).
		Select("COALESCE(client_id, '') as client_id, COALESCE(NULLIF(user_id, ''), 'unknown') as user_id, COUNT(*) as count").
		Group("client_id, user_id").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// ListRecentCommands returns the latest command invocation records.
func (d *DB) ListRecentCommands(limit int, clientID *string) ([]models.CommandStat, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var commands []models.CommandStat
	err := applyClientFilter(d.Model(&models.CommandStat{}), clientID).
		Order("created_at DESC").
		Limit(limit).
		Find(&commands).Error
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

func commandStatsBase(db *gorm.DB, since time.Time, clientID *string) *gorm.DB {
	return applyClientFilter(db.Model(&models.CommandStat{}).Where("created_at > ?", since), clientID)
}

func applyClientFilter(q *gorm.DB, clientID *string) *gorm.DB {
	if clientID == nil {
		return q
	}
	return q.Where("client_id = ?", normalizeClientID(*clientID))
}

func normalizePlatform(platform string) string {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return "unknown"
	}
	return platform
}

func normalizeClientID(clientID string) string {
	return strings.TrimSpace(clientID)
}

func normalizeGroupID(groupID string) string {
	return strings.TrimSpace(groupID)
}
