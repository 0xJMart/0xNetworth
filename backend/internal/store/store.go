package store

import (
	"sync"
	"time"

	"0xnetworth/backend/internal/models"
)

// Store is an in-memory store for investment data
type Store struct {
	mu          sync.RWMutex
	accounts    map[string]*models.Account
	investments map[string]*models.Investment
	networth    *models.NetWorth
	lastSync    time.Time
}

// NewStore creates a new in-memory store
func NewStore() *Store {
	return &Store{
		accounts:    make(map[string]*models.Account),
		investments: make(map[string]*models.Investment),
		networth:    &models.NetWorth{},
	}
}

// Account operations

// GetAllAccounts returns all accounts
func (s *Store) GetAllAccounts() []*models.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]*models.Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		accounts = append(accounts, a)
	}
	return accounts
}

// GetAccountsByPlatform returns accounts for a specific platform
func (s *Store) GetAccountsByPlatform(platform models.Platform) []*models.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]*models.Account, 0)
	for _, a := range s.accounts {
		if a.Platform == platform {
			accounts = append(accounts, a)
		}
	}
	return accounts
}

// GetAccountByID returns an account by ID
func (s *Store) GetAccountByID(id string) (*models.Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.accounts[id]
	return account, exists
}

// CreateOrUpdateAccount creates or updates an account
func (s *Store) CreateOrUpdateAccount(account *models.Account) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts[account.ID] = account
}

// DeleteAccount deletes an account by ID
func (s *Store) DeleteAccount(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accounts[id]; !exists {
		return false
	}
	delete(s.accounts, id)
	return true
}

// Investment operations

// GetAllInvestments returns all investments
func (s *Store) GetAllInvestments() []*models.Investment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	investments := make([]*models.Investment, 0, len(s.investments))
	for _, inv := range s.investments {
		investments = append(investments, inv)
	}
	return investments
}

// GetInvestmentsByAccount returns investments for a specific account
func (s *Store) GetInvestmentsByAccount(accountID string) []*models.Investment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	investments := make([]*models.Investment, 0)
	for _, inv := range s.investments {
		if inv.AccountID == accountID {
			investments = append(investments, inv)
		}
	}
	return investments
}

// GetInvestmentsByPlatform returns investments for a specific platform
func (s *Store) GetInvestmentsByPlatform(platform models.Platform) []*models.Investment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	investments := make([]*models.Investment, 0)
	for _, inv := range s.investments {
		if inv.Platform == platform {
			investments = append(investments, inv)
		}
	}
	return investments
}

// CreateOrUpdateInvestment creates or updates an investment
func (s *Store) CreateOrUpdateInvestment(investment *models.Investment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.investments[investment.ID] = investment
}

// DeleteInvestment deletes an investment by ID
func (s *Store) DeleteInvestment(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.investments[id]; !exists {
		return false
	}
	delete(s.investments, id)
	return true
}

// NetWorth operations

// GetNetWorth returns the current net worth
func (s *Store) GetNetWorth() *models.NetWorth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	networth := *s.networth
	return &networth
}

// UpdateNetWorth updates the net worth calculation
func (s *Store) UpdateNetWorth(networth *models.NetWorth) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.networth = networth
}

// RecalculateNetWorth recalculates net worth from current accounts and investments
func (s *Store) RecalculateNetWorth() {
	s.mu.Lock()
	defer s.mu.Unlock()

	networth := &models.NetWorth{
		ByPlatform:   make(map[models.Platform]float64),
		ByAssetType:  make(map[string]float64),
		Currency:     "USD", // Default currency
		LastCalculated: time.Now().UTC().Format(time.RFC3339),
	}

	// Calculate total from accounts
	totalValue := 0.0
	accountCount := 0
	for _, account := range s.accounts {
		totalValue += account.Balance
		networth.ByPlatform[account.Platform] += account.Balance
		accountCount++
	}

	// Add investment values
	for _, investment := range s.investments {
		totalValue += investment.Value
		networth.ByPlatform[investment.Platform] += investment.Value
		networth.ByAssetType[investment.AssetType] += investment.Value
	}

	networth.TotalValue = totalValue
	networth.AccountCount = accountCount
	s.networth = networth
}

// GetLastSyncTime returns the last sync time
func (s *Store) GetLastSyncTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastSync
}

// SetLastSyncTime sets the last sync time
func (s *Store) SetLastSyncTime(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastSync = t
}

