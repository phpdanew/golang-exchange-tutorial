package market

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"crypto-exchange/internal/config"
	"crypto-exchange/internal/svc"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// MockSqlResult 模拟SQL结果
type MockSqlResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (m *MockSqlResult) LastInsertId() (int64, error) {
	return m.lastInsertId, nil
}

func (m *MockSqlResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

// MockTradingPairModel 模拟交易对模型
type MockTradingPairModel struct {
	mock.Mock
}

func (m *MockTradingPairModel) Insert(ctx context.Context, data *model.TradingPair) (sql.Result, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sql.Result), args.Error(1)
}

func (m *MockTradingPairModel) FindOne(ctx context.Context, id uint64) (*model.TradingPair, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.TradingPair), args.Error(1)
}

func (m *MockTradingPairModel) Update(ctx context.Context, data *model.TradingPair) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockTradingPairModel) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTradingPairModel) FindBySymbol(ctx context.Context, symbol string) (*model.TradingPair, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TradingPair), args.Error(1)
}

func (m *MockTradingPairModel) FindByStatus(ctx context.Context, status int64) ([]*model.TradingPair, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*model.TradingPair), args.Error(1)
}

func (m *MockTradingPairModel) FindActivePairs(ctx context.Context) ([]*model.TradingPair, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.TradingPair), args.Error(1)
}

func createTestServiceContext() *svc.ServiceContext {
	return &svc.ServiceContext{
		Config: config.Config{
			Redis: redis.RedisConf{
				Host: "localhost:6379",
				Type: "node",
			},
		},
	}
}

func TestTradingPairManager_GetTradingPairBySymbol(t *testing.T) {
	ctx := context.Background()
	svcCtx := createTestServiceContext()
	mockModel := &MockTradingPairModel{}
	svcCtx.TradingPairModel = mockModel

	manager := NewTradingPairManager(ctx, svcCtx)

	tests := []struct {
		name      string
		symbol    string
		mockSetup func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "valid symbol found",
			symbol: "BTC/USDT",
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(&model.TradingPair{
					ID:            1,
					Symbol:        "BTC/USDT",
					BaseCurrency:  "BTC",
					QuoteCurrency: "USDT",
					MinAmount:     "0.00001",
					MaxAmount:     "1000",
					PriceScale:    2,
					AmountScale:   8,
					Status:        1,
					CreatedAt:     time.Now(),
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "invalid symbol format",
			symbol: "BTCUSDT",
			mockSetup: func() {
				// No mock setup needed as validation fails first
			},
			wantErr: true,
			errMsg:  "invalid trading pair symbol format",
		},
		{
			name:   "symbol not found",
			symbol: "ETH/BTC",
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "ETH/BTC").Return(nil, model.ErrNotFound)
			},
			wantErr: true,
			errMsg:  "trading pair not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockModel.ExpectedCalls = nil
			mockModel.Calls = nil

			tt.mockSetup()

			pair, err := manager.GetTradingPairBySymbol(tt.symbol)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, pair)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pair)
				assert.Equal(t, tt.symbol, pair.Symbol)
			}

			mockModel.AssertExpectations(t)
		})
	}
}

func TestTradingPairManager_ValidateOrderForTradingPair(t *testing.T) {
	ctx := context.Background()
	svcCtx := createTestServiceContext()
	mockModel := &MockTradingPairModel{}
	svcCtx.TradingPairModel = mockModel

	manager := NewTradingPairManager(ctx, svcCtx)

	// Setup mock trading pair
	testPair := &model.TradingPair{
		ID:            1,
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		MinAmount:     "0.00001",
		MaxAmount:     "1000",
		PriceScale:    2,
		AmountScale:   8,
		Status:        1,
		CreatedAt:     time.Now(),
	}

	tests := []struct {
		name      string
		symbol    string
		amount    string
		price     string
		orderType int64
		mockSetup func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid limit order",
			symbol:    "BTC/USDT",
			amount:    "0.1",
			price:     "50000.12",
			orderType: 1, // limit order
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(testPair, nil)
			},
			wantErr: false,
		},
		{
			name:      "valid market order",
			symbol:    "BTC/USDT",
			amount:    "0.1",
			price:     "",
			orderType: 2, // market order
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(testPair, nil)
			},
			wantErr: false,
		},
		{
			name:      "amount below minimum",
			symbol:    "BTC/USDT",
			amount:    "0.000001",
			price:     "50000.12",
			orderType: 1,
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(testPair, nil)
			},
			wantErr: true,
			errMsg:  "is less than minimum",
		},
		{
			name:      "amount above maximum",
			symbol:    "BTC/USDT",
			amount:    "2000",
			price:     "50000.12",
			orderType: 1,
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(testPair, nil)
			},
			wantErr: true,
			errMsg:  "exceeds maximum",
		},
		{
			name:      "price precision too high",
			symbol:    "BTC/USDT",
			amount:    "0.1",
			price:     "50000.123",
			orderType: 1,
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(testPair, nil)
			},
			wantErr: true,
			errMsg:  "precision exceeds 2 decimal places",
		},
		{
			name:      "disabled trading pair",
			symbol:    "BTC/USDT",
			amount:    "0.1",
			price:     "50000.12",
			orderType: 1,
			mockSetup: func() {
				disabledPair := *testPair
				disabledPair.Status = 2 // disabled
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(&disabledPair, nil)
			},
			wantErr: true,
			errMsg:  "is currently disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockModel.ExpectedCalls = nil
			mockModel.Calls = nil

			tt.mockSetup()

			err := manager.ValidateOrderForTradingPair(tt.symbol, tt.amount, tt.price, tt.orderType)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockModel.AssertExpectations(t)
		})
	}
}

