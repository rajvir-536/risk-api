package models

import (
	"time"
)

// Case represents a risk/fraud case.
type Case struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Status    string    `json:"status"` // open, escalated, closed
	OpenedAt  time.Time `json:"opened_at"`
	Summary   string    `json:"summary"`
}

// Transaction represents a financial transaction.
type Transaction struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Merchant  string    `json:"merchant"`
	Timestamp time.Time `json:"timestamp"`
}

// KYCRecord represents an account's KYC information.
type KYCRecord struct {
	AccountID         string   `json:"account_id"`
	VerificationLevel string   `json:"verification_level"` // e.g. Level 1, Level 2, Level 3
	PriorFlags        int      `json:"prior_flags"`
	Jurisdiction      string   `json:"jurisdiction"`
	WatchlistMatches  []string `json:"watchlist_matches"` // JSON array, nullable
}

// Flag represents a raised risk flag.
type Flag struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Reason      string    `json:"reason"`
	RaisedBy    string    `json:"raised_by"`
	CreatedAt   time.Time `json:"created_at"`
	RequiresAck bool      `json:"requires_ack"`
}
