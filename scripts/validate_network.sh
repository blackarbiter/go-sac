#!/bin/bash
set -eo pipefail

# 定义网络名称
NETWORK="security-scan-net"

# 创建自定义网络
docker network create "$NETWORK" --driver bridge 2>/dev/null || true

# 启动临时测试容器
docker run --rm --network "$NETWORK" \
  -v $(pwd)/scripts:/scripts \
  alpine sh -c "
    apk add --no-cache mysql-client curl >/dev/null 2>&1

    echo 'Testing MySQL connection...'
    until mysql -h mysql -uroot -prootpass -e 'SELECT 1;' 2>/dev/null; do
      sleep 2
    done

    echo 'Testing RabbitMQ API...'
    until curl -s -u admin:admin123 http://rabbitmq:15672/api/overview >/dev/null; do
      sleep 2
    done

    echo 'Testing MinIO health...'
    until curl -s -u minioadmin:minioadmin123 http://minio:9000/minio/health/live >/dev/null; do
      sleep 2
    done

    echo 'All containers are reachable within the network!'
"