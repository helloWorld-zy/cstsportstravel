# Implementation Plan: 境内跟团游 MVP

**Branch**: `main` | **Date**: 2026-06-19 | **Spec**: specs/001-domestic-tour-mvp/spec.md

**Input**: Feature specification from `specs/001-domestic-tour-mvp/spec.md`

## Summary

This plan delivers the domestic group tour (境内跟团游) complete transaction loop as the first MVP of the travel booking OTA platform. The backend is a Go monolith with DDD module boundaries (user/product/order/payment/admin/common), exposing RESTful APIs consumed by three frontend applications: a Nuxt.js 3 SSR web platform, a WeChat mini program via Uni-App, and a Vue 3 + Element Plus admin system. The MVP covers user registration/login, product browsing with departure calendars, a four-step booking flow with 30-minute payment countdown, Alipay + WeChat Pay integration, order management with tiered refund approval, and RBAC-based admin product/order management.

## Technical Context

**Language/Version**: Go 1.26+ (MVP dev with 1.24+ compatible code)

**Primary Dependencies**: Gin (web), GORM v2 + pgx (database), zap (logging), viper (config)

**Storage**: PostgreSQL 17.x (primary), Redis 7.2+ (cache/session)

**Testing**: go test with testify, integration tests for payment/order flows

**Target Platform**: Windows Server 2022/2025, WinSW service wrapper

**Project Type**: Web service (monolith) + 3 frontend apps

**Performance Goals**: 10,000 concurrent users, product list P99 <=200ms, order confirm P99 <=500ms

**Constraints**: TLS 1.3 mandatory, AES-256-GCM field encryption, JWT RS256, audit logs >=6 months

**Scale/Scope**: 300 orders/day MVP target, 7 user stories, 3 frontend platforms

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. API-First Design | PASS | OpenAPI 3.0 contracts defined in contracts/ before implementation; unified response envelope `{code, message, data, trace_id}` |
| II. Domain-Driven Boundaries | PASS | Monolith with clear module boundaries (user/product/order/payment/admin/common); each domain owns its data model |
| III. Security-by-Design | PASS | JWT RS256 (15min lifetime), AES-256-GCM for sensitive fields, TLS 1.3, audit logs >=6 months, RBAC on all admin operations |
| IV. Progressive Delivery | PASS | MVP scope clearly bounded (domestic group tour only); independently deployable; monolith-first with service extraction deferred |
| V. Code Quality | PASS | golangci-lint, ESLint+Prettier, Conventional Commits, >=70% test coverage on core logic, integration tests for payment/order paths |

## Project Structure

### Documentation (this feature)

```
specs/001-domestic-tour-mvp/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── user-api.yaml
│   ├── product-api.yaml
│   ├── order-api.yaml
│   ├── payment-api.yaml
│   └── admin-api.yaml
└── tasks.md             # Phase 2 output (by /speckit-tasks)
```

### Source Code (repository root)

