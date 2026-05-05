// Package user defines the User domain model and its value objects.
package user

import (
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
)

// --- FullName Value Object ---

// FullName represents a validated user full name.
// It ensures no symbol characters except spaces, cannot be only spaces,
// and has no leading or trailing spaces.
type FullName string

var (
	// symbolExceptSpaceRegex matches any character that is NOT a letter, digit, or space.
	symbolExceptSpaceRegex = regexp.MustCompile(`[^a-zA-Z0-9\s]`)
)

// NewFullName creates and validates a FullName value object.
func NewFullName(name string) (FullName, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("full name cannot be empty or only spaces")
	}

	if name != strings.TrimLeft(name, " ") {
		return "", errors.New("full name cannot have leading spaces")
	}

	if name != strings.TrimRight(name, " ") {
		return "", errors.New("full name cannot have trailing spaces")
	}

	if symbolExceptSpaceRegex.MatchString(name) {
		return "", errors.New("full name cannot contain symbol characters")
	}

	return FullName(name), nil
}

// String returns the string representation of the FullName.
func (f FullName) String() string {
	return string(f)
}

// Value implements the driver.Valuer interface for database writes.
func (f FullName) Value() (driver.Value, error) {
	return string(f), nil
}

// Scan implements the sql.Scanner interface for database reads.
func (f *FullName) Scan(value interface{}) error {
	if value == nil {
		*f = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*f = FullName(v)
	case []byte:
		*f = FullName(v)
	default:
		return errors.New("cannot scan FullName from non-string type")
	}

	return nil
}

// --- Email Value Object ---

// Email represents a validated email address.
type Email string

// emailRegex provides basic email format validation.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates and validates an Email value object.
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", errors.New("email cannot be empty")
	}

	if !emailRegex.MatchString(email) {
		return "", errors.New("invalid email format")
	}

	return Email(email), nil
}

// String returns the string representation of the Email.
func (e Email) String() string {
	return string(e)
}

// Value implements the driver.Valuer interface for database writes.
func (e Email) Value() (driver.Value, error) {
	return string(e), nil
}

// Scan implements the sql.Scanner interface for database reads.
func (e *Email) Scan(value interface{}) error {
	if value == nil {
		*e = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*e = Email(v)
	case []byte:
		*e = Email(v)
	default:
		return errors.New("cannot scan Email from non-string type")
	}

	return nil
}
