#!/bin/bash

# 加密货币交易所启动脚本
# Crypto Exchange Startup Script

set -e

echo "🚀 Starting Crypto Exchange..."

# 检查配置文件
CONFIG_FILE="etc/exchange-api.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

# 检查依赖
echo "📦 Checking dependencies..."
go mod tidy

# 构建应用
echo "🔨 Building application..."
go build -o exchange .

# 检查数据库连接
echo "🗄️  Checking database connection..."
if ! docker-compose ps postgres | grep -q "Up"; then
    echo "⚠️  PostgreSQL is not running. Starting Docker services..."
    docker-compose up -d postgres redis
    echo "⏳ Waiting for database to be ready..."
    sleep 10
fi

# 检查 Redis 连接
echo "🔴 Checking Redis connection..."
if ! docker-compose ps redis | grep -q "Up"; then
    echo "⚠️  Redis is not running. Starting Redis..."
    docker-compose up -d redis
    sleep 5
fi

# 创建日志目录
mkdir -p logs

# 启动应用
echo "✅ Starting Exchange API Server..."
echo "📍 Server will be available at: http://localhost:8888"
echo "📊 API Documentation: http://localhost:8888/swagger/"
echo "🛑 Press Ctrl+C to stop the server"
echo ""

./exchange -f "$CONFIG_FILE"