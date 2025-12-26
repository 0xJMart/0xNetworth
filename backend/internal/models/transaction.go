package models

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeBuy     TransactionType = "buy"
	TransactionTypeSell    TransactionType = "sell"
	TransactionTypeDeposit TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID          string          `json:"id"`
	AccountID   string          `json:"account_id"`
	Platform    Platform        `json:"platform"`
	Type        TransactionType `json:"type"`
	Symbol      string          `json:"symbol,omitempty"`      // Asset symbol if applicable
	Quantity    float64         `json:"quantity,omitempty"`    // Quantity if applicable
	Amount      float64         `json:"amount"`                // Transaction amount
	Currency    string          `json:"currency"`
	Fee         float64         `json:"fee,omitempty"`        // Transaction fee
	Timestamp   string          `json:"timestamp"`            // ISO 8601 timestamp
	Description string          `json:"description,omitempty"`
}

