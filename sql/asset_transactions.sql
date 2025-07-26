-- 创建资产交易记录表
CREATE TABLE IF NOT EXISTS asset_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    transaction_id VARCHAR(64) NOT NULL UNIQUE,
    currency VARCHAR(10) NOT NULL,
    type SMALLINT NOT NULL, -- 1-充值，2-提现
    amount DECIMAL(36,18) NOT NULL,
    fee DECIMAL(36,18) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1, -- 1-待处理，2-成功，3-失败，4-已取消
    address VARCHAR(255) DEFAULT '',
    tx_hash VARCHAR(128) DEFAULT '',
    remark TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_asset_transactions_user_id ON asset_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_transaction_id ON asset_transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_user_type ON asset_transactions(user_id, type);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_currency ON asset_transactions(currency);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_status ON asset_transactions(status);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_created_at ON asset_transactions(created_at);

-- 添加外键约束（假设users表存在）
-- ALTER TABLE asset_transactions ADD CONSTRAINT fk_asset_transactions_user_id 
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 添加检查约束
ALTER TABLE asset_transactions ADD CONSTRAINT chk_asset_transactions_type 
    CHECK (type IN (1, 2));

ALTER TABLE asset_transactions ADD CONSTRAINT chk_asset_transactions_status 
    CHECK (status IN (1, 2, 3, 4));

ALTER TABLE asset_transactions ADD CONSTRAINT chk_asset_transactions_amount 
    CHECK (amount > 0);

ALTER TABLE asset_transactions ADD CONSTRAINT chk_asset_transactions_fee 
    CHECK (fee >= 0);

-- 添加注释
COMMENT ON TABLE asset_transactions IS '资产交易记录表';
COMMENT ON COLUMN asset_transactions.id IS '交易记录ID，主键';
COMMENT ON COLUMN asset_transactions.user_id IS '用户ID，关联users表';
COMMENT ON COLUMN asset_transactions.transaction_id IS '交易ID，唯一标识';
COMMENT ON COLUMN asset_transactions.currency IS '币种代码，如BTC、ETH、USDT等';
COMMENT ON COLUMN asset_transactions.type IS '交易类型：1-充值，2-提现';
COMMENT ON COLUMN asset_transactions.amount IS '交易金额';
COMMENT ON COLUMN asset_transactions.fee IS '手续费';
COMMENT ON COLUMN asset_transactions.status IS '交易状态：1-待处理，2-成功，3-失败，4-已取消';
COMMENT ON COLUMN asset_transactions.address IS '地址（提现时有值，充值时可为空）';
COMMENT ON COLUMN asset_transactions.tx_hash IS '区块链交易哈希';
COMMENT ON COLUMN asset_transactions.remark IS '备注信息';
COMMENT ON COLUMN asset_transactions.created_at IS '创建时间';
COMMENT ON COLUMN asset_transactions.updated_at IS '更新时间';