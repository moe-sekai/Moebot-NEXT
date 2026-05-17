package renderer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func TestEnsureDeckRecommendSnapshotSkipsFreshVersions(t *testing.T) {
	var uploads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deck-recommend/snapshot" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		uploads.Add(1)
		var req DeckRecommendSnapshotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(DeckRecommendSnapshotResponse{
			OK:     true,
			Region: req.Region,
			Master: &DeckRecommendSnapshotMasterStatus{Version: req.Master.Version, KeyCount: len(req.Master.Data)},
		})
	}))
	defer server.Close()

	client := New(config.RendererConfig{Host: "127.0.0.1", Port: 3001})
	client.baseURL = server.URL
	builds := 0
	builder := func() map[string]any {
		builds++
		return map[string]any{"cards": []any{map[string]any{"id": 1}}}
	}

	status, err := client.EnsureDeckRecommendSnapshot("jp", "v1", builder, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !status.MasterUploaded {
		t.Fatalf("first ensure should upload master")
	}
	status, err = client.EnsureDeckRecommendSnapshot("jp", "v1", builder, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status.MasterUploaded {
		t.Fatalf("second ensure should be no-op")
	}
	if got := uploads.Load(); got != 1 {
		t.Fatalf("uploads = %d, want 1", got)
	}
	if builds != 1 {
		t.Fatalf("builder calls = %d, want 1", builds)
	}
}

func TestEnsureDeckRecommendSnapshotCoalescesConcurrentUploads(t *testing.T) {
	var uploads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deck-recommend/snapshot" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		uploads.Add(1)
		time.Sleep(30 * time.Millisecond)
		var req DeckRecommendSnapshotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(DeckRecommendSnapshotResponse{
			OK:     true,
			Region: req.Region,
			Master: &DeckRecommendSnapshotMasterStatus{Version: req.Master.Version, KeyCount: len(req.Master.Data)},
		})
	}))
	defer server.Close()

	client := New(config.RendererConfig{Host: "127.0.0.1", Port: 3001})
	client.baseURL = server.URL
	var builds atomic.Int32
	builder := func() map[string]any {
		builds.Add(1)
		return map[string]any{"cards": []any{map[string]any{"id": 1}}}
	}

	const workers = 8
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	statuses := make(chan DeckRecommendSnapshotEnsureStatus, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := client.EnsureDeckRecommendSnapshot("jp", "v2", builder, "", nil)
			if err != nil {
				errs <- err
				return
			}
			statuses <- status
		}()
	}
	wg.Wait()
	close(errs)
	close(statuses)
	for err := range errs {
		t.Fatal(err)
	}
	if got := uploads.Load(); got != 1 {
		t.Fatalf("uploads = %d, want 1", got)
	}
	if got := builds.Load(); got != 1 {
		t.Fatalf("builder calls = %d, want 1", got)
	}
	uploaded := 0
	shared := 0
	for status := range statuses {
		if status.MasterUploaded {
			uploaded++
		}
		if status.Shared {
			shared++
		}
	}
	if uploaded != 1 {
		t.Fatalf("uploaded statuses = %d, want 1", uploaded)
	}
	if shared == 0 {
		t.Fatalf("expected at least one shared status")
	}
}
