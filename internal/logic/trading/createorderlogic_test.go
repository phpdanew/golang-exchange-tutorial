package trading

import (
	"context"
	"database/sql"
	"testing"

	"crypto-exchange/internal/matching"
	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Mock implementations for testing
type mockOrderModel struct {
	mock.Mock
}

func (m *mockOrderModel) Insert(ctx context.Context, data *model.Order) (sql.Result, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(sql.Result), args.Error(1)
}

func (m *mockOrderModel) FindOne(ctx context.Context, id uint64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderModel) Update(ctx context.Context, data *model.Order) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *mockOrderModel) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockOrderModel) FindByUserID(ctx context.Context, userID uint64) ([]*model.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderModel) FindByStatus(ctx context.Context, status int64) ([]*model.Order, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderModel) FindBySymbol(ctx context.Context, symbol string) ([]*model.Order, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderModel) FindByUserIDAndStatus(ctx context.Context, userID uint64, status int64) ([]*model.Order, error) {
	args := m.Called(ctx, userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderModel) FindBySymbolAndStatus(ctx context.Context, symbol string, status int64) ([]*model.Order, error) {
	args := m.Called(ctx, symbol, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderModel) FindBySymbolAndSideAndStatus(ctx context.Context, symbol string, side int64, status int64) ([]*model.Order, error) {
	args := m.Called(ctx, symbol, side, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderModel) FindByUserIDWithPagination(ctx context.Context, userID uint64, symbol string, status int64, page, size int64) ([]*model.Order, int64, error) {
	args := m.Called(ctx, userID, symbol, status, page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *mockOrderModel) UpdateStatus(ctx context.Context, id uint64, status int64) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockOrderModel) UpdateFilledAmount(ctx context.Context, id uint64, filledAmount string) error {
	args := m.Called(ctx, id, filledAmount)
	return args.Error(0)
}

func (m *mockOrderModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	m.Called(ctx, fn)
	// 执行事务函数进行测试
	return fn(ctx, nil)
}

// Mock撮合引擎
type mockMatchingEngine struct {
	mock.Mock
}

func (m *mockMatchingEngine) ProcessOrder(order *model.Order) (*matching.MatchResult, error) {
	args := m.Called(order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matching.MatchResult), args.Error(1)
}

func (m *mockMatchingEngine) CancelOrder(order *model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *mockMatchingEngine) GetMarketDepth(symbol string, depth int) ([]matching.PriceLevel, []matching.PriceLevel) {
	args := m.Called(symbol, depth)
	return args.Get(0).([]matching.PriceLevel), args.Get(1).([]matching.PriceLevel)
}

func (m *mockMatchingEngine) GetOrderBookSnapshot(symbol string) map[string]interface{} {
	args := m.Called(symbol)
	return args.Get(0).(map[string]interface{})
}

type mockTradingPairModel struct {
	mock.Mock
}

func (m *mockTradingPairModel) FindBySymbol(ctx context.Context, symbol string) (*model.TradingPair, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TradingPair), args.Error(1)
}

func (m *mockTradingPairModel) Insert(ctx context.Context, data *model.TradingPair) (sql.Result, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(sql.Result), args.Error(1)
}

func (m *mockTradingPairModel) FindOne(ctx context.Context, id uint64) (*model.TradingPair, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TradingPair), args.Error(1)
}

func (m *mockTradingPairModel) Update(ctx context.Context, data *model.TradingPair) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *mockTradingPairModel) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTradingPairModel) FindByStatus(ctx context.Context, status int64) ([]*model.TradingPair, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*model.TradingPair), args.Error(1)
}

func (m *mockTradingPairModel) FindActivePairs(ctx context.Context) ([]*model.TradingPair, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.TradingPair), args.Error(1)
}

type mockBalanceModel struct {
	mock.Mock
}

func (m *mockBalanceModel) Insert(ctx context.Context, data *model.Balance) (sql.Result, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(sql.Result), args.Error(1)
}

func (m *mockBalanceModel) FindOne(ctx context.Context, id uint64) (*model.Balance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Balance), args.Error(1)
}

func (m *mockBalanceModel) Update(ctx context.Context, data *model.Balance) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *mockBalanceModel) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockBalanceModel) FindByUserID(ctx context.Context, userID uint64) ([]*model.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.Balance), args.Error(1)
}

