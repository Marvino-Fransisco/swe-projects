package product

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

// ProductSummary is a slim representation of a product for mobile list views.
type ProductSummary struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ListProductsResponse struct {
	Products []ProductSummary `json:"products"`
	Total    int64            `json:"total"`
}

type SearchProductsResponse struct {
	Products []ProductSummary `json:"products"`
	Total    int64            `json:"total"`
}

type TrackViewResponse struct {
	ProductID string `json:"product_id"`
	ViewCount int64  `json:"view_count"`
}
