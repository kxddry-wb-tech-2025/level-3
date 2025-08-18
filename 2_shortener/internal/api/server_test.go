package api

import (
	"context"
	"shortener/internal/domain"
	"shortener/internal/validator"
	"testing"
	"time"
)

// MockURLStorage is a mock implementation of URLStorage interface
type MockURLStorage struct {
	saveURLFunc func(ctx context.Context, url string, withAlias bool, alias string) (string, error)
	getURLFunc  func(ctx context.Context, shortCode string) (string, error)
}

func (m *MockURLStorage) SaveURL(ctx context.Context, url string, withAlias bool, alias string) (string, error) {
	if m.saveURLFunc != nil {
		return m.saveURLFunc(ctx, url, withAlias, alias)
	}
	return "", nil
}

func (m *MockURLStorage) GetURL(ctx context.Context, shortCode string) (string, error) {
	if m.getURLFunc != nil {
		return m.getURLFunc(ctx, shortCode)
	}
	return "", nil
}

// MockClickStorage is a mock implementation of ClickStorage interface
type MockClickStorage struct {
	saveClickFunc           func(ctx context.Context, click domain.Click) error
	getClicksFunc           func(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error)
	clickCountFunc          func(ctx context.Context, shortCode string) (int64, error)
	uniqueClickCountFunc    func(ctx context.Context, shortCode string) (int64, error)
	clicksByDayFunc         func(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error)
	clicksByMonthFunc       func(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error)
	clicksByUserAgentFunc   func(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error)
	analyticsFunc           func(ctx context.Context, shortCode string, from, to *time.Time, topLimit int) (domain.AnalyticsResponse, error)
	clicksByRefererFunc     func(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error)
	clicksByIPFunc          func(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error)
}

func (m *MockClickStorage) SaveClick(ctx context.Context, click domain.Click) error {
	if m.saveClickFunc != nil {
		return m.saveClickFunc(ctx, click)
	}
	return nil
}

func (m *MockClickStorage) GetClicks(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error) {
	if m.getClicksFunc != nil {
		return m.getClicksFunc(ctx, shortCode, limit, offset)
	}
	return nil, nil
}

func (m *MockClickStorage) ClickCount(ctx context.Context, shortCode string) (int64, error) {
	if m.clickCountFunc != nil {
		return m.clickCountFunc(ctx, shortCode)
	}
	return 0, nil
}

func (m *MockClickStorage) UniqueClickCount(ctx context.Context, shortCode string) (int64, error) {
	if m.uniqueClickCountFunc != nil {
		return m.uniqueClickCountFunc(ctx, shortCode)
	}
	return 0, nil
}

func (m *MockClickStorage) ClicksByDay(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error) {
	if m.clicksByDayFunc != nil {
		return m.clicksByDayFunc(ctx, shortCode, start, end)
	}
	return nil, nil
}

func (m *MockClickStorage) ClicksByMonth(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error) {
	if m.clicksByMonthFunc != nil {
		return m.clicksByMonthFunc(ctx, shortCode, start, end)
	}
	return nil, nil
}

func (m *MockClickStorage) ClicksByUserAgent(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error) {
	if m.clicksByUserAgentFunc != nil {
		return m.clicksByUserAgentFunc(ctx, shortCode, start, end, limit)
	}
	return nil, nil
}

func (m *MockClickStorage) Analytics(ctx context.Context, shortCode string, from, to *time.Time, topLimit int) (domain.AnalyticsResponse, error) {
	if m.analyticsFunc != nil {
		return m.analyticsFunc(ctx, shortCode, from, to, topLimit)
	}
	return domain.AnalyticsResponse{}, nil
}

func (m *MockClickStorage) ClicksByReferer(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error) {
	if m.clicksByRefererFunc != nil {
		return m.clicksByRefererFunc(ctx, shortCode, start, end, limit)
	}
	return nil, nil
}

func (m *MockClickStorage) ClicksByIP(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error) {
	if m.clicksByIPFunc != nil {
		return m.clicksByIPFunc(ctx, shortCode, start, end, limit)
	}
	return nil, nil
}

// MockCacheStorage is a mock implementation of CacheStorage interface
type MockCacheStorage struct {
	getURLFunc func(ctx context.Context, shortCode string) (string, error)
	setURLFunc func(ctx context.Context, shortCode, url string, usage int64) error
}

func (m *MockCacheStorage) GetURL(ctx context.Context, shortCode string) (string, error) {
	if m.getURLFunc != nil {
		return m.getURLFunc(ctx, shortCode)
	}
	return "", nil
}

func (m *MockCacheStorage) SetURL(ctx context.Context, shortCode, url string, usage int64) error {
	if m.setURLFunc != nil {
		return m.setURLFunc(ctx, shortCode, url, usage)
	}
	return nil
}

// MockValidator is a mock implementation of validator interface
type MockValidator struct {
	urlFunc    func(url string) error
	structFunc func(s interface{}) error
}

func (m *MockValidator) URL(url string) error {
	if m.urlFunc != nil {
		return m.urlFunc(url)
	}
	return nil
}

func (m *MockValidator) Struct(s interface{}) error {
	if m.structFunc != nil {
		return m.structFunc(s)
	}
	return nil
}

