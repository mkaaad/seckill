# Seckill (Flash Sale) System

[中文](README_zh.md) | [English](README.md)

![Architecture Diagram](docs/代码流程图.png)

A distributed flash sale system built with Go microservices, Redis for inventory management, Kafka for message queueing, and MySQL for order persistence.

## Overview

This system handles high-concurrency flash sale scenarios with two independent microservices:

1. **Order Create Service** (`order-create-service/`): Accepts flash sale requests, validates inventory via Redis atomic operations, enforces rate limiting, and publishes successful orders to Kafka.
2. **Order Store Service** (`order-store-service/`): Consumes orders from Kafka and persists them to MySQL database.

## Architecture

```
Client → [order-create-service:8080] → Redis (inventory/rate limit) → Kafka → [order-store-service] → MySQL
```

### Key Features
- **Atomic inventory management** using Redis `DECR` operation
- **Rate limiting** (1 request per second per user) via Redis TTL keys
- **Asynchronous order processing** through Kafka message queue
- **Separate read/write concerns** with independent microservices
- **File-based logging** for error tracking

## Prerequisites

- **Go 1.25.0+** (for building and running services)
- **Redis** (running on `localhost:6379`, no password)
- **Kafka** (running on `localhost:9092`, topic `write-order-to-mysql`)
- **MySQL** (running on `127.0.0.1:3306`, database `SecKill`, user `root`, password `123456`)

## Quick Start

### 1. Clone the repository
```bash
git clone <repository-url>
cd seckill
```

### 2. Start required services
```bash
# Start Redis (assuming Redis is installed)
redis-server

# Start MySQL (ensure database exists)
mysql -u root -p123456 -e "CREATE DATABASE IF NOT EXISTS SecKill;"

# Start Kafka (requires Kafka installed at /opt/kafka_2.13-4.0.0/)
./kafka/start_kafka.sh

# Create Kafka topic
/opt/kafka_2.13-4.0.0/bin/kafka-topics.sh --create --topic write-order-to-mysql --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

### 3. Build and run the services
```bash
# Terminal 1: Order Create Service
cd order-create-service
go run ./cmd
# Server starts on http://localhost:8080

# Terminal 2: Order Store Service  
cd order-store-service
go run ./cmd
# Starts Kafka consumer and MySQL writer
```

## API Documentation

### Create Flash Sale Product
**POST** `/seckill`

Creates a flash sale product with limited stock available for a specific time window.

| 名称 (Name) | 类型 (Type) | 必选 (Required) | 说明 (Description) |
| :---------- | :---------- | :-------------- | :----------------- |
| product_id  | int         | 是 (Yes)        | 商品ID (Product ID) |
| start_time  | time        | 是 (Yes)        | 开始时间 (Start time) |
| end_time    | time        | 是 (Yes)        | 结束时间 (End time) |
| price       | float       | 是 (Yes)        | 商品价格 (Price) |
| stock       | int         | 是 (Yes)        | 库存 (Stock quantity) |

**Example Request:**
```json
{
  "product_id": 1001,
  "start_time": "2024-12-01T10:00:00Z",
  "end_time": "2024-12-01T12:00:00Z",
  "price": 99.99,
  "stock": 1000
}
```

**Responses:**
- `200 OK`: `{"info": "添加秒杀成功"}` (Flash sale created successfully)
- `400 Bad Request`: `{"info": "数据格式错误"}` (Invalid data format)
- `500 Internal Server Error`: `{"info": "添加秒杀失败"}` (Failed to create flash sale)

### Place Flash Sale Order
**GET** `/order`

Attempts to purchase a product during a flash sale. Includes rate limiting (1 request per second per user) and atomic inventory checks.

| 名称 (Name) | 位置 (Location) | 类型 (Type) | 必选 (Required) | 说明 (Description) |
| :---------- | :-------------- | :---------- | :-------------- | :----------------- |
| product_id  | query           | int         | 是 (Yes)        | 商品ID (Product ID) |
| user_id     | query           | int         | 是 (Yes)        | 用户ID (User ID) |

**Example Request:**
```
GET /order?product_id=1001&user_id=5001
```

**Responses:**
- `200 OK`: `{"info": "订单创建成功", "order": {"product_id": 1001, "user_id": 5001}}` (Order created successfully)
- `400 Bad Request`: Various error messages (invalid parameters, sale not active, insufficient stock)
- `429 Too Many Requests`: `{"info": "请求过于频繁，稍后再试"}` (Rate limit exceeded)
- `500 Internal Server Error`: `{"info": "服务器内部错误"}` (Internal server error)

### Order Status Query (Not Implemented)
**GET** `/order/search`

*Note: This endpoint is documented but not yet implemented in the current codebase.*

| 名称 (Name) | 位置 (Location) | 类型 (Type) | 必选 (Required) | 说明 (Description) |
| :---------- | :-------------- | :---------- | :-------------- | :----------------- |
| user_id     | query           | int         | 是 (Yes)        | 用户ID (User ID) |

## Project Structure

```
seckill/
├── order-create-service/     # Flash sale order creation service
│   ├── cmd/main.go          # HTTP server (Gin) on port 8080
│   ├── handlers/            # Request handlers
│   ├── dao/db.go            # Redis client configuration
│   ├── model/model.go       # Product and Order structs
│   ├── logs/log.go          # File logging
│   └── go.mod               # Go module dependencies
├── order-store-service/     # Order persistence service
│   ├── cmd/main.go          # Kafka consumer and MySQL writer
│   ├── handlers/            # Kafka and MySQL handlers
│   ├── dao/db.go            # MySQL/GORM configuration
│   ├── model/model.go       # Order struct with GORM tags
│   ├── logs/log.go          # File logging with data logging
│   └── go.mod               # Go module dependencies
├── kafka/                   # Kafka management scripts
│   ├── start_kafka.sh       # Starts Kafka server
│   └── stop_kafka.sh        # Stops Kafka server
├── docs/                    # Documentation
│   └── 代码流程图.png       # Architecture diagram (Chinese)
├── AGENTS.md                # Guide for AI agents working on this project
└── README.md                # This file
```

## Development

### Building
```bash
# Build both services
cd order-create-service && go build ./cmd
cd ../order-store-service && go build ./cmd
```

### Dependencies
Each service manages its own dependencies via `go.mod`:
- **order-create-service**: Gin (HTTP), go-redis (Redis), sarama (Kafka producer)
- **order-store-service**: GORM (MySQL), sarama (Kafka consumer)

### Logging
- Both services write logs to `app.log` in their respective directories
- Errors are logged with `logs.WriteLog(err)`
- The order-store-service includes `WriteData(order)` for logging order data

## Configuration

### Service Endpoints (Hardcoded)
- **Redis**: `localhost:6379` (no password, DB 0)
- **Kafka**: `localhost:9092` (topic: `write-order-to-mysql`)
- **MySQL**: `127.0.0.1:3306` (database: `SecKill`, user: `root`, password: `123456`)

*To change these values, modify the respective `dao/db.go` and `handlers/kafka_*.go` files.*

## Known Limitations

1. **Hardcoded configuration** - All service endpoints are hardcoded
2. **Single Kafka partition** - Consumer assumes partition 0
3. **No graceful shutdown** - Services don't handle SIGTERM/SIGINT
4. **No health checks** - Missing readiness/liveness endpoints
5. **Chinese error messages** - Error responses are in Chinese only
6. **Missing `/order/search` endpoint** - Documented but not implemented

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes following existing code patterns
4. Test with Redis, Kafka, and MySQL running
5. Submit a pull request

## License

[Add appropriate license information here]