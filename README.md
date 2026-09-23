# Product Microservice

Product catalog microservice built in **Go** with **hexagonal architecture**, **CQRS**, and the **Transactional Outbox** pattern.

- **Writes (commands):** PostgreSQL — source of truth.
- **Reads (queries):** MongoDB — projected read model.
- **Integration:** Kafka + outbox relay + internal consumer.

---

## Architecture

```text
                    ┌─────────────────────────────────────────┐
  HTTP Commands     │           Product Microservice          │
  POST/PATCH/DELETE │                                         │
        │           │  Handler → Use Case → Postgres + Outbox │
        ▼           │              (same transaction)         │
   PostgreSQL       └──────────────────┬──────────────────────┘
   (write model)                      │
                                      │ relay (every 10s)
                                      ▼
                                 Kafka topic
                              (product-events)
                                      │
                                      ▼
                               Kafka Consumer
                                      │
                                      ▼
                                  MongoDB
                               (read model)
        │
        ▼
  HTTP Queries
  GET /products/{id}
  GET /products/search
  GET /products/category
```

### Layers (hexagonal)

| Folder                   |  Role                                                    |
|--------------------------|----------------------------------------------------------|
| `internal/a_domain`      | Domain models (`Product`, `ProductPatch`, `OutboxEvent`) |
| `internal/b_application` | Use cases, in/out ports, application errors              |
| `internal/c_adapter/in`  | HTTP handlers, Kafka consumer                            |
| `internal/c_adapter/out` | Postgres, Mongo, Kafka publisher                         |
| `internal/d_config`      | Configuration from `.env`                                |
| `cmd/products`           | Composition root (`main`)                                |

### Outbox flow

1. The use case persists the product and an `outbox` event (`published = false`) in the **same transaction**.
2. It returns the HTTP response without publishing to Kafka directly.
3. `PublishOutbox` (relay) reads pending events, publishes to Kafka, and marks `published = true`.
4. The consumer projects the event into Mongo (upsert or delete tombstone).

---

## Stack

- Go 1.26+
- PostgreSQL 16
- MongoDB 7
- Apache Kafka (KRaft, Confluent image 7.6)
- [franz-go](https://github.com/twmb/franz-go) (Kafka producer/consumer)

---

## Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Docker](https://www.docker.com/) and Docker Compose
- (Optional) [Postman](https://www.postman.com/) or `curl`

---

## Configuration

1. Copy the environment file:

```bash
cp .env.example .env
```

2. Make sure the values match Docker Compose (host `localhost`, default ports).

| Variable               | Description                         |
|------------------------|-------------------------------------|
| `HTTP_ADDR`            | HTTP port for the service (`:8080`) |
| `POSTGRES_*`           | Postgres connection (write side)    |
| `MONGO_*`              | Mongo connection (read side)        |
| `KAFKA_BROKERS`        | Kafka broker (`localhost:9092`)     |
| `KAFKA_TOPIC`          | Event topic (`product-events`)      |
| `KAFKA_CONSUMER_GROUP` | Consumer group for the projector    |

---

## Running locally

### 1. Infrastructure

```bash
docker compose up -d
```

Wait until Postgres, Mongo, and Kafka are healthy. The `kafka-init` service creates the `product-events` topic.

### 2. Application

```bash
go run ./cmd/products
```

The service exposes HTTP at `http://localhost:8080` and starts in the background:

- Kafka consumer → Mongo
- Outbox relay → Kafka (10-second tick)

---

## HTTP API

Base URL: `http://localhost:8080`

### Commands (Postgres)

| Method   | Route            | Description    |
|----------|------------------|----------------|
| `POST`   | `/products`      | Create product |
| `PATCH`  | `/products/{id}` | Partial update |
| `DELETE` | `/products/{id}` | Delete product |

### Queries (Mongo)

| Method | Route                                   | Description                |
|--------|-----------------------------------------|----------------------------|
| `GET`  | `/products/{id}`                        | Get by ID                  |
| `GET`  | `/products/search?query=...`            | Search (min. 3 characters) |
| `GET`  | `/products/category?category_id={uuid}` | List by category           |

### Error codes

Business errors are mapped via `AppError`:

| Kind           | HTTP |
|----------------|------|
| Validation     | 400  |
| Not found      | 404  |
| Conflict       | 409  |
| Infrastructure | 500  |

---

## Examples (curl)

### Create product

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Latex gloves\",
    \"price\": 25000,
    \"stock\": 100,
    \"description\": \"Disposable gloves, size M\",
    \"short_description\": \"Gloves M\",
    \"hidden\": false,
    \"images_urls\": [\"https://example.com/img.jpg\"],
    \"category_id\": \"550e8400-e29b-41d4-a716-446655440000\"
  }"
```

### Partial PATCH (price only)

```bash
curl -X PATCH http://localhost:8080/products/{PRODUCT_ID} \
  -H "Content-Type: application/json" \
  -d "{\"price\": 19900}"
```

### Get product (read model)

> After create/update/delete, wait ~10–20 s (relay + consumer) before calling GET.

```bash
curl http://localhost:8080/products/{PRODUCT_ID}
```

### Delete

```bash
curl -X DELETE http://localhost:8080/products/{PRODUCT_ID}
```

---

## Manual E2E test

1. `docker compose up -d` + `go run ./cmd/products`
2. `POST /products` → 201 response
3. In Postgres:

```sql
SELECT id, event_type, published FROM outbox ORDER BY created_at DESC LIMIT 5;
```

4. After ~10 s, `published` should become `true`
5. `GET /products/{id}` should return the product from Mongo
6. Repeat with `PATCH` / `DELETE`

---

## Design decisions

| Decision                      | Rationale                                                |
|-------------------------------|----------------------------------------------------------|
| **CQRS**                      | Scale reads (Mongo) independently from writes (Postgres) |
| **Outbox**                    | Atomicity between DB and events; retries if Kafka fails  |
| **PATCH with `ProductPatch`** | Distinguish omitted fields from zero values in Go        |
| **Kafka key = product ID**    | Per-aggregate ordering on the same partition             |
| **Delete = tombstone**        | `Value: nil` in Kafka; consumer projects the delete      |

---

## Repository structure

```text
product_microservice/
├── cmd/products/           # main, wiring
├── docker/postgres/        # init.sql
├── internal/
│   ├── a_domain/model/
│   ├── b_application/    # services, ports, app_errors
│   └── c_adapter/        # http, kafka, postgres, mongo
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Roadmap

- [ ] Return `id` in `POST /products` response
- [ ] Health check (`GET /health`)
- [ ] Unit tests (services) and integration tests
- [ ] GitHub Actions (build + test)
- [ ] Cloud deployment

---

## Author

**Oscar Alberto Gómez Rodríguez** — Backend Developer (Java · Go · Microservices · Hexagonal Architecture)
