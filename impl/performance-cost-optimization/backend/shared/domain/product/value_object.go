// Package product defines the Product domain model, value objects, and status enum.
package product

import (
	"database/sql/driver"
	"errors"
)

// --- Price Value Object ---

// Price represents a validated product price that cannot be negative.
type Price float64

// NewPrice creates and validates a Price value object.
func NewPrice(p float64) (Price, error) {
	if p < 0 {
		return 0, errors.New("price cannot be negative")
	}

	return Price(p), nil
}

// Float64 returns the float64 representation of the Price.
func (p Price) Float64() float64 {
	return float64(p)
}

// Value implements the driver.Valuer interface for database writes.
func (p Price) Value() (driver.Value, error) {
	return float64(p), nil
}

// Scan implements the sql.Scanner interface for database reads.
func (p *Price) Scan(value interface{}) error {
	if value == nil {
		*p = 0
		return nil
	}

	switch v := value.(type) {
	case float64:
		*p = Price(v)
	case int64:
		*p = Price(v)
	default:
		return errors.New("cannot scan Price from non-numeric type")
	}

	return nil
}

// --- Stock Value Object ---

// Stock represents a validated product stock quantity that cannot be negative.
type Stock int

// NewStock creates and validates a Stock value object.
func NewStock(s int) (Stock, error) {
	if s < 0 {
		return 0, errors.New("stock cannot be negative")
	}

	return Stock(s), nil
}

// Int returns the int representation of the Stock.
func (s Stock) Int() int {
	return int(s)
}

// Value implements the driver.Valuer interface for database writes.
func (s Stock) Value() (driver.Value, error) {
	return int64(s), nil
}

// Scan implements the sql.Scanner interface for database reads.
func (s *Stock) Scan(value interface{}) error {
	if value == nil {
		*s = 0
		return nil
	}

	switch v := value.(type) {
	case int64:
		*s = Stock(v)
	case float64:
		*s = Stock(int(v))
	default:
		return errors.New("cannot scan Stock from non-numeric type")
	}

	return nil
}
