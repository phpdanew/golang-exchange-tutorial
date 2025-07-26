-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS exchange;

-- 使用数据库
\c exchange;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,                                    -- 用户ID，主键
    email VARCHAR(255) UNIQUE NOT NULL,                       -- 用户邮箱，唯一索引
    password VARCHAR(255) NOT NULL,                           -- 密码哈希值
    nickname VARCHAR(100),                                    -- 用户昵称
    status INTEGER DEFAULT 1,                                 -- 用户状态：1-正常，2-禁用，3-删除
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,           -- 创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP            -- 更新时间
);

COMMENT ON TABLE users IS '用户基础信息表';
COMMENT ON COLUMN users.id IS '用户ID，主键';
COMMENT ON COLUMN users.email IS '用户邮箱，唯一标识';
COMMENT ON COLUMN users.password IS '密码哈希值，使用bcrypt加密';
COMMENT ON COLUMN users.nickname IS '用户昵称，显示名称';
COMMENT ON COLUMN users.status IS '用户状态：1-正常，2-禁用，3-删除';
COMMENT ON COLUMN users.created_at IS '账户创建时间';
COMMENT ON COLUMN users.updated_at IS '最后更新时间';

-- 用户余额表
CREATE TABLE IF NOT EXISTS balances (
    id SERIAL PRIMARY KEY,                                    -- 余额记录ID
    user_id INTEGER REFERENCES users(id),                     -- 用户ID，外键
    currency VARCHAR(20) NOT NULL,                            -- 币种代码，如BTC、ETH、USDT
    available VARCHAR(50) DEFAULT '0',                        -- 可用余额，使用string存储避免精度问题
    frozen VARCHAR(50) DEFAULT '0',                           -- 冻结余额，用于挂单
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,           -- 最后更新时间
    UNIQUE(user_id, currency)                                 -- 用户+币种唯一索引
);

COMMENT ON TABLE balances IS '用户资产余额表';
COMMENT ON COLUMN balances.id IS '余额记录ID，主键';
COMMENT ON COLUMN balances.user_id IS '用户ID，关联users表';
COMMENT ON COLUMN balances.currency IS '币种代码，如BTC、ETH、USDT等';
COMMENT ON COLUMN balances.available IS '可用余额，可以用于交易的金额';
COMMENT ON COLUMN balances.frozen IS '冻结余额，挂单时冻结的金额';
COMMENT ON COLUMN balances.updated_at IS '余额最后更新时间';

-- 交易对表
CREATE TABLE IF NOT EXISTS trading_pairs (
    id SERIAL PRIMARY KEY,                                    -- 交易对ID
    symbol VARCHAR(20) UNIQUE NOT NULL,                       -- 交易对符号，如BTC/USDT
    base_currency VARCHAR(10) NOT NULL,                       -- 基础币种，如BTC
    quote_currency VARCHAR(10) NOT NULL,                      -- 计价币种，如USDT
    min_amount VARCHAR(50) DEFAULT '0',                       -- 最小交易数量
    max_amount VARCHAR(50) DEFAULT '0',                       -- 最大交易数量
    price_scale INTEGER DEFAULT 8,                            -- 价格精度，小数位数
    amount_scale INTEGER DEFAULT 8,                           -- 数量精度，小数位数
    status INTEGER DEFAULT 1,                                 -- 交易对状态：1-正常，2-禁用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP            -- 创建时间
);

COMMENT ON TABLE trading_pairs IS '交易对配置表';
COMMENT ON COLUMN trading_pairs.id IS '交易对ID，主键';
COMMENT ON COLUMN trading_pairs.symbol IS '交易对符号，格式：基础币种/计价币种，如BTC/USDT';
COMMENT ON COLUMN trading_pairs.base_currency IS '基础币种代码，交易的目标币种';
COMMENT ON COLUMN trading_pairs.quote_currency IS '计价币种代码，用于定价的币种';
COMMENT ON COLUMN trading_pairs.min_amount IS '单笔交易最小数量限制';
COMMENT ON COLUMN trading_pairs.max_amount IS '单笔交易最大数量限制';
COMMENT ON COLUMN trading_pairs.price_scale IS '价格显示精度，小数点后位数';
COMMENT ON COLUMN trading_pairs.amount_scale IS '数量显示精度，小数点后位数';
COMMENT ON COLUMN trading_pairs.status IS '交易对状态：1-正常交易，2-禁用交易';
COMMENT ON COLUMN trading_pairs.created_at IS '交易对创建时间';

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,                                    -- 订单ID
    user_id INTEGER REFERENCES users(id),                     -- 用户ID
    symbol VARCHAR(20) NOT NULL,                              -- 交易对符号
    type INTEGER NOT NULL,                                    -- 订单类型：1-限价单，2-市价单
    side INTEGER NOT NULL,                                    -- 交易方向：1-买入，2-卖出
    amount VARCHAR(50) NOT NULL,                              -- 订单数量
    price VARCHAR(50),                                        -- 订单价格（市价单为NULL）
    filled_amount VARCHAR(50) DEFAULT '0',                    -- 已成交数量
    status INTEGER DEFAULT 1,                                 -- 订单状态：1-待成交，2-部分成交，3-完全成交，4-已取消
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,           -- 创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP            -- 更新时间
);

