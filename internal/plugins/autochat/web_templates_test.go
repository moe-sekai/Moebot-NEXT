package autochat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
)

func TestTemplateUpsertKeepsDistinctNamesAfterReload(t *testing.T) {
	oldCfg := GetConfig()
	defer setConfig(oldCfg)

	cfg := &Config{}
	applyDefaults(cfg)
	setConfig(cfg)

	p := &pluginImpl{configPath: t.TempDir() + "/autochat.yml"}
	app := fiber.New()
	p.registerWebRoutes(app.Group("/api"))

	putTemplate := func(name, persona string) {
		t.Helper()
		body, err := json.Marshal(templatePayload{
			Name:     name,
			Persona:  persona,
			Models:   []string{"openai:gpt-4o-mini"},
			Keywords: []string{name + "-keyword"},
		})
		if err != nil {
			t.Fatalf("marshal template %s: %v", name, err)
		}
		req := httptest.NewRequest(http.MethodPut, "/api/plugins/autochat/templates/"+name, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("PUT template %s: %v", name, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("PUT template %s status = %d, want %d", name, resp.StatusCode, http.StatusOK)
		}
	}

	putTemplate("mnr", "persona-mnr")
	putTemplate("mzk", "persona-mzk")

	req := httptest.NewRequest(http.MethodGet, "/api/plugins/autochat/templates", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("GET templates: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET templates status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var listed struct {
		Templates []templatePayload `json:"templates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listed); err != nil {
		t.Fatalf("decode templates response: %v", err)
	}
	if len(listed.Templates) != 2 {
		t.Fatalf("templates len = %d, want 2: %#v", len(listed.Templates), listed.Templates)
	}
	gotPersona := map[string]string{}
	for _, tpl := range listed.Templates {
		gotPersona[tpl.Name] = tpl.Persona
	}
	if gotPersona["mnr"] != "persona-mnr" {
		t.Fatalf("mnr persona = %q, want persona-mnr (all=%#v)", gotPersona["mnr"], gotPersona)
	}
	if gotPersona["mzk"] != "persona-mzk" {
		t.Fatalf("mzk persona = %q, want persona-mzk (all=%#v)", gotPersona["mzk"], gotPersona)
	}

	var diskCfg Config
	if err := plugin.ReadYAMLInto(p.configPath, &diskCfg); err != nil {
		t.Fatalf("read yaml: %v", err)
	}
	if len(diskCfg.Chat.Templates) != 2 {
		t.Fatalf("disk templates len = %d, want 2: %#v", len(diskCfg.Chat.Templates), diskCfg.Chat.Templates)
	}
	if diskCfg.Chat.Templates["mnr"].Persona != "persona-mnr" {
		t.Fatalf("disk mnr persona = %q, want persona-mnr", diskCfg.Chat.Templates["mnr"].Persona)
	}
	if diskCfg.Chat.Templates["mzk"].Persona != "persona-mzk" {
		t.Fatalf("disk mzk persona = %q, want persona-mzk", diskCfg.Chat.Templates["mzk"].Persona)
	}
}

func TestTemplateUpsertRejectsBodyNameMismatch(t *testing.T) {
	oldCfg := GetConfig()
	defer setConfig(oldCfg)

	cfg := &Config{}
	applyDefaults(cfg)
	setConfig(cfg)

	p := &pluginImpl{configPath: t.TempDir() + "/autochat.yml"}
	app := fiber.New()
	p.registerWebRoutes(app.Group("/api"))

	body, err := json.Marshal(templatePayload{Name: "mzk", Persona: "wrong"})
	if err != nil {
		t.Fatalf("marshal template: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/plugins/autochat/templates/mnr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("PUT mismatched template: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("mismatched PUT status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}
