package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kxddry/wbf/ginext"
)

func TestPostNotifyTelegramValidation(t *testing.T) {
	r := ginext.New()
	RegisterRoutes(r, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// invalid: not 9 digits
	body := []byte(`{"channel":"telegram","recipient":"123","message":"hi"}`)
	res, err := http.Post(ts.URL+"/notify", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http post error: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
