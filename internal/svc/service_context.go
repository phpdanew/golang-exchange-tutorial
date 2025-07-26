package svc

import (
	"crypto-exchange/internal/config"
	"crypto-exchange/internal/matching"
	"crypto-exchange/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	// Import PostgreSQL driver
	_ "github.com/lib/pq"
)

type ServiceContext struct {
	Config                 config.Config
	UserModel              model.UserModel
	BalanceModel           model.BalanceModel
	AssetTransactionModel  model.AssetTransactionModel
	OrderModel             model.OrderModel
	TradeModel             model.TradeModel
	TradingPairModel       model.TradingPairModel
	TickerModel            model.TickerModel
	KlineModel             model.KlineModel
	RedisClient            *redis.Redis
	MatchingEngine         matching.Engine  // 使用接口避免循环引用
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("postgres", c.DataSource)
	return &ServiceContext{
		Config:                 c,
		UserModel:              model.NewUserModel(conn),
		BalanceModel:           model.NewBalanceModel(conn),
		AssetTransactionModel:  model.NewAssetTransactionModel(conn),
		OrderModel:             model.NewOrderModel(conn),
		TradeModel:             model.NewTradeModel(conn),
		TradingPairModel:       model.NewTradingPairModel(conn),
		TickerModel:            model.NewTickerModel(conn),
		KlineModel:             model.NewKlineModel(conn),
		RedisClient:            redis.MustNewRedis(c.Redis),
		MatchingEngine:         matching.NewMatchingEngine(),  // 初始化撮合引擎
	}
}