```
# Backend (Go monolith)
cmd/
└── server/
    └── main.go                    # Application entry point

internal/
├── user/                          # User domain module
│   ├── handler/                   # HTTP handlers (Gin)
│   ├── service/                   # Business logic
│   ├── repository/                # Data access (GORM)
│   └── model/                     # Domain models
├── product/                       # Product domain module (same structure)
├── order/                         # Order domain module
├── payment/                       # Payment domain module
├── admin/                         # Admin domain module
└── common/                        # Shared infrastructure
    ├── middleware/                 # Auth, RBAC, audit, rate-limit
    ├── database/                  # PostgreSQL + Redis connections
    ├── cache/                     # Cache abstraction
    ├── logger/                    # Zap structured logging
    ├── config/                    # Viper configuration
    ├── encrypt/                   # AES-256-GCM field encryption
    └── response/                  # Unified API response envelope

migrations/                        # PostgreSQL migrations (golang-migrate)
├── 001_create_users.up.sql
├── 002_create_products.up.sql
├── 003_create_orders.up.sql
├── 004_create_payments.up.sql
└── 005_create_admin.up.sql

configs/
├── config.yaml                    # Default configuration
└── config.production.yaml         # Production overrides

# Frontend - Web Sales Platform (Nuxt.js 3 SSR)
web/
├── nuxt.config.ts
├── pages/
│   ├── index.vue                  # Homepage
│   ├── products/
│   │   ├── index.vue              # Product list
│   │   └── [id].vue               # Product detail
│   ├── booking/
│   │   └── [productId].vue        # Booking flow (4-step wizard)
│   ├── payment/
│   │   └── [orderId].vue          # Payment page
│   ├── user/
│   │   ├── index.vue              # Personal center
│   │   ├── orders.vue             # Order list
│   │   ├── order-[id].vue         # Order detail
│   │   └── travellers.vue         # Frequent travellers
│   └── auth/
│       └── login.vue              # Login/register
├── components/
│   ├── ProductCard.vue
│   ├── DepartureCalendar.vue
│   ├── PriceCalendar.vue
│   ├── TravellerForm.vue
│   ├── OrderStatusTag.vue
│   ├── PaymentCountdown.vue
│   └── BookingWizard.vue
├── composables/
│   ├── useApi.ts                  # API client
│   ├── useAuth.ts                 # Auth state
│   └── useCart.ts                 # Booking state
├── shared/                        # Shared with mini-program
│   ├── types/                     # TypeScript types (auto-generated from OpenAPI)
│   ├── utils/                     # Date, amount, ID validation
│   └── validators/                # Form validation rules
└── assets/
    └── tokens.css                 # Design tokens

# Frontend - WeChat Mini Program (Uni-App)
miniapp/
├── pages/
│   ├── index.vue                  # Homepage
│   ├── products/
│   │   ├── list.vue               # Product list
│   │   └── detail.vue             # Product detail
│   ├── booking/
│   │   └── index.vue              # Booking flow
│   ├── payment/
│   │   └── index.vue              # Payment (wx.requestPayment)
│   ├── orders/
│   │   ├── list.vue               # Order list
│   │   └── detail.vue             # Order detail
│   └── auth/
│       └── login.vue              # Login (wx.login)
├── components/                    # Shared with web where possible
├── shared/                        # Shared business logic
└── static/

# Frontend - Admin System (Vue 3 + Element Plus)
admin-web/
├── src/
│   ├── views/
│   │   ├── login.vue
│   │   ├── product/
│   │   │   ├── ProductList.vue
│   │   │   ├── ProductForm.vue
│   │   │   ├── ProductReview.vue
│   │   │   ├── PriceCalendar.vue
│   │   │   └── ItineraryEditor.vue
│   │   ├── order/
│   │   │   ├── OrderList.vue
│   │   │   ├── OrderDetail.vue
│   │   │   └── RefundReview.vue
│   │   ├── config/
│   │   │   └── CancellationRule.vue
│   │   └── system/
│   │       ├── UserManage.vue
│   │       ├── RoleManage.vue
│   │       └── PermissionManage.vue
│   ├── router/
│   │   └── index.ts               # Dynamic route from RBAC
│   ├── stores/                    # Pinia stores
│   ├── api/                       # API client layer
│   └── utils/
└── vite.config.ts
```

**Structure Decision**: Go monolith with DDD module boundaries (internal/), three independent frontend projects sharing types and utilities via a shared/ directory. The monolith approach per Constitution Principle IV avoids premature microservice complexity while preserving clean domain boundaries for future extraction.

## Configuration Management

**Decision**: Viper for static config loading + Consul KV for dynamic config updates.

**Rationale**:
- **Viper** loads initial configuration from YAML files (`configs/config.yaml`, `configs/config.production.yaml`) and environment variables at startup. This covers database credentials, Redis connection, JWT keys, payment channel configs, and other rarely-changed settings.
- **Consul KV** serves as the dynamic configuration center for settings that may change at runtime without service restart: rate limiting thresholds, feature flags, homepage banner config, maintenance mode toggle, payment timeout duration. The application watches Consul KV prefixes and updates in-memory config automatically.
- This approach aligns with PRD §10.4.3's requirement for "动态热更新" for specific config items, while avoiding the complexity of a full config center (Nacos/Apollo) for the MVP single-server deployment.
- Sensitive config values (DB passwords, API keys, encryption keys) are stored as environment variables per Constitution deployment constraints, NOT in config files or Consul KV.

