# go-payment-api — Task Board & Technical Specification

> เอกสารนี้ครอบคลุมทุกงานที่ต้องทำ, spec แต่ละ endpoint, และ architecture decisions
> อัพเดทล่าสุด: 2026-05-06

---

## 📋 สารบัญ

1. [Overview](#overview)
2. [Task Board](#task-board)
3. [API Specification](#api-specification)
4. [Database Schema](#database-schema)
5. [Architecture Decisions](#architecture-decisions)
6. [Testing Checklist](#testing-checklist)
7. [Deployment Checklist](#deployment-checklist)

---

## Overview

**go-payment-api** คือ REST API สำหรับจัดการ payment lifecycle ครบวงจร โดยใช้ Omise Payment Gateway ซึ่งเป็น payment gateway มาตรฐานในไทย

### Tech Stack

| Layer | Technology | เหตุผล |
|-------|-----------|--------|
| Language | Go 1.22 | High performance, statically typed, production-proven |
| Router | Chi v5 | Lightweight, idiomatic Go, middleware-friendly |
| Database | PostgreSQL 15 | ACID, reliable, widely used in fintech |
| Cache / Queue | Redis 7 | Token blacklist, idempotency keys |
| Auth | JWT (HS256) | Stateless, scalable |
| Payment | Omise API | Thai-market payment gateway |
| Metrics | Prometheus | Industry standard observability |
| Container | Docker + Compose | Reproducible environments |
| CI/CD | GitHub Actions | Auto test + deploy on push |

---

## Task Board

### 🔴 Phase 1 — Core Foundation (สัปดาห์ 1)

#### DB Layer

- [ ] **DB-01** ต่อ PostgreSQL ด้วย pgx driver ใน `internal/db/db.go`
  - Connection pooling: max 25, idle 10, lifetime 5 min
  - Health check ping on startup
  - ใส่ retry logic 3 ครั้งถ้า connect ไม่ได้

- [ ] **DB-02** เขียน migration runner ใน `internal/db/migrate.go`
  - อ่าน `migrations/*.sql` ตามลำดับ
  - เก็บ migration history ใน table `schema_migrations`
  - Support `make migrate` และ `make migrate-rollback`

- [ ] **DB-03** สร้าง `migrations/001_create_payments.sql`
  - Table: `payments` (id, amount, currency, status, paid, description, idempotency_key, omise_charge_id, created_at, updated_at)
  - Index: status, created_at, idempotency_key (UNIQUE)

- [ ] **DB-04** สร้าง `migrations/002_create_users.sql`
  - Table: `users` (id, email, password_hash, role, active, created_at)
  - Index: email (UNIQUE)

#### Auth Layer

- [ ] **AUTH-01** เขียน `internal/auth/repository.go`
  - `FindByEmail(email string) (*User, error)`
  - `CreateUser(u *User) error`
  - Password hashing ด้วย bcrypt cost 12

- [ ] **AUTH-02** เขียน `internal/auth/service.go`
  - `Login(email, password) (accessToken, refreshToken, error)`
  - `RefreshToken(refreshToken) (newAccessToken, error)`
  - เก็บ refresh token ใน Redis พร้อม TTL 7 วัน
  - Blacklist refresh token เมื่อ logout

- [ ] **AUTH-03** อัพเดท `internal/auth/handler.go`
  - แทนที่ mock ด้วย service จริง
  - เพิ่ม `POST /auth/logout` endpoint
  - Response validation และ error handling ครบ

- [ ] **AUTH-04** อัพเดท JWT Middleware ใน `internal/middleware/middleware.go`
  - Extract `user_id` และ `role` จาก claims ใส่ใน context
  - ตรวจ token ใน Redis blacklist
  - Helper: `GetUserID(ctx)`, `GetRole(ctx)`

---

### 🟡 Phase 2 — Payment Integration (สัปดาห์ 2)

#### Omise Client

- [ ] **OMISE-01** สร้าง `pkg/omise/client.go`
  - Wrap omise-go library
  - Interface: `ChargeAPI` เพื่อให้ mock ใน test ได้ง่าย
  - Retry: 3 ครั้ง, exponential backoff สำหรับ network error
  - Log ทุก request/response (mask sensitive data)

```go
// Interface ที่ต้อง implement
type ChargeAPI interface {
    Create(amount int64, currency, token, description string) (*Charge, error)
    Retrieve(chargeID string) (*Charge, error)
    Refund(chargeID string, amount int64) (*Refund, error)
}
```

#### Payment Layer

- [ ] **PAY-01** สร้าง `internal/payment/repository.go`
  - `Create(p *Payment) error`
  - `FindByID(id string) (*Payment, error)`
  - `FindByIdempotencyKey(key string) (*Payment, error)`
  - `UpdateStatus(id, status string) error`

- [ ] **PAY-02** สร้าง `internal/payment/service.go`
  - `Checkout(req CheckoutRequest) (*Payment, error)`:
    1. ตรวจ idempotency key ใน DB — ถ้ามีแล้ว return result เดิม
    2. Call Omise CreateCharge
    3. บันทึก payment ลง DB
    4. Return response
  - `Refund(chargeID string, amount int64) (*Refund, error)`:
    1. ตรวจว่า payment มีอยู่และ paid = true
    2. ตรวจ refund amount ≤ original amount
    3. Call Omise Refund
    4. Update payment status ใน DB

- [ ] **PAY-03** อัพเดท `internal/payment/handler.go`
  - แทนที่ mock ด้วย service จริง
  - Validate request body ทุก field
  - Handle Omise error codes (declined, insufficient_fund, etc.)
  - เพิ่ม `GET /payments` — list payments พร้อม pagination

#### Webhook

- [ ] **WH-01** อัพเดท `internal/webhook/handler.go`
  - `charge.complete` → update `payments.status = 'successful'`, `paid = true`
  - `charge.expire` → update `payments.status = 'expired'`
  - `charge.fail` → update `payments.status = 'failed'`
  - เก็บ webhook event log ลง DB ทุกครั้ง

- [ ] **WH-02** สร้าง `migrations/003_create_webhook_events.sql`
  - Table: `webhook_events` (id, event_key, payload, processed_at, created_at)

---

### 🟢 Phase 3 — Production Ready (สัปดาห์ 3)

#### Observability

- [ ] **OBS-01** Prometheus metrics ใน `internal/middleware/metrics.go`
  - `http_requests_total` (counter, labels: method, path, status)
  - `http_request_duration_seconds` (histogram)
  - `payment_charges_total` (counter, labels: status)
  - `payment_amount_total` (counter, currency)

- [ ] **OBS-02** Structured logging ทุก layer
  - Request ID propagation (ผ่าน context)
  - Log level: DEBUG ใน dev, INFO ใน prod
  - ไม่ log sensitive data (card token, secret key)

#### Security

- [ ] **SEC-01** Rate limiting ละเอียดขึ้น
  - `/auth/login`: 5 req/min ต่อ IP (brute force protection)
  - `/payments/*`: 100 req/min ต่อ user
  - `/webhooks/*`: bypass rate limit (ใช้ HMAC แทน)

- [ ] **SEC-02** Input validation ทุก endpoint
  - Amount: > 0, ≤ 10,000,000 satangs
  - Currency: enum ["THB"]
  - Token: ต้องขึ้นต้นด้วย "tokn_"
  - Email: valid format

- [ ] **SEC-03** CORS configuration
  - Allow: production domain เท่านั้น
  - Methods: GET, POST, OPTIONS
  - Headers: Authorization, Content-Type

#### CI/CD

- [ ] **CI-01** สร้าง `.github/workflows/ci.yml`
  - Trigger: push to main, pull request
  - Jobs: lint → test → build → docker push
  - Test ต้องผ่าน coverage ≥ 80%

- [ ] **CI-02** สร้าง `.github/workflows/deploy.yml`
  - Trigger: tag `v*.*.*`
  - Deploy to Railway / Fly.io
  - Smoke test หลัง deploy

---

## API Specification

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
```
Authorization: Bearer <access_token>
```

---

### POST /auth/login

**Request**
```json
{
  "email": "agent@clickbroker.co.th",
  "password": "secret123"
}
```

**Response 200**
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "expires_in": 900
}
```

**Error Codes**
| Code | Reason |
|------|--------|
| 400 | Missing email or password |
| 401 | Invalid credentials |
| 429 | Too many attempts |

---

### POST /auth/refresh

**Request**
```json
{
  "refresh_token": "eyJhbGci..."
}
```

**Response 200**
```json
{
  "access_token": "eyJhbGci..."
}
```

---

### POST /payments/checkout ✅ Auth required

**Request**
```json
{
  "amount": 150000,
  "currency": "THB",
  "token": "tokn_test_xxxx",
  "description": "Insurance premium - Policy #INS-001",
  "idempotency_key": "order-abc-123"
}
```

> `amount` หน่วยเป็น **สตางค์** (150000 = 1,500.00 บาท)

**Response 201**
```json
{
  "id": "chrg_test_xxxx",
  "status": "successful",
  "amount": 150000,
  "currency": "THB",
  "paid": true,
  "description": "Insurance premium - Policy #INS-001",
  "created_at": "2026-05-06T10:30:00Z"
}
```

**Error Codes**
| Code | Reason |
|------|--------|
| 400 | Invalid request body |
| 402 | Card declined |
| 409 | Duplicate idempotency_key |
| 422 | Insufficient funds |

---

### GET /payments/:id ✅ Auth required

**Response 200**
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

### POST /payments/:id/refund ✅ Auth required

**Request**
```json
{
  "amount": 150000,
  "reason": "customer request"
}
```

**Response 200**
```json
{
  "id": "rfnd_test_xxxx",
  "charge_id": "chrg_test_xxxx",
  "amount": 150000,
  "status": "pending",
  "created_at": "2026-05-06T11:00:00Z"
}
```

---

### POST /webhooks/omise 🔑 HMAC required

**Headers**
```
OmiseKey: <hmac-sha256-signature>
Content-Type: application/json
```

**Payload (charge.complete)**
```json
{
  "key": "charge.complete",
  "data": {
    "object": "charge",
    "id": "chrg_test_xxxx",
    "status": "successful",
    "paid": true
  }
}
```

**Response 200**
```json
{ "received": true }
```

---

## Database Schema

```sql
-- payments
CREATE TABLE payments (
    id               VARCHAR(64)   PRIMARY KEY,
    amount           BIGINT        NOT NULL,
    currency         VARCHAR(3)    NOT NULL DEFAULT 'THB',
    status           VARCHAR(20)   NOT NULL DEFAULT 'pending',
    paid             BOOLEAN       NOT NULL DEFAULT FALSE,
    description      TEXT,
    idempotency_key  VARCHAR(128)  UNIQUE,
    omise_charge_id  VARCHAR(64),
    user_id          UUID,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- users
CREATE TABLE users (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    email          VARCHAR(255)  UNIQUE NOT NULL,
    password_hash  VARCHAR(255)  NOT NULL,
    role           VARCHAR(20)   NOT NULL DEFAULT 'user',
    active         BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- webhook_events
CREATE TABLE webhook_events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_key    VARCHAR(64) NOT NULL,
    payload      JSONB       NOT NULL,
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Architecture Decisions

### Idempotency Keys
ทุก checkout request ต้องส่ง `idempotency_key` ป้องกัน double charge กรณี network retry
- เก็บใน DB พร้อม UNIQUE constraint
- ถ้า key ซ้ำ → return result เดิมทันที (HTTP 200 ไม่ใช่ 409)

### Webhook Signature Verification
ใช้ HMAC-SHA256 ตรวจทุก webhook request จาก Omise
- ถ้า `OMISE_WEBHOOK_SECRET` ไม่ set → skip verification (dev mode เท่านั้น)
- Production ต้อง set เสมอ

### Error Handling Pattern
```go
// ทุก handler ใช้ pattern นี้
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// ตัวอย่าง
{"code": "card_declined", "message": "Your card was declined"}
```

### Amount Unit
ใช้หน่วย **สตางค์** (integer) ตลอด เพื่อหลีกเลี่ยง floating-point ปัญหา
- 100 สตางค์ = 1 บาท
- แสดงผลให้ user: `amount / 100`

---

## Testing Checklist

### Unit Tests (ต้อง mock Omise + DB)

- [ ] `auth/service_test.go` — login success, wrong password, expired token
- [ ] `payment/service_test.go` — checkout success, duplicate idempotency_key, card declined
- [ ] `webhook/handler_test.go` — valid HMAC, invalid HMAC, unknown event
- [ ] `middleware/middleware_test.go` — valid JWT, expired JWT, missing token

### Integration Tests (ใช้ test DB จริง)

- [ ] Full checkout flow: login → checkout → get payment
- [ ] Idempotency: checkout สอง request เดียวกัน → result เดียวกัน
- [ ] Refund flow: checkout → refund → verify status
- [ ] Webhook flow: simulate Omise event → verify DB updated

### Load Test

- [ ] k6 script: 100 concurrent users, checkout endpoint
- [ ] Target: p99 < 500ms, error rate < 0.1%

---

## Deployment Checklist

### Pre-deploy
- [ ] Set ENV variables ทุกตัวใน production
- [ ] Run `make migrate` บน production DB
- [ ] ตรวจ `OMISE_WEBHOOK_SECRET` set แล้ว
- [ ] JWT_SECRET ยาว ≥ 32 chars, random

### Post-deploy
- [ ] Smoke test: `GET /health` → 200
- [ ] Smoke test: `GET /metrics` → Prometheus data
- [ ] ตรวจ logs ว่าไม่มี ERROR ใน 5 นาทีแรก
- [ ] ตั้ง Prometheus alert: error rate > 1%

---

*สร้างโดย Pongtep Pratuan · go-payment-api v1.0.0*
