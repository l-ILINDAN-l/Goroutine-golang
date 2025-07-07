package http

import (
	"L0/internal/domain"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestToOrderResponse(t *testing.T) {
	now := time.Now()
	testOrder := &domain.Order{
		OrderUID:          "test_uid",
		TrackNumber:       "test_track",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "secret_signature",
		CustomerID:        "test_customer",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       now,
		OOFShard:          "1",
		Delivery: &domain.Delivery{
			Name: "Test Testov",
		},
		Payment: &domain.Payment{
			Transaction: "test_transaction",
			Currency:    "USD",
		},
		Items: []domain.Item{
			{ChrtID: 12345},
		},
	}

	responseDTO := toOrderResponse(testOrder)

	assert.NotNil(t, responseDTO, "DTO не должен быть nil")
	assert.Equal(t, testOrder.OrderUID, responseDTO.OrderUID, "OrderUID должен совпадать")
	assert.Equal(t, testOrder.TrackNumber, responseDTO.TrackNumber, "TrackNumber должен совпадать")
	assert.Equal(t, testOrder.DeliveryService, responseDTO.DeliveryService, "DeliveryService должен совпадать")
	assert.Equal(t, testOrder.DateCreated, responseDTO.DateCreated, "DateCreated должен совпадать")

	assert.NotNil(t, responseDTO.Delivery, "Delivery в DTO не должен быть nil")
	assert.Equal(t, testOrder.Delivery.Name, responseDTO.Delivery.Name, "Имя в доставке должно совпадать")

	assert.NotNil(t, responseDTO.Payment, "Payment в DTO не должен быть nil")
	assert.Equal(t, testOrder.Payment.Currency, responseDTO.Payment.Currency, "Валюта в оплате должна совпадать")

	assert.Len(t, responseDTO.Items, 1, "Должен быть один товар")
	assert.Equal(t, testOrder.Items[0].ChrtID, responseDTO.Items[0].ChrtID, "ChrtID товара должен совпадать")
}
