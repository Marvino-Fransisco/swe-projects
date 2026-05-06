package workers

import (
	"context"
	"log"

	"shared/domain/cart"
)

func CartSync(ctx context.Context, cartSvc *cart.CartService, cartCacheRepo cart.CartCacheRepository) error {
	userIDs, err := cartCacheRepo.GetCartDirtyMembers(ctx)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := cartSvc.SyncToDB(ctx, userID); err != nil {
			log.Printf("failed to sync cart for user %s: %v", userID, err)
			continue
		}
	}

	return nil
}
