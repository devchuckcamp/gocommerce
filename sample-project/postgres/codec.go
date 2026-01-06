package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

func scanNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func scanNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func toJSONB(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func fromJSONB[T any](b []byte, out *T) error {
	if len(b) == 0 {
		return nil
	}
	// Some drivers return "null" for JSONB.
	if string(b) == "null" {
		return nil
	}
	return json.Unmarshal(b, out)
}

func moneyFrom(amount int64, currency string) (money.Money, error) {
	return money.New(amount, currency)
}

func mustSameCurrency(currency string, m money.Money) (string, error) {
	if currency == "" {
		return m.Currency, nil
	}
	if m.Currency != "" && m.Currency != currency {
		return "", fmt.Errorf("currency mismatch: %s vs %s", currency, m.Currency)
	}
	return currency, nil
}
