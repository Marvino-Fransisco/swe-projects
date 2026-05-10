package product

import productDomain "shared/domain/product"

type GetProductRequest struct {
	ID string
}

type ListProductsRequest struct {
	CategoryID string
	Search     string
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
	MinPrice   float64
	MaxPrice   float64
}

type SearchProductsRequest struct {
	Query    string
	Page     int
	PageSize int
}

type TrackViewRequest struct {
	ProductID string
}

type ListProductsResponse struct {
	Products []productDomain.Product `json:"products"`
	Total    int64                    `json:"total"`
}

type SearchProductsResponse struct {
	Products []productDomain.Product `json:"products"`
	Total    int64                   `json:"total"`
}

type TrackViewResponse struct {
	ProductID string `json:"product_id"`
	ViewCount int64  `json:"view_count"`
}
