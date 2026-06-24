# Research: 境内跟团游 MVP

**Date**: 2026-06-19

## R1: Inventory Concurrency Control

**Decision**: Redis atomic decrement + PostgreSQL row-level lock (SELECT FOR UPDATE)

**Rationale**:
- Redis `DECRBY` provides O(1) atomic inventory check-and-decrement for the hot path
- PostgreSQL `SELECT FOR UPDATE` serves as the consistency guarantee when Redis is unavailable or during reconciliation
- Two-phase approach: Redis handles real-time concurrency; DB is the source of truth for stock counts
- Scheduled reconciliation task (Asynq) periodically syncs Redis counters with DB to prevent drift

**Alternatives considered**:
- Database-only with optimistic locking: Higher latency under contention (p99 >500ms under load)
- Redis-only without DB sync: Risk of data loss on Redis failure; inventory count becomes inconsistent
- Distributed lock (Redlock) per request: Higher latency due to lock acquisition round-trips; unnecessary given Redis atomic operations

## R2: Payment Idempotency

**Decision**: Database unique constraint on (order_id, channel, attempt_no) + Redis dedup key with 24h TTL

**Rationale**: Payment callbacks from Alipay/WeChat can arrive multiple times (at-least-once delivery). The DB constraint prevents duplicate payment records at the storage level. Redis provides a fast-path dedup check before hitting the database, reducing latency for the common case (first callback).

**Implementation details**:
- On first callback: check Redis key `payment:dedup:{order_id}:{channel}` -> if absent, process payment, set Redis key with 24h TTL, insert DB record
- On duplicate callback: Redis key exists -> return success immediately without reprocessing
- Fallback: If Redis is down, rely on DB unique constraint to reject duplicates

## R3: SSR vs SPA Page Strategy

**Decision**:
- SSR: Homepage, Product list, Product detail (SEO-critical, public-facing)
- SPA: User center, Order management, Booking flow, Payment page (authenticated, no SEO need)

**Rationale**: SSR adds server-side rendering overhead (CPU, memory per request). Applying SSR only to SEO-critical pages balances crawlability with server resource efficiency. Nuxt.js 3 supports per-route rendering strategy configuration via `routeRules`.

**Cache strategy for SSR pages**:
- CDN edge cache: 5-minute TTL for product list, 10-minute TTL for product detail
- Redis cache: Product data with 5-minute TTL, invalidated on product update
- Nuxt built-in payload caching for SSR responses

## R4: Mini Program Login Flow

**Decision**: wx.login -> get code -> backend exchanges for openid -> if phone bound, login directly -> if not, prompt bind phone number via SMS verification

**Rationale**: This is the industry-standard WeChat mini program login flow. Phone binding is required for cross-platform identity (Web + Mini Program share the same user account). The flow ensures every user has a phone number for SMS notifications and order communication.

**Implementation details**:
1. Mini program calls `wx.login()` to get a temporary `code`
2. Mini program sends `code` to backend `/api/v1/auth/wechat`
3. Backend calls WeChat `jscode2session` API to exchange `code` for `openid` and `session_key`
4. If `openid` exists in `user_account.wechat_openid`, return JWT
5. If `openid` is new, return `{need_bindphone: true}`; mini program then calls `wx.getPhoneNumber()` or shows SMS verification form
6. On successful phone binding, create/link account and return JWT

## R5: Field Encryption Strategy

**Decision**: AES-256-GCM with per-field random IV (12 bytes), keys stored in environment variables (MVP), KMS in Phase 2

**Rationale**: Meets 等保三级 (MLPS Level 3) requirement for sensitive personal data protection. AES-256-GCM provides both confidentiality and integrity (authenticated encryption). Per-field IV ensures that identical plaintext values produce different ciphertext, preventing pattern analysis.

**Storage format**: `base64(iv + ciphertext + tag)` where iv=12 bytes, tag=16 bytes (GCM authentication tag)

**Encrypted fields**:
- `user_account.real_name`, `user_account.id_card_no`
- `frequent_traveller.real_name`, `frequent_traveller.id_card_no`
- `order_traveller.name`, `order_traveller.id_card_no`

**Key management plan**:
- MVP: Key stored in `ENCRYPTION_KEY` environment variable; rotation requires re-encryption batch job
- Phase 2: Migrate to cloud KMS (e.g., Tencent Cloud KMS) with automatic key rotation

## R6: Go Module Structure

**Decision**: Single Go module with `internal/` packages per domain, `cmd/server` as entry point

**Rationale**: Monolith-first per Constitution Principle IV. The `internal/` directory enforces package-level encapsulation (Go compiler prevents external imports). Each domain package (user, product, order, payment, admin) follows a consistent handler/service/repository layering pattern. This structure enables future service extraction by simply moving a domain package to its own module.

**Package dependency rules**:
- `common/` packages can be imported by any domain package
- Domain packages (user, product, etc.) MUST NOT import each other directly
- Cross-domain communication goes through service interfaces defined in the consuming domain
- `cmd/server` is the composition root that wires all domain packages together

## R7: Order Auto-Cancel on Payment Timeout

**Decision**: Asynq delayed task enqueued at order creation, fires after 30 minutes

**Rationale**: Asynq (Redis-backed task queue) provides reliable delayed execution. When an order is created in `pending_pay` status, a delayed task is enqueued with a 30-minute delay. When the task fires, it checks order status: if still `pending_pay`, it transitions to `cancelled`, releases inventory, and closes any open payment records.

**Alternatives considered**:
- Cron job scanning for expired orders: Introduces up to 1-minute delay; requires polling
- Redis key expiry notification (KEYSPACE): Fragile; depends on Redis notification configuration; not guaranteed delivery
- In-memory timer: Lost on service restart; not suitable for multi-instance deployment

## R8: Refund Rule Calculation Engine

**Decision**: Server-side calculation using product-level cancellation rules matched against days-before-departure

**Rationale**: Cancellation rules are stored per product as a set of `(days_before_departure_min, days_before_departure_max, refund_percentage)` rows. When a refund request is submitted, the engine calculates `days_remaining = departure_date - today`, finds the matching rule, and computes `refund_amount = payable_amount * refund_percentage`. This approach allows each product to have its own cancellation policy.

**Edge cases handled**:
- No matching rule (days_remaining exceeds all rules): 100% refund
- Refund on departure day or after: 0% refund (policy-driven)
- Partial refund already processed: Remaining amount tracked in `refund_record`

## R9: Batch Price Update Strategy

**Decision**: Five modes implemented as a single API endpoint with `mode` parameter

**Rationale**: Suppliers need flexible pricing tools. The five modes cover all common scenarios:
1. **Fixed price**: Set exact values (adult/child/infant/single_supplement) for selected dates
2. **Percentage**: Adjust current prices by +/- N% for selected dates
3. **Fixed amount**: Adjust current prices by +/- N yuan for selected dates
4. **Formula**: Apply expression (e.g., `adult_price * 1.1 + 50`) - deferred to Phase 2 for MVP simplicity
5. **Follow**: Copy prices from a reference date range to target dates

For MVP, modes 1-3 and 5 are implemented; mode 4 (formula) is deferred.
