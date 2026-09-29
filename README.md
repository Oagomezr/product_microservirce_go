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

> If you already have a Postgres volume with an old schema, remove it before recreating:  
> `docker compose down -v` (deletes local data).

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

### Health

| Method | Route     | Description                                   |
|--------|-----------|-----------------------------------------------|
| `GET`  | `/health` | Liveness/readiness — pings Postgres and Mongo |

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Commands (Postgres)

| Method | Route | Description |
|--------|-------|-------------|
| `POST` | `/products` | Create product — returns `product_id` |
| `PATCH` | `/products/{id}` | Partial update |
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

Response (`201 Created`):

```json
{"product_id": "550e8400-e29b-41d4-a716-446655440001"}
```

> JSON keys must be **snake_case** (`short_description`), not PascalCase.

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
2. `curl http://localhost:8080/health` → `{"status":"ok"}`
3. `POST /products` → `201` + copy `product_id` from the response
4. In Postgres:

```sql
SELECT id, event_type, published FROM outbox ORDER BY created_at DESC LIMIT 5;
```

5. After ~10 s, `published` should become `true`
6. In Mongo: `db.products.find().pretty()` (database from your `.env`, e.g. `query_product_db`)
7. `GET /products/{product_id}` should return the product from Mongo
8. Repeat with `PATCH` / `DELETE`

---

## Testing

### Run tests

```bash
go test ./...
```

Verbose output:

```bash
go test -v ./internal/b_application/service/...
```

### CI (GitHub Actions)

On every push or pull request to `main` / `master`, the [CI workflow](.github/workflows/ci.yml) runs:

1. `go test ./...`
2. `go build ./...`

You will see a green or red check on GitHub commits and PRs once the repo is pushed to GitHub.

### Go vs Java (Maven) test layout

| Java (Maven) | Go |
|--------------|-----|
| `src/main/java/.../Foo.java` | `internal/.../foo.go` |
| `src/test/java/.../FooTest.java` | `internal/.../foo_test.go` (same folder) |
| `@Test`, JUnit, Mockito | `func TestXxx(t *testing.T)`, stubs, `httptest` |
| `mvn test` | `go test ./...` |

In Go there is **no separate test source tree**. Any file ending in `_test.go` is a test file and lives **next to** the code it tests. The `testing` package is built in; `_test.go` files are excluded from the compiled binary.

### Test files in this project

| File | Layer | What it tests |
|------|-------|---------------|
| `internal/a_domain/model/product_patch_test.go` | Domain | `ProductPatch.IsEmpty()` — PATCH with no fields must fail |
| `internal/b_application/service/create_product_test.go` | Application | Validation, reject preset ID, happy path + outbox `AggregateID` |
| `internal/b_application/service/update_product_test.go` | Application | `mergeProduct`, empty patch, product not found |
| `internal/b_application/service/publish_outbox_test.go` | Application | Outbox relay for create/update and delete events |
| `internal/c_adapter/in/a_web/command/create_product_handler_test.go` | HTTP adapter | POST returns `product_id`, 405, 400 |

**Why `product_patch_test.go`?**  
`IsEmpty()` is domain logic on `ProductPatch`, not application logic. Domain rules are tested in `a_domain`; `update_product_test.go` tests `mergeProduct` and `Execute` in the service layer separately.

Test doubles (stubs) implement the same interfaces as real adapters (`ProductSaver`, `OutboxSaver`, etc.) — similar to Mockito mocks, but written as small structs. `go-sqlmock` fakes Postgres transactions for the create happy path.

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
│   ├── a_domain/model/     # domain + product_patch_test.go
│   ├── b_application/      # services + *_test.go
│   └── c_adapter/          # http, kafka, postgres, mongo
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Roadmap

- [x] Return `product_id` in `POST /products` response
- [x] Health check (`GET /health`)
- [x] Unit tests (services, domain, handler)
- [ ] Integration tests (Postgres/Mongo testcontainers)
- [x] GitHub Actions (build + test)
- [ ] Cloud deployment

---

## Author

**Oscar Alberto Gómez Rodríguez** — Backend Developer (Java · Go · Microservices · Hexagonal Architecture)