**Config categories**:

| Category | Storage | Hot Reload | Examples |
|----------|---------|------------|----------|
| Static | YAML file + env vars | No (restart required) | DB connection, Redis addr, JWT keys, payment channel secrets |
| Dynamic | Consul KV | Yes (watch + auto-reload) | Rate limits, feature flags, maintenance mode, payment timeout |
| Sensitive | Environment variables only | No | DB passwords, API keys, encryption master key |

## API Design

### Unified Response Envelope

All API responses follow:

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "trace_id": "uuid"
}
```

Error codes: 0 = success, 1xxx = client errors (validation, auth), 2xxx = business errors (insufficient stock, payment failed), 5xxx = server errors.

### API Endpoint Summary

#### Auth & User Module (`/api/v1/auth/*`, `/api/v1/users/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/v1/auth/sms-code | Send SMS verification code | None |
| POST | /api/v1/auth/login | Phone + code login/register | None |
| POST | /api/v1/auth/wechat | WeChat OAuth login | None |
| POST | /api/v1/auth/admin/login | Admin username + password login | None |
| GET | /api/v1/users/me | Get current user profile | JWT |
| PUT | /api/v1/users/me | Update profile | JWT |
| POST | /api/v1/users/me/real-name | Submit real-name verification | JWT |
| GET | /api/v1/users/me/travellers | List frequent travellers | JWT |
| POST | /api/v1/users/me/travellers | Add frequent traveller | JWT |
| PUT | /api/v1/users/me/travellers/{id} | Update frequent traveller | JWT |
| DELETE | /api/v1/users/me/travellers/{id} | Delete frequent traveller | JWT |

#### Product Module (`/api/v1/products/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/v1/products | List products with filters | None |
| GET | /api/v1/products/{id} | Get product detail | None |
| GET | /api/v1/products/{id}/departures | Get departure calendar | None |
| GET | /api/v1/products/{id}/itinerary | Get itinerary details | None |
| GET | /api/v1/products/{id}/reviews | Get product reviews | None |
| GET | /api/v1/products/search/suggest | Search autocomplete | None |

#### Order Module (`/api/v1/orders/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/v1/orders | Create order (4-step flow) | JWT |
| GET | /api/v1/orders | List user orders | JWT |
| GET | /api/v1/orders/{id} | Get order detail | JWT |
| POST | /api/v1/orders/{id}/cancel | Cancel order | JWT |
| POST | /api/v1/orders/{id}/refund | Submit refund request | JWT |
| POST | /api/v1/orders/{id}/confirm | Confirm travel completion | JWT |

#### Payment Module (`/api/v1/payments/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/v1/payments/create | Create payment order | JWT |
| GET | /api/v1/payments/{id}/status | Query payment status | JWT |
| POST | /api/v1/payments/notify/alipay | Alipay callback | Signature |
| POST | /api/v1/payments/notify/wechat | WeChat Pay callback | Signature |
| POST | /api/v1/payments/{id}/query | Active query (fallback) | JWT |

#### Admin Module (`/api/v1/admin/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/v1/admin/products | List products (admin view) | JWT + RBAC |
| POST | /api/v1/admin/products | Create product | JWT + RBAC |
| PUT | /api/v1/admin/products/{id} | Update product | JWT + RBAC |
| POST | /api/v1/admin/products/{id}/submit-review | Submit for review | JWT + RBAC |
| PUT | /api/v1/admin/products/{id}/approve | Approve product | JWT + RBAC |
| PUT | /api/v1/admin/products/{id}/reject | Reject product | JWT + RBAC |
| PUT | /api/v1/admin/products/{id}/suspend | Suspend product | JWT + RBAC |
| GET | /api/v1/admin/products/{id}/departures | Manage departures | JWT + RBAC |
| PUT | /api/v1/admin/products/{id}/departures/batch-price | Batch price update | JWT + RBAC |
| GET | /api/v1/admin/orders | List orders (admin view) | JWT + RBAC |
| GET | /api/v1/admin/orders/{id} | Order detail (admin view) | JWT + RBAC |
| GET | /api/v1/admin/refunds | List refund requests | JWT + RBAC |
| PUT | /api/v1/admin/refunds/{id}/approve | Approve refund | JWT + RBAC |
| PUT | /api/v1/admin/refunds/{id}/reject | Reject refund | JWT + RBAC |
| GET | /api/v1/admin/users | List admin users | JWT + RBAC |
| POST | /api/v1/admin/users | Create admin user | JWT + RBAC |
| PUT | /api/v1/admin/users/{id}/status | Freeze/activate user | JWT + RBAC |
| GET | /api/v1/admin/roles | List roles | JWT + RBAC |
| POST | /api/v1/admin/roles | Create role | JWT + RBAC |
| PUT | /api/v1/admin/roles/{id} | Update role permissions | JWT + RBAC |
| GET | /api/v1/admin/menus | Get menu tree | JWT + RBAC |
| GET | /api/v1/admin/cancellation-rules | List cancellation rule templates | JWT + RBAC |
| POST | /api/v1/admin/cancellation-rules | Create cancellation rule template | JWT + RBAC |

## Database Design

Reference `data-model.md` for the complete schema. Key design decisions:

- **Sensitive field encryption**: `real_name`, `id_card_no` in user_account, frequent_traveller, and order_traveller tables are stored as AES-256-GCM ciphertext
- **Inventory management**: `departure_date` table tracks `total_stock`, `sold_count`, `locked_count`; available = total_stock - sold_count - locked_count
- **Order state machine**: 9 states as defined in PRD table 6-5, enforced at service layer with status_log audit trail
- **Refund tiered approval**: `refund_record` tracks approval level (operator/finance_director/director) based on amount thresholds
- **Index strategy**: Composite indexes on (user_id, order_status, created_at) for C-side order queries; (product_id, departure_date) for departure lookups; (status, created_at) for product listing

## Security Architecture

### Authentication Flow

1. **SMS Login**: Client sends phone + code -> server validates code from Redis -> creates/returns user + JWT (RS256, 15min access token, 7-day refresh token)
2. **WeChat Login**: Client calls wx.login -> sends code to backend -> backend exchanges for openid via WeChat API -> creates/binds user -> returns JWT
3. **Admin Login**: Username + password (Argon2id) -> returns JWT with role/permissions embedded in claims

### Field Encryption

- Algorithm: AES-256-GCM with per-field random IV
- Encrypted fields: real_name, id_card_no, passport_no, bank_card_no
- Key storage: Environment variables for MVP (single server); migration to KMS in Phase 2
- API responses: Mask sensitive fields (e.g., ID card shows last 4 digits only)

### RBAC Implementation

- Three-level permission model: menu permission (page visibility) + button permission (action enable/disable) + API permission (endpoint access)
- Data isolation: Supplier users have `supplier_id` filter automatically applied to product and order queries
- Audit logging: All admin operations produce audit entries with operator, action, target, timestamp, IP address

### Rate Limiting

- SMS code: 1 per 60s per phone, 10 per hour per IP
- Login: 5 failed attempts per 15min per phone (account lockout)
- API general: 100 req/s per user (configurable)

## Frontend Architecture

### Three-Frontend Strategy

| Platform | Framework | Rendering | Primary Users |
|----------|-----------|-----------|---------------|
| Web (C-side) | Nuxt.js 3 | SSR (product pages), SPA (user/order pages) | Consumers |
| Mini Program | Uni-App (Vue 3) | Native mini program | Mobile consumers |
| Admin | Vue 3 + Element Plus | SPA | Operations staff, suppliers |

### Shared Code

- `web/shared/types/`: TypeScript interfaces auto-generated from OpenAPI contracts using openapi-typescript
- `web/shared/utils/`: Date formatting, amount calculation (integer cents), ID card validation (ISO 7064:1983.MOD 11-2), phone validation
- `web/shared/validators/`: Form validation rules shared between web and mini program

### SSR Strategy

SSR applies to SEO-critical pages only: homepage, product list, product detail. Authenticated pages (user center, order management, booking flow) use SPA mode to reduce server load.

## Development Phases

### Phase 1: Foundation (Week 1-2)

- Go module initialization, project scaffolding (cmd/server, internal/ packages)
- PostgreSQL database setup with golang-migrate migrations
- Common infrastructure: config (viper), logging (zap), database connection pool (pgx/GORM), Redis client
- Unified API response envelope and error handling
- JWT authentication middleware (RS256 key generation/validation)
- RBAC middleware skeleton
- Frontend project scaffolding (Nuxt.js, Uni-App, Vue 3 admin)
- CI/CD pipeline (GitHub Actions, Windows runner)

### Phase 2: User Module (Week 2-3)

- SMS verification code service (Redis-backed, 5min TTL, 60s cooldown)
- Phone + code registration and login API
- WeChat OAuth 2.0 integration (openid exchange, account binding)
- Real-name verification API (ID card format validation, mock verification for MVP)
- Frequent traveller CRUD API (max 20 per user, encrypted storage)
- Admin login API (username + password, Argon2id)
- Frontend: Login/register pages (Web + Mini Program), Admin login page

### Phase 3: Product Module (Week 3-5)

- Product CRUD API with status state machine (draft -> pending_review -> approved / returned_for_revision -> suspended; approved -> change_pending_review -> approved for key field edits per spec FR-008)
- Itinerary editor (per-day JSONB storage, template save/reuse)
- Departure date management (per-day pricing, stock tracking)
- Price calendar API with batch update (5 modes: fixed/percentage/amount/formula/follow)
- Cancellation rule template CRUD
- Product list API with multi-filter + sort + pagination
- Product detail API with itinerary, fee breakdown, departure calendar, reviews
- Search autocomplete API (hot destinations, product names)
- Product review workflow (supplier submit -> operator approve/reject)
- Frontend: Product list/detail (Web + Mini Program), Product management + review + price calendar (Admin)

### Phase 4: Order & Payment Module (Week 5-7)

- Order creation with inventory pre-lock (Redis atomic decrement + DB row lock fallback)
- Order state machine implementation (9 states, transition rules from PRD table 6-5)
- 30-minute payment countdown (Asynq delayed task for auto-cancel)
- Alipay integration (PC web pay + mobile web pay)
- WeChat Pay integration (Native QR + JSAPI + Mini Program)
- Payment callback handling with idempotency (DB unique constraint + Redis dedup)
- Active payment status query fallback (when callback is lost)
- Refund request with auto-calculation (cancellation rule engine by days-before-departure)
- Tiered refund approval (<=1000 operator, 1000-5000 finance director, >5000 director)
- Payment channel refund execution (Alipay + WeChat refund APIs)
- Order list/detail/filter APIs (C-side + admin)
- Frontend: Booking wizard (4-step), Payment page with countdown, Order management (all platforms), Refund review (Admin)

### Phase 5: Integration & Hardening (Week 7-8)

- End-to-end integration tests (registration -> product browse -> booking -> payment -> order complete -> refund)
- Security hardening: field encryption audit, audit log completeness, rate limiting verification
- Performance optimization: product list caching (Redis), departure calendar caching, DB query optimization
- Deployment setup: WinSW service configuration, Traefik reverse proxy, TLS certificate
- Load testing (target: 10k concurrent users, product list P99 <=200ms)
- UAT with stakeholder feedback

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Payment callback reliability | Active query fallback + idempotent processing with DB unique constraint |
| Inventory overselling | Redis atomic decrement for hot path + PostgreSQL SELECT FOR UPDATE as consistency guarantee |
| Sensitive data exposure | AES-256-GCM field encryption + API response masking (last 4 digits) |
| SSR performance degradation | Multi-level cache strategy: CDN for static assets, Redis for product data, local in-memory for config |
| WeChat mini program review rejection | Strict compliance with WeChat content policies; no external links in mini program |
| Refund amount calculation errors | Unit tests covering all cancellation rule edge cases; manual review for amounts >5000 |
| Concurrent booking conflicts | Redis distributed lock on departure_id during order creation; 30min auto-release on payment timeout |
| SMS delivery failure | Retry with exponential backoff; fallback to voice verification code (Phase 2) |
