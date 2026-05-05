package checkout

import "shared/domain/order"

type PlaceOrderRequest struct {
	UserID string
}

type GetOrderRequest struct {
	UserID  string
	OrderID string
}

type GetOrderHistoryRequest struct {
	UserID   string
	Page     int
	PageSize int
}

type OrderWithDetailsResponse struct {
	Order   *order.Order        `json:"order"`
	Details []order.OrderDetail `json:"details"`
}

type OrderHistoryResponse struct {
	Orders []order.Order `json:"orders"`
	Total  int64         `json:"total"`
}
