package product

import (
	"context"

	"mobile-backend/apperror"

	productDomain "shared/domain/product"
)

type productUseCase struct {
	productSvc *productDomain.ProductService
}

func NewProductUseCase(productSvc *productDomain.ProductService) ProductUseCase {
	return &productUseCase{productSvc: productSvc}
}

// mapToSummary converts a slice of domain Products to slim ProductSummary responses.
func mapToSummary(products []productDomain.Product) []ProductSummary {
	result := make([]ProductSummary, 0, len(products))
	for _, p := range products {
		result = append(result, ProductSummary{
			ID:    p.ID,
			Name:  p.Name,
			Price: p.Price.Float64(),
		})
	}
	return result
}

func (uc *productUseCase) GetByID(ctx context.Context, req GetProductRequest) (*productDomain.Product, error) {
	return uc.productSvc.GetByID(ctx, req.ID)
}

func (uc *productUseCase) List(ctx context.Context, req ListProductsRequest) (*ListProductsResponse, error) {
	filter := productDomain.ProductFilter{
		CategoryID: req.CategoryID,
		Search:     req.Search,
		Page:       req.Page,
		PageSize:   req.PageSize,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
	}

	products, total, err := uc.productSvc.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ListProductsResponse{
		Products: mapToSummary(products),
		Total:    total,
	}, nil
}

func (uc *productUseCase) Search(ctx context.Context, req SearchProductsRequest) (*SearchProductsResponse, error) {
	products, total, err := uc.productSvc.Search(ctx, req.Query, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &SearchProductsResponse{
		Products: mapToSummary(products),
		Total:    total,
	}, nil
}

func (uc *productUseCase) GetCategories(ctx context.Context) ([]productDomain.Category, error) {
	return uc.productSvc.GetCategories(ctx)
}

func (uc *productUseCase) TrackView(ctx context.Context, req TrackViewRequest) (*TrackViewResponse, error) {
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	if err := uc.productSvc.IncrementViewCount(ctx, req.ProductID); err != nil {
		return nil, apperror.NewBadRequest("failed to track view")
	}

	count, err := uc.productSvc.GetViewCount(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewBadRequest("failed to get view count")
	}

	return &TrackViewResponse{
		ProductID: req.ProductID,
		ViewCount: count,
	}, nil
}
