package workers

import (
	"context"

	"shared/domain/product"
)

// ProductViewSync flushes all cached product view counts to the database.
func ProductViewSync(ctx context.Context, productSvc *product.ProductService) error {
	if err := productSvc.FlushAllViewCounts(ctx); err != nil {
		return err
	}
	return nil
}
