package product

import (
	"context"

	productDomain "shared/domain/product"
)

type ProductUseCase interface {
	GetByID(ctx context.Context, req GetProductRequest) (*productDomain.Product, error)
	List(ctx context.Context, req ListProductsRequest) (*ListProductsResponse, error)
	Search(ctx context.Context, req SearchProductsRequest) (*SearchProductsResponse, error)
	GetCategories(ctx context.Context) ([]productDomain.Category, error)
	TrackView(ctx context.Context, req TrackViewRequest) (*TrackViewResponse, error)
}