COMMENT ON TABLE orders IS '交易订单表';
COMMENT ON COLUMN orders.id IS '订单ID，主键';
COMMENT ON COLUMN orders.user_id IS '下单用户ID，关联users表';
COMMENT ON COLUMN orders.symbol IS '交易对符号，如BTC/USDT';
COMMENT ON COLUMN orders.type IS '订单类型：1-限价单（指定价格），2-市价单（按市场价格）';
COMMENT ON COLUMN orders.side IS '交易方向：1-买入（买入基础币种），2-卖出（卖出基础币种）';
COMMENT ON COLUMN orders.amount IS '订单总数量，基础币种数量';
COMMENT ON COLUMN orders.price IS '订单价格，限价单必填，市价单为NULL';
COMMENT ON COLUMN orders.filled_amount IS '已成交数量，累计成交的基础币种数量';
COMMENT ON COLUMN orders.status IS '订单状态：1-待成交，2-部分成交，3-完全成交，4-已取消';
COMMENT ON COLUMN orders.created_at IS '订单创建时间';
COMMENT ON COLUMN orders.updated_at IS '订单最后更新时间';

-- 成交记录表
CREATE TABLE IF NOT EXISTS trades (
    id SERIAL PRIMARY KEY,                                    -- 成交记录ID
    symbol VARCHAR(20) NOT NULL,                              -- 交易对符号
    buy_order_id INTEGER REFERENCES orders(id),               -- 买单ID
    sell_order_id INTEGER REFERENCES orders(id),              -- 卖单ID
    buy_user_id INTEGER REFERENCES users(id),                 -- 买方用户ID
    sell_user_id INTEGER REFERENCES users(id),                -- 卖方用户ID
    price VARCHAR(50) NOT NULL,                               -- 成交价格
    amount VARCHAR(50) NOT NULL,                              -- 成交数量
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP            -- 成交时间
);

COMMENT ON TABLE trades IS '交易成交记录表';
COMMENT ON COLUMN trades.id IS '成交记录ID，主键';
COMMENT ON COLUMN trades.symbol IS '交易对符号，如BTC/USDT';
COMMENT ON COLUMN trades.buy_order_id IS '买单订单ID，关联orders表';
COMMENT ON COLUMN trades.sell_order_id IS '卖单订单ID，关联orders表';
COMMENT ON COLUMN trades.buy_user_id IS '买方用户ID，关联users表';
COMMENT ON COLUMN trades.sell_user_id IS '卖方用户ID，关联users表';
COMMENT ON COLUMN trades.price IS '成交价格，以计价币种计价';
COMMENT ON COLUMN trades.amount IS '成交数量，基础币种数量';
COMMENT ON COLUMN trades.created_at IS '成交时间戳';

-- K线数据表
CREATE TABLE IF NOT EXISTS klines (
    id SERIAL PRIMARY KEY,                                    -- K线记录ID
    symbol VARCHAR(20) NOT NULL,                              -- 交易对符号
    interval VARCHAR(10) NOT NULL,                            -- 时间周期：1m,5m,15m,1h,4h,1d等
    open_time TIMESTAMP NOT NULL,                             -- K线开始时间
    close_time TIMESTAMP NOT NULL,                            -- K线结束时间
    open VARCHAR(50) NOT NULL,                                -- 开盘价
    high VARCHAR(50) NOT NULL,                                -- 最高价
    low VARCHAR(50) NOT NULL,                                 -- 最低价
    close VARCHAR(50) NOT NULL,                               -- 收盘价
    volume VARCHAR(50) NOT NULL,                              -- 成交量（基础币种）
    UNIQUE(symbol, interval, open_time)                       -- 唯一索引：交易对+周期+开始时间
);