func (m *mockBalanceModel) FindByUserIDAndCurrency(ctx context.Context, userID uint64, currency string) (*model.Balance, error) {
	args := m.Called(ctx, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Balance), args.Error(1)
}

func (m *mockBalanceModel) UpdateBalance(ctx context.Context, userID uint64, currency string, available, frozen string) error {
	args := m.Called(ctx, userID, currency, available, frozen)
	return args.Error(0)
}

func (m *mockBalanceModel) FreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	args := m.Called(ctx, userID, currency, amount)
	return args.Error(0)
}

func (m *mockBalanceModel) UnfreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	args := m.Called(ctx, userID, currency, amount)
	return args.Error(0)
}

func (m *mockBalanceModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	_ = m.Called(ctx, fn)
	// 执行事务函数进行测试
	return fn(ctx, nil)
}

type mockSqlResult struct {
	lastInsertId int64
}

func (m *mockSqlResult) LastInsertId() (int64, error) {
	return m.lastInsertId, nil
}

func (m *mockSqlResult) RowsAffected() (int64, error) {
	return 1, nil
}

func TestCreateOrderLogic_CreateOrder_Success(t *testing.T) {
	// 准备测试数据
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}
	mockMatchingEngine := &mockMatchingEngine{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
		MatchingEngine:   mockMatchingEngine,
	}

	logic := &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期
	tradingPair := &model.TradingPair{
		ID:            1,
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		MinAmount:     "0.001",
		MaxAmount:     "1000",
		PriceScale:    2,
		AmountScale:   8,
		Status:        1,
	}

	// Mock撮合引擎返回结果（无成交）
	matchResult := &matching.MatchResult{
		Trades:        []*model.Trade{},
		UpdatedOrders: []*model.Order{},
		FilledOrders:  []*model.Order{},
	}

	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockBalanceModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)
	mockBalanceModel.On("FreezeBalance", mock.Anything, uint64(1), "USDT", "50000").Return(nil)
	mockOrderModel.On("Insert", mock.Anything, mock.AnythingOfType("*model.Order")).Return(&mockSqlResult{lastInsertId: 123}, nil)
	mockOrderModel.On("Update", mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
	mockMatchingEngine.On("ProcessOrder", mock.AnythingOfType("*model.Order")).Return(matchResult, nil)

	// 测试请求
	req := &types.CreateOrderRequest{
		Symbol: "BTC/USDT",
		Type:   1, // 限价单
		Side:   1, // 买入
		Amount: "1.00000000",
		Price:  "50000.00",
	}

	// 执行测试
	resp, err := logic.CreateOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(123), resp.ID)
	assert.Equal(t, uint64(1), resp.UserID)
	assert.Equal(t, "BTC/USDT", resp.Symbol)
	assert.Equal(t, int64(1), resp.Type)
	assert.Equal(t, int64(1), resp.Side)
	assert.Equal(t, "1.00000000", resp.Amount)
	assert.Equal(t, "50000.00", resp.Price)
	assert.Equal(t, "0", resp.FilledAmount)
	assert.Equal(t, int64(1), resp.Status)

	// 验证mock调用
	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
	mockOrderModel.AssertExpectations(t)
	mockMatchingEngine.AssertExpectations(t)
}

func TestCreateOrderLogic_CreateOrder_TradingPairNotFound(t *testing.T) {
	mockTradingPairModel := &mockTradingPairModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		TradingPairModel: mockTradingPairModel,
	}

	logic := &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 交易对不存在
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "INVALID/USDT").Return(nil, model.ErrNotFound)

	req := &types.CreateOrderRequest{
		Symbol: "INVALID/USDT",
		Type:   1,
		Side:   1,
		Amount: "1.00000000",
		Price:  "50000.00",
	}

	// 执行测试
	resp, err := logic.CreateOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrTradingPairNotFound, err)

	mockTradingPairModel.AssertExpectations(t)
}

