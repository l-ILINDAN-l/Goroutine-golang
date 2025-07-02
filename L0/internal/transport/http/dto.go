package http

import (
	"L0/internal/domain"
	"github.com/google/uuid"
	"time"
)

type DeliveryResponse struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
}

type PaymentResponse struct {
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int64  `json:"amount"`
	Bank         string `json:"bank"`
	DeliveryCost int64  `json:"delivery_cost"`
	GoodsTotal   int64  `json:"goods_total"`
	CustomFee    int64  `json:"custom_fee"`
}

type ItemResponse struct {
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int64  `json:"price"`
	Name        string `json:"name"`
	Sale        int64  `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int64  `json:"total_price"`
	Brand       string `json:"brand"`
}

type OrderResponse struct {
	OrderUID    uuid.UUID `json:"order_uid"`
	TrackNumber string    `json:"track_number"`
	Entry       string    `json:"entry"`

	Delivery *DeliveryResponse `json:"delivery"`
	Payment  *PaymentResponse  `json:"payment"`
	Items    []ItemResponse    `json:"items"`

	DeliveryService string    `json:"delivery_service"`
	Locale          string    `json:"locale"`
	DateCreated     time.Time `json:"date_created"`
}

func toDeliveryResponse(delivery *domain.Delivery) *DeliveryResponse {
	return &DeliveryResponse{
		Name:    delivery.Name,
		City:    delivery.City,
		Address: delivery.Address,
		Region:  delivery.Region,
	}
}

func toPaymentResponse(payment *domain.Payment) *PaymentResponse {
	return &PaymentResponse{
		Currency:     payment.Currency,
		Provider:     payment.Provider,
		Amount:       payment.Amount,
		Bank:         payment.Bank,
		DeliveryCost: payment.DeliveryCost,
		GoodsTotal:   payment.GoodsTotal,
		CustomFee:    payment.CustomFee,
	}
}

func toItemResponse(item domain.Item) ItemResponse {
	return ItemResponse{
		ChrtID:      item.ChrtID,
		TrackNumber: item.TrackNumber,
		Price:       item.Price,
		Name:        item.Name,
		Sale:        item.Sale,
		Size:        item.Size,
		TotalPrice:  item.TotalPrice,
		Brand:       item.Brand,
	}
}

func toItemsResponse(items []domain.Item) []ItemResponse {
	itemsResponse := make([]ItemResponse, len(items))
	for i, item := range items {
		itemsResponse[i] = toItemResponse(item)
	}
	return itemsResponse
}

func toOrderResponse(order *domain.Order) *OrderResponse {
	return &OrderResponse{
		OrderUID:    order.OrderUID,
		TrackNumber: order.TrackNumber,
		Entry:       order.Entry,

		Delivery: toDeliveryResponse(order.Delivery),
		Payment:  toPaymentResponse(order.Payment),
		Items:    toItemsResponse(order.Items),

		DeliveryService: order.DeliveryService,
		Locale:          order.Locale,
		DateCreated:     order.DateCreated,
	}
}