COMMENT ON TABLE klines IS 'K线数据表，存储各时间周期的OHLCV数据';
COMMENT ON COLUMN klines.id IS 'K线记录ID，主键';
COMMENT ON COLUMN klines.symbol IS '交易对符号，如BTC/USDT';
COMMENT ON COLUMN klines.interval IS '时间周期，如1m(1分钟)、5m(5分钟)、1h(1小时)、1d(1天)等';
COMMENT ON COLUMN klines.open_time IS 'K线周期开始时间';
COMMENT ON COLUMN klines.close_time IS 'K线周期结束时间';
COMMENT ON COLUMN klines.open IS '开盘价，周期内第一笔成交价格';
COMMENT ON COLUMN klines.high IS '最高价，周期内最高成交价格';
COMMENT ON COLUMN klines.low IS '最低价，周期内最低成交价格';
COMMENT ON COLUMN klines.close IS '收盘价，周期内最后一笔成交价格';
COMMENT ON COLUMN klines.volume IS '成交量，周期内累计成交的基础币种数量';

-- 24小时统计表
CREATE TABLE IF NOT EXISTS tickers (
    symbol VARCHAR(20) PRIMARY KEY,                           -- 交易对符号，主键
    last_price VARCHAR(50) NOT NULL,                          -- 最新成交价
    change_24h VARCHAR(50) DEFAULT '0',                       -- 24小时价格变化
    volume_24h VARCHAR(50) DEFAULT '0',                       -- 24小时成交量
    high_24h VARCHAR(50) DEFAULT '0',                         -- 24小时最高价
    low_24h VARCHAR(50) DEFAULT '0',                          -- 24小时最低价
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP            -- 最后更新时间
);

COMMENT ON TABLE tickers IS '24小时市场统计数据表';
COMMENT ON COLUMN tickers.symbol IS '交易对符号，如BTC/USDT，主键';
COMMENT ON COLUMN tickers.last_price IS '最新成交价格';
COMMENT ON COLUMN tickers.change_24h IS '24小时价格变化金额（当前价格-24小时前价格）';
COMMENT ON COLUMN tickers.volume_24h IS '24小时累计成交量，基础币种数量';
COMMENT ON COLUMN tickers.high_24h IS '24小时内最高成交价格';
COMMENT ON COLUMN tickers.low_24h IS '24小时内最低成交价格';
COMMENT ON COLUMN tickers.updated_at IS '数据最后更新时间';

-- 创建索引优化
-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- 余额表索引
CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id);
CREATE INDEX IF NOT EXISTS idx_balances_currency ON balances(currency);
CREATE INDEX IF NOT EXISTS idx_balances_user_currency ON balances(user_id, currency);

-- 交易对表索引
CREATE INDEX IF NOT EXISTS idx_trading_pairs_symbol ON trading_pairs(symbol);
CREATE INDEX IF NOT EXISTS idx_trading_pairs_status ON trading_pairs(status);
CREATE INDEX IF NOT EXISTS idx_trading_pairs_base_currency ON trading_pairs(base_currency);
CREATE INDEX IF NOT EXISTS idx_trading_pairs_quote_currency ON trading_pairs(quote_currency);

-- 订单表索引（撮合引擎性能关键）
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_symbol ON orders(symbol);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_symbol_status ON orders(symbol, status);
CREATE INDEX IF NOT EXISTS idx_orders_symbol_side_status ON orders(symbol, side, status);
CREATE INDEX IF NOT EXISTS idx_orders_symbol_side_price ON orders(symbol, side, price) WHERE status IN (1, 2); -- 只对活跃订单建索引
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_updated_at ON orders(updated_at);

-- 成交记录表索引
CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
CREATE INDEX IF NOT EXISTS idx_trades_created_at ON trades(created_at);
CREATE INDEX IF NOT EXISTS idx_trades_symbol_created_at ON trades(symbol, created_at);
CREATE INDEX IF NOT EXISTS idx_trades_buy_user_id ON trades(buy_user_id);
CREATE INDEX IF NOT EXISTS idx_trades_sell_user_id ON trades(sell_user_id);
CREATE INDEX IF NOT EXISTS idx_trades_buy_order_id ON trades(buy_order_id);
CREATE INDEX IF NOT EXISTS idx_trades_sell_order_id ON trades(sell_order_id);

-- K线数据表索引
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval ON klines(symbol, interval);
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval_open_time ON klines(symbol, interval, open_time);
CREATE INDEX IF NOT EXISTS idx_klines_open_time ON klines(open_time);

-- 24小时统计表索引
CREATE INDEX IF NOT EXISTS idx_tickers_updated_at ON tickers(updated_at);

-- 插入初始交易对数据
INSERT INTO trading_pairs (symbol, base_currency, quote_currency, min_amount, max_amount, price_scale, amount_scale, status) VALUES
('BTC/USDT', 'BTC', 'USDT', '0.00001', '1000', 2, 8, 1),
('ETH/USDT', 'ETH', 'USDT', '0.001', '10000', 2, 6, 1),
('BNB/USDT', 'BNB', 'USDT', '0.01', '100000', 2, 4, 1)
ON CONFLICT (symbol) DO NOTHING;