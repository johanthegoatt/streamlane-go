package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRankEndpointReturnsPlan(t *testing.T) {
	body := `{"tasks":[{"name":"a","impact":8,"urgency":8,"effort":2,"risk":1}],"budget":5}`
	rec := httptest.NewRecorder()
	newMux(10).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/rank", strings.NewReader(body)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
	}
	var resp rankResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if len(resp.Plan) != 1 || resp.Used != 2 {
		t.Fatalf("unexpected plan: %+v", resp)
	}
}

func TestRankEndpointRejectsGet(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux(10).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/rank", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("expected Allow: POST header")
	}
}

func TestRankEndpointRejectsOversizedBody(t *testing.T) {
	huge := `{"tasks":[{"name":"` + strings.Repeat("x", maxBodyBytes) + `"}]}`
	rec := httptest.NewRecorder()
	newMux(10).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/rank", strings.NewReader(huge)))

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestRankEndpointRejectsMalformedJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux(10).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/rank", strings.NewReader("{")))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
