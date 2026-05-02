package tx

import (
	"context"

	"gorm.io/gorm"
)

// txKey is the unexported context key type used to store the transaction-scoped *gorm.DB.
type txKey struct{}

// TxKey is the exported context key for accessing the transaction from context.
var TxKey txKey

// DBTransaction executes a function within a database transaction.
// The transaction's *gorm.DB is stored in the context so that repository methods
// automatically participate in the transaction via DBFromContext.
type DBTransaction func(ctx context.Context, fn func(ctx context.Context) error) error

// NewDBTransaction creates a DBTransaction using a GORM database connection.
//
// Usage:
//
//	tx := tx.NewDBTransaction(db)
//	err := tx(ctx, func(txCtx context.Context) error {
//	    // All repository calls within this function will use the same transaction
//	    // as long as they call tx.DBFromContext(txCtx, defaultDB)
//	    return repo.Save(txCtx, entity)
//	})
func NewDBTransaction(db *gorm.DB) DBTransaction {
	return func(ctx context.Context, fn func(ctx context.Context) error) error {
		return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := context.WithValue(ctx, TxKey, tx)
			return fn(txCtx)
		})
	}
}

// DBFromContext returns the transaction-scoped *gorm.DB from the context,
// or falls back to the provided defaultDB if no transaction is active.
//
// Repository methods should use this to transparently participate in transactions:
//
//	func (r *MyRepo) Save(ctx context.Context, entity Entity) error {
//	    db := tx.DBFromContext(ctx, r.db)
//	    return db.WithContext(ctx).Create(&entity).Error
//	}
func DBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(TxKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
