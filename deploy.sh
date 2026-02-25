#!/bin/bash
# ModernFM One-Click Deploy Script

echo "🚀 Starting ModernFM Deployment..."

# 1. 检查 Docker 环境
if ! [ -x "$(command -v docker-compose)" ]; then
  echo "❌ Error: docker-compose is not installed." >&2
  exit 1
fi

# 2. 准备数据目录
mkdir -p ./data/postgres ./data/redis

# 3. 启动容器集群
docker-compose -f deploy/docker-compose.yml up -d --build

echo "✅ ModernFM is now running!"
echo "🔗 UI URL: http://localhost (Port 80)"
echo "📡 Backend API: http://localhost:38866"
echo "🛠️ Database: PostgreSQL 15"
echo "⚡ Cache: Redis Alpine"
