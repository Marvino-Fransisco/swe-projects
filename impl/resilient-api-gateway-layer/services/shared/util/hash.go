package util

// PasswordHasher defines the interface for password hashing and comparison.
// The actual implementation (e.g., bcrypt) is provided by the BFF services.
type PasswordHasher interface {
	// Hash generates a hashed version of the given plain-text password.
	Hash(password string) (string, error)

	// Compare checks if the given plain-text password matches the stored hash.
	Compare(password, hash string) error
}