func TestTradingPairManager_UpdateTradingPairStatus(t *testing.T) {
	ctx := context.Background()
	svcCtx := createTestServiceContext()
	mockModel := &MockTradingPairModel{}
	svcCtx.TradingPairModel = mockModel

	manager := NewTradingPairManager(ctx, svcCtx)

	testPair := &model.TradingPair{
		ID:            1,
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		MinAmount:     "0.00001",
		MaxAmount:     "1000",
		PriceScale:    2,
		AmountScale:   8,
		Status:        1,
		CreatedAt:     time.Now(),
	}

	tests := []struct {
		name      string
		symbol    string
		status    int64
		mockSetup func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "valid status update to disabled",
			symbol: "BTC/USDT",
			status: 2,
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(testPair, nil)
				updatedPair := *testPair
				updatedPair.Status = 2
				mockModel.On("Update", ctx, &updatedPair).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "invalid status",
			symbol: "BTC/USDT",
			status: 3,
			mockSetup: func() {
				// No mock setup needed as validation fails first
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name:   "trading pair not found",
			symbol: "ETH/BTC",
			status: 2,
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "ETH/BTC").Return(nil, model.ErrNotFound)
			},
			wantErr: true,
			errMsg:  "trading pair not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockModel.ExpectedCalls = nil
			mockModel.Calls = nil

			tt.mockSetup()

			err := manager.UpdateTradingPairStatus(tt.symbol, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockModel.AssertExpectations(t)
		})
	}
}

func TestTradingPairManager_GetActiveTradingPairs(t *testing.T) {
	ctx := context.Background()
	svcCtx := createTestServiceContext()
	mockModel := &MockTradingPairModel{}
	svcCtx.TradingPairModel = mockModel

	manager := NewTradingPairManager(ctx, svcCtx)

	activePairs := []*model.TradingPair{
		{
			ID:            1,
			Symbol:        "BTC/USDT",
			BaseCurrency:  "BTC",
			QuoteCurrency: "USDT",
			Status:        1,
		},
		{
			ID:            2,
			Symbol:        "ETH/USDT",
			BaseCurrency:  "ETH",
			QuoteCurrency: "USDT",
			Status:        1,
		},
	}

	mockModel.On("FindActivePairs", ctx).Return(activePairs, nil)

	pairs, err := manager.GetActiveTradingPairs()

	assert.NoError(t, err)
	assert.Len(t, pairs, 2)
	assert.Equal(t, "BTC/USDT", pairs[0].Symbol)
	assert.Equal(t, "ETH/USDT", pairs[1].Symbol)

	mockModel.AssertExpectations(t)
}

func TestTradingPairManager_CreateTradingPair(t *testing.T) {
	ctx := context.Background()
	svcCtx := createTestServiceContext()
	mockModel := &MockTradingPairModel{}
	svcCtx.TradingPairModel = mockModel

	manager := NewTradingPairManager(ctx, svcCtx)

	tests := []struct {
		name      string
		pair      *model.TradingPair
		mockSetup func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid new trading pair",
			pair: &model.TradingPair{
				Symbol:        "LTC/USDT",
				BaseCurrency:  "LTC",
				QuoteCurrency: "USDT",
				MinAmount:     "0.01",
				MaxAmount:     "10000",
				PriceScale:    2,
				AmountScale:   4,
				Status:        1,
			},
			mockSetup: func() {
				mockModel.On("FindBySymbol", ctx, "LTC/USDT").Return(nil, model.ErrNotFound)
				mockModel.On("Insert", ctx, mock.AnythingOfType("*model.TradingPair")).Return(&MockSqlResult{lastInsertId: 1, rowsAffected: 1}, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid symbol format",
			pair: &model.TradingPair{
				Symbol:        "LTCUSDT",
				BaseCurrency:  "LTC",
				QuoteCurrency: "USDT",
				MinAmount:     "0.01",
				MaxAmount:     "10000",
				PriceScale:    2,
				AmountScale:   4,
				Status:        1,
			},
			mockSetup: func() {
				// No mock setup needed as validation fails first
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "trading pair already exists",
			pair: &model.TradingPair{
				Symbol:        "BTC/USDT",
				BaseCurrency:  "BTC",
				QuoteCurrency: "USDT",
				MinAmount:     "0.00001",
				MaxAmount:     "1000",
				PriceScale:    2,
				AmountScale:   8,
				Status:        1,
			},
			mockSetup: func() {
				existingPair := &model.TradingPair{ID: 1, Symbol: "BTC/USDT"}
				mockModel.On("FindBySymbol", ctx, "BTC/USDT").Return(existingPair, nil)
			},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockModel.ExpectedCalls = nil
			mockModel.Calls = nil

			tt.mockSetup()

			err := manager.CreateTradingPair(tt.pair)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Verify that CreatedAt was set
				assert.False(t, tt.pair.CreatedAt.IsZero())
			}

			mockModel.AssertExpectations(t)
		})
	}
}