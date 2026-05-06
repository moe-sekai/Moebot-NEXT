package renderer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"moebot-next/internal/config"
)

func TestRenderChartURLUsesDefaultOutputWidthAndResvgHeader(t *testing.T) {
	var req ChartRenderRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/render/chart" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("x-render-total-ms", "12")
		w.Header().Set("x-render-resvg-ms", "9")
		w.Header().Set("x-render-size-bytes", "4")
		_, _ = w.Write([]byte("png"))
	}))
	defer server.Close()

	client := New(config.RendererConfig{Host: "127.0.0.1", Port: 3001, Precision: 1.5, ChartPrecision: 4})
	client.baseURL = server.URL

	result, err := client.RenderChartURLWithTrace("https://charts.example.test/1/master.svg", 0)
	if err != nil {
		t.Fatal(err)
	}
	if req.Width != DefaultChartRenderWidth {
		t.Fatalf("width = %d, want %d", req.Width, DefaultChartRenderWidth)
	}
	if req.Precision != 4 {
		t.Fatalf("precision = %v", req.Precision)
	}
	if result.ResvgMS != "9" {
		t.Fatalf("resvg ms = %q", result.ResvgMS)
	}
}

func TestRenderChartURLKeepsExplicitWidth(t *testing.T) {
	var req ChartRenderRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&req)
		_, _ = w.Write([]byte("png"))
	}))
	defer server.Close()

	client := New(config.RendererConfig{Host: "127.0.0.1", Port: 3001, Precision: 1.5, ChartPrecision: 4})
	client.baseURL = server.URL

	if _, err := client.RenderChartURLWithTrace("https://charts.example.test/1/master.svg", 1800); err != nil {
		t.Fatal(err)
	}
	if req.Width != 1800 {
		t.Fatalf("width = %d, want 1800", req.Width)
	}
}
