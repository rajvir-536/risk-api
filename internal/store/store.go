package store

import (
	"context"
	"errors"
	"risk-api/internal/models"
)

var (
	// ErrNotFound is returned when a record is not found in the store.
	ErrNotFound = errors.New("record not found")
	// ErrDuplicate is returned when attempting to insert a record with a duplicate unique key.
	ErrDuplicate = errors.New("record already exists")
)

// Store defines the storage operations for the risk-api service.
type Store interface {
	GetCase(ctx context.Context, id string) (*models.Case, error)
	GetTransactions(ctx context.Context, accountID string, sinceDays *int) ([]models.Transaction, error)
	GetKYCRecord(ctx context.Context, accountID string) (*models.KYCRecord, error)
	CreateFlag(ctx context.Context, flag *models.Flag) error
}
