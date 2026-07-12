package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"risk-api/internal/handlers"
	"risk-api/internal/models"
	"risk-api/internal/store"
)

func TestHandlers(t *testing.T) {
	// Initialize memory store
	s, err := store.NewMemoryStore()
	if err != nil {
		t.Fatalf("failed to initialize test store: %v", err)
	}

	h := handlers.NewHandler(s)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	t.Run("GET /healthz", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp["status"] != "ok" {
			t.Errorf("expected status ok, got %s", resp["status"])
		}
	})

	t.Run("GET /cases/CASE-0001", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/cases/CASE-0001", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var c models.Case
		if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
			t.Fatal(err)
		}
		if c.ID != "CASE-0001" {
			t.Errorf("expected CASE-0001, got %s", c.ID)
		}
	})

	t.Run("GET /cases/CASE-INVALID (404)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/cases/CASE-INVALID", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}

		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp["error"] != "case not found" {
			t.Errorf("expected 'case not found' error, got %q", resp["error"])
		}
	})

	t.Run("GET /accounts/ACC-0001/kyc", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/accounts/ACC-0001/kyc", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var k models.KYCRecord
		if err := json.NewDecoder(rec.Body).Decode(&k); err != nil {
			t.Fatal(err)
		}
		if k.AccountID != "ACC-0001" {
			t.Errorf("expected ACC-0001, got %s", k.AccountID)
		}
	})

	t.Run("GET /accounts/ACC-0001/transactions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/accounts/ACC-0001/transactions", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var txs []models.Transaction
		if err := json.NewDecoder(rec.Body).Decode(&txs); err != nil {
			t.Fatal(err)
		}
		if len(txs) == 0 {
			t.Errorf("expected transactions, got none")
		}
	})

	t.Run("GET /accounts/ACC-0001/transactions with since_days filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/accounts/ACC-0001/transactions?since_days=30", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var txs []models.Transaction
		if err := json.NewDecoder(rec.Body).Decode(&txs); err != nil {
			t.Fatal(err)
		}
		// Because it filters, let's just make sure it compiles and parses successfully.
	})

	t.Run("GET /accounts/ACC-INVALID/transactions (404)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/accounts/ACC-INVALID/transactions", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("POST /flags (Success)", func(t *testing.T) {
		body := `{"account_id": "ACC-0001", "reason": "Suspicious volume", "raised_by": "tester"}`
		req := httptest.NewRequest("POST", "/flags", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}

		var f models.Flag
		if err := json.NewDecoder(rec.Body).Decode(&f); err != nil {
			t.Fatal(err)
		}
		if f.AccountID != "ACC-0001" || f.Reason != "Suspicious volume" || f.RaisedBy != "tester" {
			t.Errorf("unexpected flag properties: %+v", f)
		}
		if !f.RequiresAck {
			t.Errorf("expected requires_ack to be true")
		}
		if f.ID == "" {
			t.Errorf("expected non-empty flag ID")
		}
	})

	t.Run("POST /flags (Bad Request)", func(t *testing.T) {
		body := `{"account_id": "", "reason": "Missing account id", "raised_by": "tester"}`
		req := httptest.NewRequest("POST", "/flags", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}