func TestURLStorageInterface(t *testing.T) {
	tests := []struct {
		name        string
		storage     *MockURLStorage
		url         string
		withAlias   bool
		alias       string
		shortCode   string
		expectSave  string
		expectGet   string
		expectError bool
	}{
		{
			name: "successful save and get",
			storage: &MockURLStorage{
				saveURLFunc: func(ctx context.Context, url string, withAlias bool, alias string) (string, error) {
					return "abc123", nil
				},
				getURLFunc: func(ctx context.Context, shortCode string) (string, error) {
					return "https://example.com", nil
				},
			},
			url:         "https://example.com",
			withAlias:   false,
			alias:       "",
			shortCode:   "abc123",
			expectSave:  "abc123",
			expectGet:   "https://example.com",
			expectError: false,
		},
		{
			name: "save with alias",
			storage: &MockURLStorage{
				saveURLFunc: func(ctx context.Context, url string, withAlias bool, alias string) (string, error) {
					if withAlias && alias == "custom" {
						return "custom", nil
					}
					return "", nil
				},
			},
			url:         "https://example.com",
			withAlias:   true,
			alias:       "custom",
			shortCode:   "custom",
			expectSave:  "custom",
			expectGet:   "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test SaveURL
			shortCode, err := tt.storage.SaveURL(ctx, tt.url, tt.withAlias, tt.alias)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && shortCode != tt.expectSave {
				t.Errorf("Expected shortCode %s, got %s", tt.expectSave, shortCode)
			}

			// Test GetURL
			url, err := tt.storage.GetURL(ctx, tt.shortCode)
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && url != tt.expectGet {
				t.Errorf("Expected URL %s, got %s", tt.expectGet, url)
			}
		})
	}
}

func TestClickStorageInterface(t *testing.T) {
	tests := []struct {
		name        string
		storage     *MockClickStorage
		click       domain.Click
		shortCode   string
		limit       int
		offset      int
		expectError bool
	}{
		{
			name: "successful click operations",
			storage: &MockClickStorage{
				saveClickFunc: func(ctx context.Context, click domain.Click) error {
					return nil
				},
				getClicksFunc: func(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error) {
					return []domain.Click{
						{
							ID:        "click-1",
							ShortCode: shortCode,
							UserAgent: "Mozilla/5.0",
							IP:        "192.168.1.1",
							Referer:   "https://google.com",
							Timestamp: time.Now(),
						},
					}, nil
				},
				clickCountFunc: func(ctx context.Context, shortCode string) (int64, error) {
					return 100, nil
				},
				uniqueClickCountFunc: func(ctx context.Context, shortCode string) (int64, error) {
					return 50, nil
				},
			},
			click: domain.Click{
				ID:        "click-1",
				ShortCode: "abc123",
				UserAgent: "Mozilla/5.0",
				IP:        "192.168.1.1",
				Referer:   "https://google.com",
				Timestamp: time.Now(),
			},
			shortCode:   "abc123",
			limit:       10,
			offset:      0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test SaveClick
			err := tt.storage.SaveClick(ctx, tt.click)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Test GetClicks
			clicks, err := tt.storage.GetClicks(ctx, tt.shortCode, tt.limit, tt.offset)
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && len(clicks) == 0 {
				t.Error("Expected clicks but got empty slice")
			}

			// Test ClickCount
			count, err := tt.storage.ClickCount(ctx, tt.shortCode)
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && count == 0 {
				t.Error("Expected count > 0")
			}

			// Test UniqueClickCount
			uniqueCount, err := tt.storage.UniqueClickCount(ctx, tt.shortCode)
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && uniqueCount == 0 {
				t.Error("Expected unique count > 0")
			}
		})
	}
}

func TestCacheStorageInterface(t *testing.T) {
	tests := []struct {
		name        string
		storage     *MockCacheStorage
		shortCode   string
		url         string
		usage       int64
		expectGet   string
		expectError bool
	}{
		{
			name: "successful cache operations",
			storage: &MockCacheStorage{
				getURLFunc: func(ctx context.Context, shortCode string) (string, error) {
					return "https://example.com", nil
				},
				setURLFunc: func(ctx context.Context, shortCode, url string, usage int64) error {
					return nil
				},
			},
			shortCode:   "abc123",
			url:         "https://example.com",
			usage:       100,
			expectGet:   "https://example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test SetURL
			err := tt.storage.SetURL(ctx, tt.shortCode, tt.url, tt.usage)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Test GetURL
			url, err := tt.storage.GetURL(ctx, tt.shortCode)
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && url != tt.expectGet {
				t.Errorf("Expected URL %s, got %s", tt.expectGet, url)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	urlStorage := &MockURLStorage{}
	clickStorage := &MockClickStorage{}
	cache := &MockCacheStorage{}

	// Create a real validator instance for testing
	validator := validator.New()

	server := New(urlStorage, clickStorage, *validator, cache, "0.0.0.0:8080")

	if server == nil {
		t.Error("Expected server to be created, got nil")
	}
	if server.urlStorage != urlStorage {
		t.Error("Expected urlStorage to be set correctly")
	}
	if server.clickStorage != clickStorage {
		t.Error("Expected clickStorage to be set correctly")
	}
	if server.cache != cache {
		t.Error("Expected cache to be set correctly")
	}
	if len(server.addrs) == 0 {
		t.Error("Expected addrs to be set")
	}
}