# Agent Guide for Seckill System

This guide provides essential information for AI agents working on this Go microservices project.

## Overview

This is a flash sale (seckill) system consisting of two independent Go microservices:
1. **order-create-service**: Handles flash sale requests, validates inventory via Redis, and pushes orders to Kafka.
2. **order-store-service**: Consumes orders from Kafka and persists them to MySQL.

Both services are separate Go modules with their own `go.mod` files.

## Project Structure

```
seckill/
├── README.md                 # API documentation (Chinese)
├── AGENTS.md                 # This file
├── docs/
│   └── 代码流程图.png        # Architecture diagram
├── kafka/
│   ├── start_kafka.sh       # Starts Kafka server (requires Kafka installed at /opt/kafka_2.13-4.0.0/)
│   └── stop_kafka.sh        # Stops Kafka server
├── order-create-service/     # First microservice
│   ├── go.mod               # Module: order-create, Go 1.25.0
│   ├── cmd/
│   │   └── main.go          # Gin server with two endpoints
│   ├── model/
│   │   └── model.go         # Product and Order structs
│   ├── dao/
│   │   └── db.go            # Redis client initialization
│   ├── handlers/
│   │   ├── place_seckill_handler.go  # POST /seckill – creates flash sale product
│   │   ├── place_order_handler.go    # GET /order – places order
│   │   └── kafka_handler.go          # Sends order to Kafka topic "write-order-to-mysql"
│   └── logs/
│       └── log.go           # File logging to app.log
└── order-store-service/     # Second microservice
    ├── go.mod               # Module: order-store, Go 1.25.0
    ├── cmd/
    │   └── main.go          # Starts Kafka consumer and writes to MySQL
    ├── model/
    │   └── model.go         # Order struct with GORM tags
    ├── dao/
    │   └── db.go            # MySQL/GORM initialization
    ├── handlers/
    │   ├── kafka_handlers.go # Consumes from Kafka topic "write-order-to-mysql"
    │   └── mysql_handlers.go # Writes orders to MySQL
    └── logs/
        └── log.go           # File logging with extra WriteData function
```

## Dependencies

### External Services (hardcoded endpoints)
- **Redis**: `localhost:6379` (no password) – used for inventory counting and rate limiting.
- **Kafka**: `localhost:9092` – topic `write-order-to-mysql`.
- **MySQL**: `127.0.0.1:3306` database `SecKill`, user `root`, password `123456`.

### Go Libraries
- **order-create-service**: `github.com/gin-gonic/gin`, `github.com/go-redis/redis/v8`, `github.com/IBM/sarama`
- **order-store-service**: `github.com/IBM/sarama`, `gorm.io/driver/mysql`, `gorm.io/gorm`

## Building and Running

### Prerequisites
- Go 1.25.0 or later
- Redis running on `localhost:6379`
- Kafka running on `localhost:9092` with topic `write-order-to-mysql`
- MySQL running on `127.0.0.1:3306` with database `SecKill`

### Building Each Service
Navigate to the service directory and use standard Go commands:

```bash
cd order-create-service
go build ./cmd
```

```bash
cd order-store-service
go build ./cmd
```

### Running Services
Start each service from its own directory:

```bash
cd order-create-service
go run ./cmd
```

```bash
cd order-store-service
go run ./cmd
```

### Kafka Management
Use the provided scripts (require Kafka installed at `/opt/kafka_2.13-4.0.0/`):

```bash
./kafka/start_kafka.sh
./kafka/stop_kafka.sh
```

## Testing

No test files (`*_test.go`) are present in the repository. To add tests, follow standard Go testing patterns.

## Code Patterns

### Error Handling
- Errors are logged using `logs.WriteLog(err)` which writes to `app.log` in the current working directory.
- HTTP errors return JSON with an `"info"` field containing Chinese error messages.
- Fatal errors call `log.Fatalln` or `log.Fatalf`.

### Logging
- Each service logs to `app.log` in its own directory (not centralized).
- Logs are written with `log.SetOutput(file)` and restored to stdout after each write.
- The order-store-service has an additional `WriteData` function to log order structs.

### Database/Redis Connections
- Connections are initialized in `dao.ClientDB()` and stored in global variables (`Rdb`, `Db`).
- The order-create-service uses Redis for:
  - Storing product stock with expiration tied to flash sale end time.
  - Rate limiting per user (1 request per second using key `"history"+userId`).
- The order-store-service uses GORM with auto‑migration for the `Order` table.

### Kafka Integration
- **Producer** (order‑create‑service): Sends JSON‑encoded `Order` to topic `write-order-to-mysql`.
- **Consumer** (order‑store‑service): Consumes messages from the same topic and inserts into MySQL.

### API Endpoints (from README.md)
- `POST /seckill` – creates a flash sale product (expects JSON: `product_id`, `start_time`, `end_time`, `price`, `stock`).
- `GET /order` – places an order (query parameters: `product_id`, `user_id`).
- `GET /order/search` – query order status (query parameter: `user_id`). *Implementation not yet present.*

## Gotchas & Important Notes

1. **Hardcoded Configuration**: All service endpoints (Redis, Kafka, MySQL) are hardcoded with localhost and default credentials. Change them in `dao/db.go` and `handlers/kafka_*.go` for production.

2. **Chinese Error Messages**: HTTP error responses use Chinese text (e.g., `"数据格式错误"`, `"添加秒杀失败"`). Keep this in mind when modifying handlers.

3. **Log File Location**: Each service writes to `app.log` in its own working directory. Ensure write permissions.

4. **Rate Limiting**: The order‑create‑service limits each user to one request per second using a Redis key with a 1‑second TTL.

5. **Inventory Management**: Stock decrement is atomic via `Redis.Decr`. If stock goes negative, it is incremented back and the request is rejected.

6. **No Graceful Shutdown**: Services do not implement graceful shutdown for Kafka consumers or HTTP servers.

7. **No Health Checks**: There are no health endpoints or readiness probes.

8. **Single Kafka Partition**: The consumer assumes a single partition (partition 0). Scaling may require changes.

9. **MySQL Schema**: The `Order` table is auto‑migrated with `ProductId` and `UserId` columns (both non‑nullable). No primary key is defined.

## Suggested Improvements (for future agents)

- Extract configuration to environment variables or config files.
- Add unit and integration tests.
- Implement graceful shutdown.
- Add health checks.
- Use structured logging (e.g., zap, logrus).
- Define a primary key for the `Order` table.
- Consider using Redis Lua scripts for more complex atomic operations.
- Add API documentation (Swagger/OpenAPI).