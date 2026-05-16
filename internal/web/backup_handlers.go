package web

import (
	"context"
	"strings"
	"time"

	"moebot-next/internal/backup"
	"moebot-next/internal/config"

	"github.com/gofiber/fiber/v2"
)

type backupConfigRequest struct {
	DataDir               *string  `json:"data_dir"`
	TempDir               *string  `json:"temp_dir"`
	ExcludePatterns       []string `json:"exclude_patterns"`
	Endpoint              *string  `json:"endpoint"`
	Region                *string  `json:"region"`
	Bucket                *string  `json:"bucket"`
	Prefix                *string  `json:"prefix"`
	AccessKey             *string  `json:"access_key"`
	SecretKey             *string  `json:"secret_key"`
	SessionToken          *string  `json:"session_token"`
	UseSSL                *bool    `json:"use_ssl"`
	ForcePathStyle        *bool    `json:"force_path_style"`
	ScheduleEnabled       *bool    `json:"schedule_enabled"`
	ScheduleIntervalHours *int     `json:"schedule_interval_hours"`
	ClearAccessKey        bool     `json:"clear_access_key"`
	ClearSecretKey        bool     `json:"clear_secret_key"`
	ClearSessionToken     bool     `json:"clear_session_token"`
}

type backupCreateRequest struct {
	Note string `json:"note"`
}

type backupKeyRequest struct {
	Key     string `json:"key"`
	Confirm string `json:"confirm"`
}

func (s *Server) backupService() *backup.Service {
	if s.Backup == nil {
		s.Backup = backup.New(s.Config, s.DB)
	}
	return s.Backup
}

func (s *Server) handleBackupConfig(c *fiber.Ctx) error {
	return c.JSON(s.backupService().PublicConfig())
}

func (s *Server) handleUpdateBackupConfig(c *fiber.Ctx) error {
	var req backupConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid backup config payload")
	}
	if s.ConfigPath == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "Config path is not configured")
	}

	next := *s.Config
	applyBackupConfigRequest(&next.Backup, req)
	config.NormalizeConfig(&next)
	if err := config.Save(&next, s.ConfigPath); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	*s.Config = next
	s.Backup = backup.New(s.Config, s.DB)
	if s.BackupScheduler != nil {
		s.BackupScheduler.Restart()
	}
	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "备份配置已保存",
		"config":  s.Backup.PublicConfig(),
	})
}

func applyBackupConfigRequest(target *config.BackupConfig, req backupConfigRequest) {
	if target == nil {
		return
	}
	if req.DataDir != nil {
		target.DataDir = strings.TrimSpace(*req.DataDir)
	}
	if req.TempDir != nil {
		target.TempDir = strings.TrimSpace(*req.TempDir)
	}
	if req.ExcludePatterns != nil {
		target.ExcludePatterns = append([]string{}, req.ExcludePatterns...)
	}
	if req.Endpoint != nil {
		target.S3.Endpoint = strings.TrimSpace(*req.Endpoint)
	}
	if req.Region != nil {
		target.S3.Region = strings.TrimSpace(*req.Region)
	}
	if req.Bucket != nil {
		target.S3.Bucket = strings.TrimSpace(*req.Bucket)
	}
	if req.Prefix != nil {
		target.S3.Prefix = strings.Trim(strings.TrimSpace(*req.Prefix), "/")
	}
	if req.UseSSL != nil {
		target.S3.UseSSL = *req.UseSSL
	}
	if req.ForcePathStyle != nil {
		target.S3.ForcePathStyle = *req.ForcePathStyle
	}
	if req.ScheduleEnabled != nil {
		target.Schedule.Enabled = *req.ScheduleEnabled
	}
	if req.ScheduleIntervalHours != nil {
		target.Schedule.IntervalHours = *req.ScheduleIntervalHours
	}
	if req.ClearAccessKey {
		target.S3.AccessKey = ""
	} else if req.AccessKey != nil && strings.TrimSpace(*req.AccessKey) != "" {
		target.S3.AccessKey = strings.TrimSpace(*req.AccessKey)
	}
	if req.ClearSecretKey {
		target.S3.SecretKey = ""
	} else if req.SecretKey != nil && strings.TrimSpace(*req.SecretKey) != "" {
		target.S3.SecretKey = strings.TrimSpace(*req.SecretKey)
	}
	if req.ClearSessionToken {
		target.S3.SessionToken = ""
	} else if req.SessionToken != nil && strings.TrimSpace(*req.SessionToken) != "" {
		target.S3.SessionToken = strings.TrimSpace(*req.SessionToken)
	}
}

func (s *Server) handleBackupTest(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 20*time.Second)
	defer cancel()
	if err := s.backupService().TestConnection(ctx); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "message": "S3 连接测试成功"})
}

func (s *Server) handleListBackups(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	objects, err := s.backupService().List(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{"data": objects, "total": len(objects)})
}

func (s *Server) handleCreateBackup(c *fiber.Ctx) error {
	var req backupCreateRequest
	_ = c.BodyParser(&req)
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Minute)
	defer cancel()
	result, err := s.backupService().Create(ctx, req.Note)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "备份已上传到 S3 兼容存储",
		"result":  result,
	})
}

func (s *Server) handleRestoreBackup(c *fiber.Ctx) error {
	var req backupKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid restore payload")
	}
	if req.Confirm != "RESTORE" {
		return fiber.NewError(fiber.StatusBadRequest, "请确认恢复操作")
	}
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Minute)
	defer cancel()
	result, err := s.backupService().Restore(ctx, req.Key)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "数据目录已恢复，请重启 Moebot 进程/容器后生效",
		"result":  result,
	})
}

func (s *Server) handleDeleteBackup(c *fiber.Ctx) error {
	var req backupKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid delete payload")
	}
	if req.Confirm != "DELETE" {
		return fiber.NewError(fiber.StatusBadRequest, "请确认删除操作")
	}
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Minute)
	defer cancel()
	if err := s.backupService().Delete(ctx, req.Key); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "message": "远端备份已删除"})
}
