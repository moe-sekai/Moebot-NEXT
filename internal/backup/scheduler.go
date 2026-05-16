package backup

import (
	"context"
	"sync"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/database"

	"github.com/rs/zerolog/log"
)

// Scheduler runs automatic periodic data backups according to the current config.
type Scheduler struct {
	mu     sync.Mutex
	cfg    *config.Config
	db     *database.DB
	cancel context.CancelFunc
}

// NewScheduler creates a scheduler bound to the shared mutable core config.
func NewScheduler(cfg *config.Config, db *database.DB) *Scheduler {
	return &Scheduler{cfg: cfg, db: db}
}

// Start starts or restarts the periodic backup loop using the latest config.
func (s *Scheduler) Start() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()

	cfg := s.effectiveConfigLocked()
	if !cfg.Schedule.Enabled {
		return
	}
	interval := time.Duration(cfg.Schedule.IntervalHours) * time.Hour
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.run(ctx, interval)
	log.Info().Int("interval_hours", int(interval/time.Hour)).Msg("Automatic backup scheduler started")
}

// Stop stops the periodic backup loop.
func (s *Scheduler) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
}

// Restart reloads the latest config and starts/stops the loop accordingly.
func (s *Scheduler) Restart() {
	if s == nil {
		return
	}
	s.Start()
}

func (s *Scheduler) stopLocked() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

func (s *Scheduler) effectiveConfigLocked() config.BackupConfig {
	cfg := config.DefaultConfig().Backup
	if s.cfg != nil {
		cfg = s.cfg.Backup
	}
	if cfg.Schedule.IntervalHours <= 0 {
		cfg.Schedule.IntervalHours = 24
	}
	return cfg
}

func (s *Scheduler) run(ctx context.Context, interval time.Duration) {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	running := false
	var runningMu sync.Mutex
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			runningMu.Lock()
			if running {
				runningMu.Unlock()
				log.Warn().Msg("Skipping scheduled backup because previous backup is still running")
				timer.Reset(interval)
				continue
			}
			running = true
			runningMu.Unlock()

			go func() {
				defer func() {
					runningMu.Lock()
					running = false
					runningMu.Unlock()
				}()
				backupCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
				defer cancel()
				started := time.Now()
				result, err := New(s.cfg, s.db).Create(backupCtx, "scheduled")
				if err != nil {
					log.Error().Err(err).Msg("Scheduled backup failed")
					return
				}
				log.Info().
					Str("key", result.Object.Key).
					Int64("size", result.Object.Size).
					Dur("duration", time.Since(started)).
					Msg("Scheduled backup completed")
			}()
			timer.Reset(interval)
		}
	}
}
