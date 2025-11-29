package money

import (
	"errors"
	"fmt"
)

// Money represents a monetary value with currency.
// It uses cents/minor units to avoid floating point issues.
type Money struct {
	Amount   int64  // Amount in smallest currency unit (cents, pence, etc.)
	Currency string // ISO 4217 currency code (USD, EUR, GBP, etc.)
}

var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrNegativeAmount   = errors.New("amount cannot be negative")
	ErrInvalidCurrency  = errors.New("invalid currency code")
)

// New creates a new Money value.
func New(amount int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}
	return Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

// NewFromFloat creates Money from a float (e.g., 19.99 USD).
// Note: Use with caution due to floating point precision.
func NewFromFloat(amount float64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}
	return Money{
		Amount:   int64(amount * 100),
		Currency: currency,
	}, nil
}

// Zero returns zero money in the given currency.
func Zero(currency string) Money {
	return Money{Amount: 0, Currency: currency}
}

// Add adds two Money values. Returns error if currencies differ.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Subtract subtracts other from m. Returns error if currencies differ.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

// Multiply multiplies the money by a factor.
func (m Money) Multiply(factor float64) Money {
	return Money{
		Amount:   int64(float64(m.Amount) * factor),
		Currency: m.Currency,
	}
}

// MultiplyInt multiplies the money by an integer.
func (m Money) MultiplyInt(factor int) Money {
	return Money{
		Amount:   m.Amount * int64(factor),
		Currency: m.Currency,
	}
}

// IsNegative returns true if the amount is negative.
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive returns true if the amount is positive.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// LessThan returns true if m is less than other.
func (m Money) LessThan(other Money) (bool, error) {
	if m.Currency != other.Currency {
		return false, ErrCurrencyMismatch
	}
	return m.Amount < other.Amount, nil
}

// GreaterThan returns true if m is greater than other.
func (m Money) GreaterThan(other Money) (bool, error) {
	if m.Currency != other.Currency {
		return false, ErrCurrencyMismatch
	}
	return m.Amount > other.Amount, nil
}

// Equals returns true if m equals other.
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// ToFloat converts to a float (dollars, euros, etc.).
func (m Money) ToFloat() float64 {
	return float64(m.Amount) / 100.0
}

// String returns a human-readable representation.
func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.Currency, m.ToFloat())
}

// Allocate divides money into n parts, handling remainders correctly.
// Useful for splitting amounts (e.g., split a discount across items).
func (m Money) Allocate(n int) []Money {
	if n <= 0 {
		return []Money{}
	}

	base := m.Amount / int64(n)
	remainder := m.Amount % int64(n)

	result := make([]Money, n)
	for i := 0; i < n; i++ {
		amount := base
		if i < int(remainder) {
			amount++
		}
		result[i] = Money{
			Amount:   amount,
			Currency: m.Currency,
		}
	}
	return result
}
