package workers

import (
	"context"
	"encoding/json"
	"shared/domain/product"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func cacheWarmer(ctx context.Context, db *gorm.DB, rdb *redis.Client) error {
	var products []product.Product

	if err := db.WithContext(ctx).Find(&products).Error; err != nil {
		return err
	}

	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	err = rdb.Set(ctx, "products", data, time.Hour*24*7).Err()
	if err != nil {
		return err
	}

	return nil

}
