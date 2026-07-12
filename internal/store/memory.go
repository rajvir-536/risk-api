package store

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"risk-api/data"
	"risk-api/internal/models"
)

type MemoryStore struct {
	mu           sync.RWMutex
	cases        map[string]models.Case
	transactions map[string][]models.Transaction // Keyed by AccountID
	kycRecords   map[string]models.KYCRecord     // Keyed by AccountID
	flags        map[string]models.Flag          // Keyed by Flag ID
}

// NewMemoryStore loads embedded seed files and returns a initialized MemoryStore.
func NewMemoryStore() (*MemoryStore, error) {
	s := &MemoryStore{
		cases:        make(map[string]models.Case),
		transactions: make(map[string][]models.Transaction),
		kycRecords:   make(map[string]models.KYCRecord),
		flags:        make(map[string]models.Flag),
	}

	if err := s.loadData(); err != nil {
		return nil, fmt.Errorf("failed to load seed data: %w", err)
	}

	return s, nil
}

func (s *MemoryStore) loadData() error {
	// 1. Cases
	casesBytes, err := data.FS.ReadFile("cases.json")
	if err != nil {
		return fmt.Errorf("failed to read cases.json: %w", err)
	}
	var casesList []models.Case
	if err := json.Unmarshal(casesBytes, &casesList); err != nil {
		return fmt.Errorf("failed to parse cases.json: %w", err)
	}
	for _, c := range casesList {
		s.cases[c.ID] = c
	}

	// 2. Transactions
	txBytes, err := data.FS.ReadFile("transactions.json")
	if err != nil {
		return fmt.Errorf("failed to read transactions.json: %w", err)
	}
	var txList []models.Transaction
	if err := json.Unmarshal(txBytes, &txList); err != nil {
		return fmt.Errorf("failed to parse transactions.json: %w", err)
	}
	for _, t := range txList {
		s.transactions[t.AccountID] = append(s.transactions[t.AccountID], t)
	}

	// 3. KYC Records
	kycBytes, err := data.FS.ReadFile("kyc_records.json")
	if err != nil {
		return fmt.Errorf("failed to read kyc_records.json: %w", err)
	}
	var kycList []models.KYCRecord
	if err := json.Unmarshal(kycBytes, &kycList); err != nil {
		return fmt.Errorf("failed to parse kyc_records.json: %w", err)
	}
	for _, k := range kycList {
		s.kycRecords[k.AccountID] = k
	}

	// 4. Flags (loaded to verify format/structure)
	flagsBytes, err := data.FS.ReadFile("flags.json")
	if err != nil {
		return fmt.Errorf("failed to read flags.json: %w", err)
	}
	var flagsList []models.Flag
	if err := json.Unmarshal(flagsBytes, &flagsList); err != nil {
		return fmt.Errorf("failed to parse flags.json: %w", err)
	}
	for _, f := range flagsList {
		s.flags[f.ID] = f
	}

	return nil
}

func (s *MemoryStore) GetCase(ctx context.Context, id string) (*models.Case, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.cases[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &c, nil
}

func (s *MemoryStore) GetTransactions(ctx context.Context, accountID string, sinceDays *int) ([]models.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	txs, ok := s.transactions[accountID]
	if !ok || len(txs) == 0 {
		return nil, ErrNotFound
	}

	if sinceDays == nil {
		res := make([]models.Transaction, len(txs))
		copy(res, txs)
		return res, nil
	}

	cutoff := time.Now().AddDate(0, 0, -*sinceDays)
	var filtered []models.Transaction
	for _, t := range txs {
		if t.Timestamp.After(cutoff) || t.Timestamp.Equal(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if filtered == nil {
		filtered = []models.Transaction{}
	}
	return filtered, nil
}

func (s *MemoryStore) GetKYCRecord(ctx context.Context, accountID string) (*models.KYCRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	k, ok := s.kycRecords[accountID]
	if !ok {
		return nil, ErrNotFound
	}
	return &k, nil
}

func (s *MemoryStore) CreateFlag(ctx context.Context, flag *models.Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if flag.ID == "" {
		flag.ID = generateUUID()
	}
	if flag.CreatedAt.IsZero() {
		flag.CreatedAt = time.Now()
	}
	flag.RequiresAck = true

	s.flags[flag.ID] = *flag
	return nil
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = crand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
