package product

import (
	"database/sql/driver"
	"errors"
)

// ProductStatus represents the inventory status of a product.
type ProductStatus string

const (
	// ProductStatusEmpty indicates the product is completely out of stock.
	ProductStatusEmpty ProductStatus = "EMPTY"

	// ProductStatusInsufficient indicates the product stock is below a threshold.
	ProductStatusInsufficient ProductStatus = "INSUFFICIENT"

	// ProductStatusDanger indicates the product stock is critically low.
	ProductStatusDanger ProductStatus = "DANGER"
)

// validProductStatuses contains all allowed product status values.
var validProductStatuses = map[ProductStatus]bool{
	ProductStatusEmpty:        true,
	ProductStatusInsufficient: true,
	ProductStatusDanger:       true,
}

// NewProductStatus creates and validates a ProductStatus value.
func NewProductStatus(status string) (ProductStatus, error) {
	ps := ProductStatus(status)
	if !validProductStatuses[ps] {
		return "", errors.New("invalid product status: must be EMPTY, INSUFFICIENT, or DANGER")
	}

	return ps, nil
}

// String returns the string representation of the ProductStatus.
func (s ProductStatus) String() string {
	return string(s)
}

// Value implements the driver.Valuer interface for database writes.
func (s ProductStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements the sql.Scanner interface for database reads.
func (s *ProductStatus) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*s = ProductStatus(v)
	case []byte:
		*s = ProductStatus(v)
	default:
		return errors.New("cannot scan ProductStatus from non-string type")
	}

	return nil
}
