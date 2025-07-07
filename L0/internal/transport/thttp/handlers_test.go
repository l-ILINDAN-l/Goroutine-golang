package thttp

import (
	"L0/internal/domain"
	"L0/internal/repository"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type OrderServiceMock struct {
	mock.Mock
}

func (m *OrderServiceMock) GetOrderByUID(ctx context.Context, uid string) (*domain.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}
func (m *OrderServiceMock) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	return m.Called(ctx, order).Error(0)
}
func (m *OrderServiceMock) WarmUpCache(ctx context.Context, limit uint) error {
	return m.Called(ctx, limit).Error(0)
}

func TestGetOrderHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(OrderServiceMock)

	now := time.Now()
	testUID := "test_order_uid_1234"
	expectedOrder := &domain.Order{
		OrderUID:          testUID,
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

	mockService.On("GetOrderByUID", mock.Anything, testUID).Return(expectedOrder, nil)

	router := gin.Default()
	server := NewHTTPServer(router, mockService, logrus.NewEntry(logrus.New()))
	router.GET("/order/:order_uid", server.getOrderHandler)

	req := httptest.NewRequest(http.MethodGet, "/order/"+testUID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder.OrderUID, response.OrderUID)

	mockService.AssertExpectations(t)
}

func TestGetOrderHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(OrderServiceMock)
	testUID := "non_existent_order_id"

	mockService.On("GetOrderByUID", mock.Anything, testUID).Return(nil, repository.ErrShardNotFound)

	router := gin.Default()
	server := NewHTTPServer(router, mockService, logrus.NewEntry(logrus.New()))
	router.GET("/order/:order_uid", server.getOrderHandler)

	req := httptest.NewRequest(http.MethodGet, "/order/"+testUID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "order not found")
	mockService.AssertExpectations(t)
}

func TestGetOrderHandler_BadRequest_EmptyUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(OrderServiceMock)

	router := gin.Default()
	server := NewHTTPServer(router, mockService, logrus.NewEntry(logrus.New()))
	router.GET("/order/:order_uid", server.getOrderHandler)

	req := httptest.NewRequest(http.MethodGet, "/order/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
