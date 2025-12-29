package store

import (
	"sync"
	"time"

	"0xnetworth/backend/internal/models"
)

// Store is an in-memory store for investment data
type Store struct {
	mu          sync.RWMutex
	portfolios  map[string]*models.Portfolio
	investments map[string]*models.Investment
	networth    *models.NetWorth
	lastSync    time.Time
}

// NewStore creates a new in-memory store
func NewStore() *Store {
	return &Store{
		portfolios:  make(map[string]*models.Portfolio),
		investments: make(map[string]*models.Investment),
		networth:    &models.NetWorth{},
	}
}

// Portfolio operations

// GetAllPortfolios returns all portfolios
func (s *Store) GetAllPortfolios() []*models.Portfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()

	portfolios := make([]*models.Portfolio, 0, len(s.portfolios))
	for _, p := range s.portfolios {
		portfolios = append(portfolios, p)
	}
	return portfolios
}

// GetPortfoliosByPlatform returns portfolios for a specific platform
func (s *Store) GetPortfoliosByPlatform(platform models.Platform) []*models.Portfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()

	portfolios := make([]*models.Portfolio, 0)
	for _, p := range s.portfolios {
		if p.Platform == platform {
			portfolios = append(portfolios, p)
		}
	}
	return portfolios
}

// GetPortfolioByID returns a portfolio by ID
func (s *Store) GetPortfolioByID(id string) (*models.Portfolio, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	portfolio, exists := s.portfolios[id]
	return portfolio, exists
}

// CreateOrUpdatePortfolio creates or updates a portfolio
func (s *Store) CreateOrUpdatePortfolio(portfolio *models.Portfolio) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.portfolios[portfolio.ID] = portfolio
}

// DeletePortfolio deletes a portfolio by ID
func (s *Store) DeletePortfolio(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.portfolios[id]; !exists {
		return false
	}
	delete(s.portfolios, id)
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

	// Calculate total from investments (portfolios don't have balances, only holdings)
	totalValue := 0.0
	for _, investment := range s.investments {
		totalValue += investment.Value
		networth.ByPlatform[investment.Platform] += investment.Value
		networth.ByAssetType[investment.AssetType] += investment.Value
	}

	networth.TotalValue = totalValue
	networth.AccountCount = len(s.portfolios) // Use portfolio count instead of account count
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

