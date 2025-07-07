package cache

import (
	"L0/internal/domain"
	"context"
	"encoding/json"
	"github.com/go-redis/redismock/v9"
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

func TestRedisCache_Cache_Hit(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()
	mockRepo := new(MockRepository)

	now := time.Now()
	testUID := "test_order_uid_1234"
	expectedKey := createKey(testUID)
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

	expectedJSON, err := json.Marshal(expectedOrder)
	assert.NoError(t, err)

	redisMock.ExpectGet(expectedKey).SetVal(string(expectedJSON))

	logger := logrus.NewEntry(logrus.New())
	redisCache := NewRedisCache(mockRepo, redisClient, logger)

	resultOrder, err := redisCache.GetByUID(context.Background(), testUID)

	assert.NoError(t, err)
	assert.NotNil(t, resultOrder)
	assert.Equal(t, expectedOrder.OrderUID, resultOrder.OrderUID)

	assert.NoError(t, redisMock.ExpectationsWereMet())

	mockRepo.AssertNotCalled(t, "GetByUID", mock.Anything, mock.Anything)
}

func TestRedisCache_GetByUID_CacheMiss(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()
	mockRepo := new(MockRepository)

	testUID := "order_miss_id_5678"
	expectedKey := createKey(testUID)
	expectedOrder := &domain.Order{OrderUID: testUID, TrackNumber: "TRACK456"}

	expectedJSON, err := json.Marshal(expectedOrder)
	assert.NoError(t, err)

	redisMock.ExpectGet(expectedKey).RedisNil()

	mockRepo.On("GetByUID", mock.Anything, testUID).Return(expectedOrder, nil)

	redisMock.ExpectSet(expectedKey, expectedJSON, 15*time.Minute).SetVal("OK")

	logger := logrus.NewEntry(logrus.New())
	redisCache := NewRedisCache(mockRepo, redisClient, logger)

	resultOrder, err := redisCache.GetByUID(context.Background(), testUID)

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, resultOrder)

	assert.NoError(t, redisMock.ExpectationsWereMet())
	mockRepo.AssertExpectations(t)
}
