package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shortener/internal/domain"
	"shortener/internal/storage"
	"shortener/internal/validator"
)

// Mock implementations
type mockURLStorage struct {
	urls map[string]string // shortCode -> URL
	err  error
}

func (m *mockURLStorage) SaveURL(ctx context.Context, url string, withAlias bool, alias string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	
	var shortCode string
	if withAlias {
		if _, exists := m.urls[alias]; exists {
			return "", errors.New("alias already exists")
		}
		shortCode = alias
	} else {
		shortCode = "abc123" // Fixed for testing
	}
	
	m.urls[shortCode] = url
	return shortCode, nil
}

func (m *mockURLStorage) GetURL(ctx context.Context, shortCode string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if url, ok := m.urls[shortCode]; ok {
		return url, nil
	}
	return "", storage.ErrNotFound
}

type mockClickStorage struct {
	clicks map[string][]domain.Click
	err    error
}

func (m *mockClickStorage) SaveClick(ctx context.Context, click domain.Click) error {
	if m.err != nil {
		return m.err
	}
	m.clicks[click.ShortCode] = append(m.clicks[click.ShortCode], click)
	return nil
}

func (m *mockClickStorage) Analytics(ctx context.Context, shortCode string, from, to *time.Time, topLimit int) (domain.AnalyticsResponse, error) {
	if m.err != nil {
		return domain.AnalyticsResponse{}, m.err
	}
	
	clicks, exists := m.clicks[shortCode]
	if !exists {
		return domain.AnalyticsResponse{}, storage.ErrNotFound
	}
	
	return domain.AnalyticsResponse{
		ShortCode:   shortCode,
		TotalClicks: int64(len(clicks)),
	}, nil
}

// Implement other required methods...
func (m *mockClickStorage) GetClicks(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error) { return nil, nil }
func (m *mockClickStorage) ClickCount(ctx context.Context, shortCode string) (int64, error) { return 0, nil }
func (m *mockClickStorage) UniqueClickCount(ctx context.Context, shortCode string) (int64, error) { return 0, nil }
func (m *mockClickStorage) ClicksByDay(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error) { return nil, nil }
func (m *mockClickStorage) ClicksByMonth(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error) { return nil, nil }
func (m *mockClickStorage) ClicksByUserAgent(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error) { return nil, nil }
func (m *mockClickStorage) ClicksByReferer(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error) { return nil, nil }
func (m *mockClickStorage) ClicksByIP(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error) { return nil, nil }

func newTestServer() (*Server, *mockURLStorage, *mockClickStorage) {
	urlStorage := &mockURLStorage{urls: make(map[string]string)}
	clickStorage := &mockClickStorage{clicks: make(map[string][]domain.Click)}
	validator := validator.New()
	
	server := New(urlStorage, clickStorage, *validator, nil)
	server.RegisterRoutes(context.Background())
	
	return server, urlStorage, clickStorage
}

// BASIC TESTS

func TestCreateLink_Success(t *testing.T) {
	server, _, _ := newTestServer()
	
	reqBody := domain.ShortenRequest{URL: "https://example.com"}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	
	if resp["short_code"] == "" {
		t.Fatalf("expected short_code in response, got: %v", resp)
	}
}

func TestRedirectLink_Success(t *testing.T) {
	server, urlStorage, _ := newTestServer()
	
	// Pre-populate storage
	urlStorage.urls["abc123"] = "https://example.com"
	
	req := httptest.NewRequest("GET", "/s/abc123", nil)
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}
	
	location := w.Header().Get("Location")
	if location != "https://example.com" {
		t.Fatalf("expected redirect to https://example.com, got %s", location)
	}
}

func TestRedirectLink_NotFound(t *testing.T) {
	server, _, _ := newTestServer()
	
	req := httptest.NewRequest("GET", "/s/missing", nil)
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAnalytics_Success(t *testing.T) {
	server, _, clickStorage := newTestServer()
	
	// Pre-populate with clicks
	clickStorage.clicks["abc123"] = []domain.Click{
		{ShortCode: "abc123", IP: "127.0.0.1"},
		{ShortCode: "abc123", IP: "127.0.0.2"},
	}
	
	req := httptest.NewRequest("GET", "/analytics/abc123", nil)
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	
	var resp domain.AnalyticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	
	if resp.TotalClicks != 2 {
		t.Fatalf("expected 2 clicks, got %d", resp.TotalClicks)
	}
}

func TestAnalytics_NotFound(t *testing.T) {
	server, _, _ := newTestServer()
	
	req := httptest.NewRequest("GET", "/analytics/missing", nil)
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// ADVANCED TESTS

func TestCreateLink_CustomAlias_Success(t *testing.T) {
	server, _, _ := newTestServer()
	
	reqBody := domain.ShortenRequest{
		URL:   "https://example.com",
		Alias: "my-custom-alias",
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	
	if resp["short_code"] != "my-custom-alias" {
		t.Fatalf("expected 'my-custom-alias', got %s", resp["short_code"])
	}
}

func TestCreateLink_CustomAlias_AlreadyExists(t *testing.T) {
	server, urlStorage, _ := newTestServer()
	
	// Pre-populate with existing alias
	urlStorage.urls["existing"] = "https://other.com"
	
	reqBody := domain.ShortenRequest{
		URL:   "https://example.com",
		Alias: "existing",
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestCreateLink_CustomAlias_ConflictWithGenerated(t *testing.T) {
	server, urlStorage, _ := newTestServer()
	
	// Pre-populate with a generated code that matches our alias
	urlStorage.urls["abc123"] = "https://other.com"
	
	reqBody := domain.ShortenRequest{
		URL:   "https://example.com",
		Alias: "abc123", // Conflicts with existing generated code
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	server.g.ServeHTTP(w, req)
	
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}