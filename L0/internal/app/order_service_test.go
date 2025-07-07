package app

import (
	"L0/internal/domain"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) GetByUID(ctx context.Context, uid string) (*domain.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockRepository) GetLatest(ctx context.Context, limit uint) ([]*domain.Order, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockCache) Get(ctx context.Context, uid string) (*domain.Order, bool, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*domain.Order), args.Bool(1), args.Error(2)
}

func TestOrderService_ProcessNewOrder_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	now := time.Now()
	testUID := "test_order_uid_1234"
	testOrder := &domain.Order{
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

	mockRepo.On("Save", mock.Anything, testOrder).Return(nil)

	logger := logrus.New().WithFields(logrus.Fields{})
	service := NewOrderService(mockRepo, nil, logger)

	err := service.ProcessNewOrder(context.Background(), testOrder)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_ProcessNewOrder_Failure(t *testing.T) {
	mockRepo := new(MockRepository)

	testOrder := &domain.Order{OrderUID: "error_order_uid_123"}
	expectedError := errors.New("database error")

	mockRepo.On("Save", mock.Anything, testOrder).Return(expectedError)

	logger := logrus.NewEntry(logrus.New())
	service := NewOrderService(mockRepo, nil, logger)

	err := service.ProcessNewOrder(context.Background(), testOrder)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_WarmUpCache(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	testOrders := []*domain.Order{
		{OrderUID: "test_order_for_cache1"},
		{OrderUID: "test_order_for_cache2"},
	}

	mockRepo.On("GetLatest", mock.Anything, uint(2)).Return(testOrders, nil)
	mockCache.On("Set", mock.Anything, testOrders[0]).Return(nil)
	mockCache.On("Set", mock.Anything, testOrders[1]).Return(nil)

	logger := logrus.NewEntry(logrus.New())
	service := NewOrderService(mockRepo, mockCache, logger)

	err := service.WarmUpCache(context.Background(), 2)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
