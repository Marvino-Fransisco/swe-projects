package workers

import (
	"context"

	"shared/domain/product"
)

func ProductViewSync(ctx context.Context, productSvc *product.ProductService) error {
	if err := productSvc.FlushAllViewCounts(ctx); err != nil {
		return err
	}
	return nil
}
