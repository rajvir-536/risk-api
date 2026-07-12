package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"risk-api/internal/models"
	"risk-api/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

// WriteJSON is a helper to write JSON responses.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteError is a helper to write JSON error responses.
func WriteError(w http.ResponseWriter, status int, errMsg string) {
	WriteJSON(w, status, map[string]string{"error": errMsg})
}

// RegisterRoutes registers the handlers on the given ServeMux using Go 1.22+ routing syntax.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Healthz)
	mux.HandleFunc("GET /cases/{id}", h.GetCase)
	mux.HandleFunc("GET /accounts/{id}/transactions", h.GetTransactions)
	mux.HandleFunc("GET /accounts/{id}/kyc", h.GetKYCRecord)
	mux.HandleFunc("POST /flags", h.CreateFlag)
}

// Healthz handles the health check endpoint.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetCase retrieves a specific case by ID.
func (h *Handler) GetCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing case id")
		return
	}

	c, err := h.store.GetCase(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "case not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get case")
		return
	}

	WriteJSON(w, http.StatusOK, c)
}

// GetTransactions retrieves transactions for an account, with optional since_days filter.
func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing account id")
		return
	}

	var sinceDays *int
	sinceDaysStr := r.URL.Query().Get("since_days")
	if sinceDaysStr != "" {
		val, err := strconv.Atoi(sinceDaysStr)
		if err != nil || val < 0 {
			WriteError(w, http.StatusBadRequest, "invalid since_days parameter")
			return
		}
		sinceDays = &val
	}

	txs, err := h.store.GetTransactions(r.Context(), id, sinceDays)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "account or transactions not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get transactions")
		return
	}

	if txs == nil {
		txs = []models.Transaction{}
	}

	WriteJSON(w, http.StatusOK, txs)
}

// GetKYCRecord retrieves the KYC record for an account.
func (h *Handler) GetKYCRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing account id")
		return
	}

	kyc, err := h.store.GetKYCRecord(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "kyc record not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get kyc record")
		return
	}

	if kyc.WatchlistMatches == nil {
		kyc.WatchlistMatches = []string{}
	}

	WriteJSON(w, http.StatusOK, kyc)
}

// CreateFlag handles creation of a new risk/fraud flag.
func (h *Handler) CreateFlag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID string `json:"account_id"`
		Reason    string `json:"reason"`
		RaisedBy  string `json:"raised_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AccountID == "" || req.Reason == "" {
		WriteError(w, http.StatusBadRequest, "account_id and reason are required")
		return
	}

	flag := &models.Flag{
		AccountID: req.AccountID,
		Reason:    req.Reason,
		RaisedBy:  req.RaisedBy,
	}

	err := h.store.CreateFlag(r.Context(), flag)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			WriteError(w, http.StatusConflict, "flag already exists")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to create flag")
		return
	}

	WriteJSON(w, http.StatusCreated, flag)
}