func TestCreateOrderLogic_CreateOrder_TradingPairDisabled(t *testing.T) {
	mockTradingPairModel := &mockTradingPairModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		TradingPairModel: mockTradingPairModel,
	}

	logic := &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 交易对被禁用
	tradingPair := &model.TradingPair{
		Symbol: "BTC/USDT",
		Status: 2, // 禁用状态
	}
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)

	req := &types.CreateOrderRequest{
		Symbol: "BTC/USDT",
		Type:   1,
		Side:   1,
		Amount: "1.00000000",
		Price:  "50000.00",
	}

	// 执行测试
	resp, err := logic.CreateOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrTradingPairDisabled, err)

	mockTradingPairModel.AssertExpectations(t)
}

func TestCreateOrderLogic_CreateOrder_InvalidAmount(t *testing.T) {
	mockTradingPairModel := &mockTradingPairModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		TradingPairModel: mockTradingPairModel,
	}

	logic := &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		MinAmount:     "0.001",
		MaxAmount:     "1000",
		PriceScale:    2,
		AmountScale:   8,
		Status:        1,
	}
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)

	req := &types.CreateOrderRequest{
		Symbol: "BTC/USDT",
		Type:   1,
		Side:   1,
		Amount: "invalid_amount", // 无效金额
		Price:  "50000.00",
	}

	// 执行测试
	resp, err := logic.CreateOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrInvalidAmount, err)

	mockTradingPairModel.AssertExpectations(t)
}

func TestCreateOrderLogic_CreateOrder_InsufficientBalance(t *testing.T) {
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	logic := &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		MinAmount:     "0.001",
		MaxAmount:     "1000",
		PriceScale:    2,
		AmountScale:   8,
		Status:        1,
	}

	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockBalanceModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(model.ErrInsufficientBalance)
	mockBalanceModel.On("FreezeBalance", mock.Anything, uint64(1), "USDT", "50000").Return(model.ErrInsufficientBalance)

	req := &types.CreateOrderRequest{
		Symbol: "BTC/USDT",
		Type:   1,
		Side:   1,
		Amount: "1.00000000",
		Price:  "50000.00",
	}

	// 执行测试
	resp, err := logic.CreateOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrInsufficientBalance, err)

	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
}

func TestCreateOrderLogic_CreateOrder_MarketBuyOrder(t *testing.T) {
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	logic := &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		MinAmount:     "0.001",
		MaxAmount:     "1000",
		PriceScale:    2,
		AmountScale:   8,
		Status:        1,
	}

	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockBalanceModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)
	mockBalanceModel.On("FreezeBalance", mock.Anything, uint64(1), "USDT", "1000").Return(nil)
	mockOrderModel.On("Insert", mock.Anything, mock.AnythingOfType("*model.Order")).Return(&mockSqlResult{lastInsertId: 124}, nil)

	// 市价买单 - Amount表示要花费的USDT数量
	req := &types.CreateOrderRequest{
		Symbol: "BTC/USDT",
		Type:   2, // 市价单
		Side:   1, // 买入
		Amount: "1000", // 花费1000 USDT
		Price:  "",     // 市价单无需价格
	}

	// 执行测试
	resp, err := logic.CreateOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(124), resp.ID)
	assert.Equal(t, int64(2), resp.Type)
	assert.Equal(t, int64(1), resp.Side)
	assert.Equal(t, "1000", resp.Amount)

	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
	mockOrderModel.AssertExpectations(t)
}

func TestCreateOrderLogic_CreateOrder_SellOrder(t *testing.T) {
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	logic := &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		MinAmount:     "0.001",
		MaxAmount:     "1000",
		PriceScale:    2,
		AmountScale:   8,
		Status:        1,
	}

	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockBalanceModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)
	mockBalanceModel.On("FreezeBalance", mock.Anything, uint64(1), "BTC", "1").Return(nil)
	mockOrderModel.On("Insert", mock.Anything, mock.AnythingOfType("*model.Order")).Return(&mockSqlResult{lastInsertId: 125}, nil)

	// 卖出订单 - 冻结基础币种
	req := &types.CreateOrderRequest{
		Symbol: "BTC/USDT",
		Type:   1, // 限价单
		Side:   2, // 卖出
		Amount: "1.00000000",
		Price:  "50000.00",
	}

	// 执行测试
	resp, err := logic.CreateOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(125), resp.ID)
	assert.Equal(t, int64(2), resp.Side)

	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
	mockOrderModel.AssertExpectations(t)
}

