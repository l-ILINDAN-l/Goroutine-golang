package tkafka

import (
	"L0/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type OrderServiceMock struct {
	mock.Mock
}

func (m *OrderServiceMock) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *OrderServiceMock) GetOrderByUID(ctx context.Context, uid string) (*domain.Order, error) {
	return nil, nil
}
func (m *OrderServiceMock) WarmUpCache(ctx context.Context, limit uint) error { return nil }

func TestConsumer_HandleSingleMessage(t *testing.T) {
	validOrder := &domain.Order{
		OrderUID: "valid_order_uid_123",
		Delivery: &domain.Delivery{},
		Payment:  &domain.Payment{},
	}
	validOrderJSON, _ := json.Marshal(validOrder)

	orderWithNilDelivery := &domain.Order{
		OrderUID: "invalid_order_uid_456",
		Delivery: nil,
		Payment:  &domain.Payment{},
	}
	invalidOrderJSON, _ := json.Marshal(orderWithNilDelivery)

	testCases := []struct {
		name          string
		message       *kafka.Message
		setupMock     func(mockService *OrderServiceMock)
		expectError   bool
		expectedError error
	}{
		{
			name:    "Success Case",
			message: &kafka.Message{Value: validOrderJSON},
			setupMock: func(mockService *OrderServiceMock) {
				mockService.On("ProcessNewOrder", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)
			},
			expectError: false,
		},
		{
			name:    "Unmarshal Error Case",
			message: &kafka.Message{Value: []byte("this is not json")},
			setupMock: func(mockService *OrderServiceMock) {
			},
			expectError: true,
		},
		{
			name:    "Invalid Data Case (nil delivery)",
			message: &kafka.Message{Value: invalidOrderJSON},
			setupMock: func(mockService *OrderServiceMock) {
			},
			expectError:   true,
			expectedError: ErrInvalidMessageData,
		},
		{
			name:    "Service Error Case",
			message: &kafka.Message{Value: validOrderJSON},
			setupMock: func(mockService *OrderServiceMock) {
				mockService.On("ProcessNewOrder", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(errors.New("database is down"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(OrderServiceMock)
			tc.setupMock(mockService)

			logger := logrus.NewEntry(logrus.New())
			consumer := &Consumer{
				orderService: mockService,
				logger:       logger,
			}
			err := consumer.handleSingleMessage(context.Background(), tc.message)

			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)
			}
			mockService.AssertExpectations(t)
		})
	}
}
