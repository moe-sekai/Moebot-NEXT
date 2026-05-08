package database

import (
	"errors"
	"fmt"

	"moebot-next/internal/models"

	"gorm.io/gorm"
)

// --- FilterGateway: single-row helpers ---

const filterGatewayID = 1

// GetOrCreateFilterGateway loads the singleton gateway record, creating it with
// defaults if it does not yet exist.
func (d *DB) GetOrCreateFilterGateway() (*models.FilterGateway, error) {
	var gw models.FilterGateway
	err := d.First(&gw, filterGatewayID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		gw = models.FilterGateway{
			ID:         filterGatewayID,
			Enabled:    true,
			Host:       "0.0.0.0",
			Port:       3939,
			Suffix:     "/ws",
			BotID:      "10000",
			UserAgent:  "Moebot",
			BufferSize: 4096,
			SleepTime:  5,
		}
		if err := d.Create(&gw).Error; err != nil {
			return nil, fmt.Errorf("create default filter gateway: %w", err)
		}
		return &gw, nil
	}
	if err != nil {
		return nil, err
	}
	return &gw, nil
}

// UpdateFilterGateway writes back the gateway settings. It always writes to ID=1.
func (d *DB) UpdateFilterGateway(gw *models.FilterGateway) error {
	gw.ID = filterGatewayID
	return d.Save(gw).Error
}

// --- FilterApp CRUD ---

// ListFilterApps returns all configured downstream bot applications, ordered.
func (d *DB) ListFilterApps() ([]models.FilterApp, error) {
	var apps []models.FilterApp
	err := d.Order("sort_order ASC, id ASC").Find(&apps).Error
	return apps, err
}

// GetFilterApp returns a single app by id, or ErrRecordNotFound.
func (d *DB) GetFilterApp(id uint) (*models.FilterApp, error) {
	var app models.FilterApp
	if err := d.First(&app, id).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// GetFilterAppByName returns a single app by name.
func (d *DB) GetFilterAppByName(name string) (*models.FilterApp, error) {
	var app models.FilterApp
	if err := d.Where("name = ?", name).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// CreateFilterApp inserts a new bot application row.
func (d *DB) CreateFilterApp(app *models.FilterApp) error {
	return d.Create(app).Error
}

// UpdateFilterApp updates an existing bot application row.
func (d *DB) UpdateFilterApp(app *models.FilterApp) error {
	if app.ID == 0 {
		return errors.New("UpdateFilterApp: id is required")
	}
	return d.Save(app).Error
}

// DeleteFilterApp removes a bot application row by id. Built-in rows cannot be deleted.
func (d *DB) DeleteFilterApp(id uint) error {
	app, err := d.GetFilterApp(id)
	if err != nil {
		return err
	}
	if app.Builtin {
		return errors.New("DeleteFilterApp: built-in app cannot be deleted")
	}
	return d.Delete(&models.FilterApp{}, id).Error
}

// --- FilterTemplate CRUD ---

// DefaultFilterTemplateName is the name reserved for the built-in default template.
const DefaultFilterTemplateName = "default"

// ListFilterTemplates returns all templates ordered by id.
func (d *DB) ListFilterTemplates() ([]models.FilterTemplate, error) {
	var ts []models.FilterTemplate
	err := d.Order("id ASC").Find(&ts).Error
	return ts, err
}

// GetFilterTemplate returns a single template by id.
func (d *DB) GetFilterTemplate(id uint) (*models.FilterTemplate, error) {
	var t models.FilterTemplate
	if err := d.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// GetFilterTemplateByName returns a single template by name.
func (d *DB) GetFilterTemplateByName(name string) (*models.FilterTemplate, error) {
	var t models.FilterTemplate
	if err := d.Where("name = ?", name).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// GetDefaultFilterTemplate returns the built-in default template (creating it if missing).
func (d *DB) GetDefaultFilterTemplate() (*models.FilterTemplate, error) {
	t, err := d.GetFilterTemplateByName(DefaultFilterTemplateName)
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	seed := &models.FilterTemplate{
		Name:                DefaultFilterTemplateName,
		Description:         "默认模板。当下游应用规则的 mode=default 时，回退到此模板的规则。",
		Builtin:             true,
		UserIDRules:         `{"mode":"on","ids":[]}`,
		GroupIDRules:        `{"mode":"on","ids":[]}`,
		MessageRules:        `{"mode":"on"}`,
		PrivateMessageRules: `{"mode":"default"}`,
		GroupMessageRules:   `{"mode":"default"}`,
	}
	if err := d.Create(seed).Error; err != nil {
		return nil, fmt.Errorf("create default filter template: %w", err)
	}
	return seed, nil
}

// CreateFilterTemplate inserts a new template row.
func (d *DB) CreateFilterTemplate(t *models.FilterTemplate) error {
	return d.Create(t).Error
}

// UpdateFilterTemplate updates an existing template row.
func (d *DB) UpdateFilterTemplate(t *models.FilterTemplate) error {
	if t.ID == 0 {
		return errors.New("UpdateFilterTemplate: id is required")
	}
	return d.Save(t).Error
}

// CountFilterAppsByTemplate returns how many FilterApp rows reference the given template.
func (d *DB) CountFilterAppsByTemplate(id uint) (int64, error) {
	var n int64
	err := d.Model(&models.FilterApp{}).Where("template_id = ?", id).Count(&n).Error
	return n, err
}

// DeleteFilterTemplate removes a template by id. Built-in templates and templates
// referenced by any FilterApp cannot be deleted.
func (d *DB) DeleteFilterTemplate(id uint) error {
	t, err := d.GetFilterTemplate(id)
	if err != nil {
		return err
	}
	if t.Builtin {
		return errors.New("DeleteFilterTemplate: built-in template cannot be deleted")
	}
	n, err := d.CountFilterAppsByTemplate(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("DeleteFilterTemplate: template is referenced by %d app(s)", n)
	}
	return d.Delete(&models.FilterTemplate{}, id).Error
}
