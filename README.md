# go-payment-api

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-316192?style=flat-square&logo=postgresql&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)
![Build](https://img.shields.io/badge/build-passing-brightgreen?style=flat-square)

> Production-ready REST API for payment processing built with Go — featuring checkout, refund, webhook verification, JWT authentication, rate limiting, and full Docker support.

---

## ✨ Features

- 💳 **Checkout & Refund** — Full payment lifecycle with idempotency keys
- 🔔 **Webhook Handling** — HMAC signature verification for secure event processing
- 🔐 **JWT Authentication** — Stateless auth with refresh token rotation
- 🚦 **Rate Limiting** — Per-IP and per-user request throttling
- 📊 **Metrics** — Prometheus endpoint for observability
- 🐳 **Docker Ready** — Multi-stage build, docker-compose for local dev
- ✅ **Test Coverage** — Unit + integration tests ≥ 80%
- 📖 **Swagger Docs** — Auto-generated API documentation

---

## 🏗️ Architecture

```
go-payment-api/
├── cmd/
│   └── api/            # Application entrypoint
├── internal/
│   ├── auth/           # JWT middleware & token management
│   ├── payment/        # Core payment logic (checkout, refund)
│   ├── webhook/        # Webhook handler & HMAC verification
│   ├── db/             # PostgreSQL connection & migrations
│   └── middleware/     # Rate limiting, logging, recovery
├── pkg/
│   └── omise/          # Omise payment gateway client
├── docs/               # Swagger / OpenAPI specs
├── migrations/         # SQL migration files
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

---

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15 (or use Docker)

### Run with Docker

```bash
git clone https://github.com/BankPongtep/go-payment-api.git
cd go-payment-api
cp .env.example .env
docker-compose up -d
```

API is now running at `http://localhost:8080`

### Run locally

```bash
go mod download
make migrate
make run
```

---

## 📡 API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/v1/auth/login` | Get JWT token | ❌ |
| `POST` | `/api/v1/auth/refresh` | Refresh token | ❌ |
| `POST` | `/api/v1/payments/checkout` | Create payment charge | ✅ |
| `GET` | `/api/v1/payments/:id` | Get payment status | ✅ |
| `POST` | `/api/v1/payments/:id/refund` | Refund a payment | ✅ |
| `POST` | `/api/v1/webhooks/omise` | Omise event webhook | 🔑 HMAC |
| `GET` | `/metrics` | Prometheus metrics | ❌ |
| `GET` | `/health` | Health check | ❌ |

---

## 🔧 Example Usage

### Create a checkout

```bash
curl -X POST http://localhost:8080/api/v1/payments/checkout \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 150000,
    "currency": "THB",
    "token": "tokn_test_xxxx",
    "description": "Insurance premium - Policy #INS-001",
    "idempotency_key": "order-abc-123"
  }'
```

### Response

```json
{
  "id": "chrg_test_xxxx",
  "status": "successful",
  "amount": 150000,
  "currency": "THB",
  "paid": true,
  "created_at": "2026-05-06T10:30:00Z"
}
```

---

## ⚙️ Environment Variables

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/payments_db

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Omise Payment Gateway
OMISE_PUBLIC_KEY=pkey_test_xxxx
OMISE_SECRET_KEY=skey_test_xxxx
OMISE_WEBHOOK_SECRET=whsec_xxxx
```

---

## 🧪 Running Tests

```bash
# Unit tests
make test

# With coverage report
make test-coverage

# Integration tests (requires Docker)
make test-integration
```

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22 |
| Router | Chi |
| Database | PostgreSQL 15 + sqlx |
| Cache | Redis |
| Auth | JWT (golang-jwt) |
| Payment | Omise API |
| Docs | Swagger / swaggo |
| Metrics | Prometheus |
| Container | Docker + Docker Compose |
| CI/CD | GitHub Actions |

---

## 📄 License

MIT © [Pongtep Pratuan](https://github.com/BankPongtep)
