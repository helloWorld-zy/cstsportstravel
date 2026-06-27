# Tasks: 境内跟团游 MVP

**Input**: Design documents from `specs/001-domestic-tour-mvp/`

**Prerequisites**: plan.md (required), spec.md (required), data-model.md, contracts/, research.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Project scaffolding, Go module, directory structure, database, frontend projects

- [x] T001 Initialize Go module `go mod init github.com/travel-booking/server` and create directory structure per plan.md (cmd/server, internal/user, internal/product, internal/order, internal/payment, internal/admin, internal/common, migrations, configs)
- [x] T002 [P] Create configuration system using Viper in `internal/common/config/config.go` with YAML loading, environment variable binding, and config struct definitions (database, redis, jwt, payment, sms)
- [x] T003 [P] Create database connection pool in `internal/common/database/postgres.go` using GORM v2 + pgx driver with connection pool settings (max open 25, max idle 5, lifetime 30min)
- [x] T004 [P] Create Redis client in `internal/common/cache/redis.go` using go-redis/v9 with connection pool and health check
- [x] T004b [P] Create Consul KV client in `internal/common/config/consul.go` — connect to Consul agent, watch key prefixes for dynamic config (rate limits, feature flags, payment timeout), auto-reload in-memory config on change, fallback to static config if Consul unavailable
- [x] T005 [P] Create structured logger in `internal/common/logger/logger.go` using zap with lumberjack rotation (100MB/file, 7 days retention, JSON format)
- [x] T006 [P] Create unified API response helper in `internal/common/response/response.go` with envelope format `{code, message, data, trace_id}` and error code constants
- [x] T007 Write database migration `migrations/001_create_users.up.sql` for user_account, real_name_verification, frequent_traveller tables per data-model.md User Domain
- [x] T008 Write database migration `migrations/002_create_products.up.sql` for category, product, itinerary, departure_date, price_rule, refund_rule, product_review, destination tables per data-model.md Product Domain
- [x] T009 Write database migration `migrations/003_create_orders.up.sql` for main_order, sub_order, order_status_log, order_traveller tables per data-model.md Order Domain
- [x] T010 Write database migration `migrations/004_create_payments.up.sql` for payment_transaction, refund_record tables per data-model.md Payment Domain
- [x] T011 Write database migration `migrations/005_create_admin.up.sql` for admin_user, role, permission, menu, admin_user_role, role_permission, role_menu, audit_log tables per data-model.md Admin Domain
- [x] T012 Create main entry point `cmd/server/main.go` with graceful shutdown, config loading, database init, Redis init, HTTP server start on configurable port
- [x] T013 [P] Initialize Nuxt.js 3 web project in `web/` with TypeScript, Element Plus, Pinia, @tanstack/vue-query, ESLint+Prettier
- [x] T014 [P] Initialize Uni-App mini-program project in `miniapp/` with Vue 3, TypeScript, uView UI 2.x
- [x] T015 [P] Initialize Vue 3 admin project in `admin-web/` with TypeScript, Element Plus, Pinia, vue-router, Vite

**Checkpoint**: All three projects scaffolded, database created with migrations applied, backend starts and connects to DB/Redis

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story

- [x] T016 Implement JWT RS256 token generation and validation in `internal/common/middleware/jwt.go` with Access Token (15min) and Refresh Token (7d), RS256 key pair loading from config
- [x] T017 Implement authentication middleware in `internal/common/middleware/auth.go` that extracts JWT from Authorization header, validates token, injects user_id and tenant_id into Gin context
- [x] T018 Implement RBAC middleware in `internal/common/middleware/rbac.go` that checks user permissions against required permission code, supports function/data/field permission levels
- [x] T019 Implement audit logging middleware in `internal/common/middleware/audit.go` that records all write operations (POST/PUT/DELETE) to audit_log table with user_id, action, resource, before/after values, IP address
- [x] T020 Implement rate limiting middleware in `internal/common/middleware/ratelimit.go` using token bucket algorithm (golang.org/x/time/rate), configurable per-IP and per-user limits
- [x] T021 Implement AES-256-GCM field encryption utility in `internal/common/encrypt/aes.go` with per-field IV generation, key loading from config, Encrypt/Decrypt functions for sensitive fields (id_card_no, phone)
- [x] T021b [P] Implement TOTP (Time-based One-Time Password) service in `internal/common/auth/totp.go` — generate TOTP secret, generate QR code URL for authenticator app enrollment, verify TOTP code (6-digit, 30s window), store encrypted TOTP secret in admin_user table per FR-030
- [x] T021c [P] Implement MFA middleware in `internal/common/middleware/mfa.go` — intercept sensitive operations (refund approval >1000 yuan, permission changes, data export), check if user has MFA enrolled, require TOTP verification before proceeding, support fallback to SMS verification code per FR-030
- [x] T021d [P] Implement API response field masking in `internal/common/response/masking.go` — MaskPhone (show first 3 + last 4, e.g. 138****8000), MaskIDCard (show first 6 + last 4, e.g. 110***********1234), MaskName (show surname only), MaskBankCard (show first 6 + last 4). Apply automatically in response serialization for all API endpoints per PRD §10.1.2
- [x] T021e [P] Implement API request signing middleware in `internal/common/middleware/signing.go` — HMAC-SHA256 signature verification, timestamp validation (±5 minute window to prevent replay attacks), nonce uniqueness check (Redis-backed dedup with 5-minute TTL), applied to all state-changing endpoints (POST/PUT/DELETE) per PRD §10.1.2
- [x] T022 [P] Create GORM models for User Domain in `internal/user/model/user.go` (UserAccount, RealNameVerification, FrequentTraveller structs with GORM tags, JSON tags, validation tags)
- [x] T023 [P] Create GORM models for Product Domain in `internal/product/model/product.go` (Product, Category, Itinerary, DepartureDate, PriceRule, RefundRule, ProductReview, Destination structs)
- [x] T024 [P] Create GORM models for Order Domain in `internal/order/model/order.go` (MainOrder, SubOrder, OrderStatusLog, OrderTraveller structs with state machine constants)
- [x] T025 [P] Create GORM models for Payment Domain in `internal/payment/model/payment.go` (PaymentTransaction, RefundRecord structs with status constants)
- [x] T026 [P] Create GORM models for Admin Domain in `internal/admin/model/admin.go` (AdminUser, Role, Permission, Menu, AuditLog structs)
- [x] T027 [P] Create HTTP router setup in `internal/common/router/router.go` with route group registration, middleware chain, health check endpoints (/health, /ready)
- [x] T028 [P] Set up web project routing framework in `web/app/router.ts` with file-system routes for all pages (index, products, booking, payment, user, auth)
- [x] T029 [P] Set up web API client layer in `web/composables/useApi.ts` with Axios instance, request/response interceptors, JWT token injection, error handling, TypeScript type integration
- [x] T030 [P] Set up web shared utilities in `web/shared/utils/` — date formatting (dayjs), amount formatting (integer cents to display), ID card validation (ISO 7064 checksum), phone validation
- [x] T031 [P] Set up web design tokens in `web/assets/tokens.css` with CSS variables for theme colors, spacing, typography
- [x] T032 [P] Set up mini-program API client in `miniapp/shared/api/request.ts` with uni.request wrapper, token management, error handling
- [x] T033 [P] Set up mini-program shared utilities in `miniapp/shared/utils/` — same business logic as web (date/amount/ID validation), conditional compilation for WeChat-specific APIs
- [x] T034 [P] Set up admin routing framework in `admin-web/src/router/index.ts` with vue-router dynamic route generation from RBAC menu permissions
- [x] T035 [P] Set up admin API client in `admin-web/src/api/request.ts` with Axios, JWT token, permission-based request headers
- [x] T036 [P] Set up admin layout component in `admin-web/src/layouts/MainLayout.vue` with sidebar menu (dynamic from RBAC), header, breadcrumb, content area

**Checkpoint**: Auth middleware works, RBAC middleware works, audit logging works, all three frontend projects have routing and API layers ready

---

## Phase 3: US1 - User Registration & Login (Priority: P1)

**Goal**: Users can register via phone+SMS, login via WeChat, complete real-name verification, manage frequent travellers

**Independent Test**: New user completes register→login→real-name verification flow; existing user logs in directly

### Backend API

- [x] T037 [US1] Implement SMS verification code service in `internal/user/service/sms.go` — generate 6-digit code, store in Redis with 5min TTL, 60s rate limit per phone, integrate with Alibaba Cloud SMS SDK
- [x] T038 [US1] Implement user repository in `internal/user/repository/user_repo.go` — Create, FindByPhone, FindByWechatOpenID, Update operations with GORM
- [x] T039 [US1] Implement user service in `internal/user/service/user.go` — Register (phone+code), Login (phone+code), GetProfile, UpdateProfile, BindWechat logic
- [x] T040 [US1] Implement user handler in `internal/user/handler/user.go` — POST /api/v1/auth/sms-code, POST /api/v1/auth/login, POST /api/v1/auth/wechat-login, GET /api/v1/user/profile, PUT /api/v1/user/profile endpoints per user-api.yaml
- [x] T041 [US1] Implement WeChat OAuth integration in `internal/user/service/wechat.go` — wx.login code exchange for openid, OAuth 2.0 flow for web, user creation/binding logic
- [x] T042 [US1] Implement real-name verification service in `internal/user/service/realname.go` — submit verification (name+id_card), call public security API for validation, update user real_name_status, encrypt id_card_no with AES-256-GCM before storage
- [x] T043 [US1] Implement real-name verification handler in `internal/user/handler/realname.go` — POST /api/v1/user/real-name-verification endpoint per user-api.yaml
- [x] T044 [US1] Implement frequent traveller repository in `internal/user/repository/traveller_repo.go` — Create, FindByUserID, Update, Delete operations
- [x] T045 [US1] Implement frequent traveller service in `internal/user/service/traveller.go` — CRUD with ID card validation (ISO 7064 checksum), max 20 per user, encrypt sensitive fields
- [x] T046 [US1] Implement frequent traveller handler in `internal/user/handler/traveller.go` — CRUD endpoints per user-api.yaml (GET/POST/PUT/DELETE /api/v1/users/me/travellers)
- [x] T047 [US1] Implement admin login endpoint in `internal/admin/handler/auth.go` — POST /api/v1/auth/admin/login with username+password, Argon2id verification, JWT generation with role permissions per admin-api.yaml

### Frontend - Web (Nuxt.js 3)

- [x] T048 [P] [US1] Create login/register page at `web/pages/auth/login.vue` — phone+SMS code form with 60s countdown, WeChat OAuth button, form validation (phone format, code 6-digit), loading/empty/error states
- [x] T049 [P] [US1] Create auth composable in `web/composables/useAuth.ts` — login/register/logout functions, token storage (httpOnly cookie), user state management, auth guard
- [x] T050 [US1] Create personal center page at `web/pages/user/index.vue` — user card (avatar/nickname/level), menu groups (account/orders/services), real-name verification status display
- [x] T051 [US1] Create real-name verification form component in `web/components/RealNameForm.vue` — name+id_card input, real-time ID card format validation, submit with loading state
- [x] T052 [US1] Create frequent traveller management page at `web/pages/user/travellers.vue` — traveller list, add/edit/delete forms (name/id_card/phone/birth/gender), ID card validation, max 20 limit

### Frontend - Mini Program (Uni-App)

- [x] T053 [P] [US1] Create mini-program login page at `miniapp/pages/auth/login.vue` — wx.login quick login, phone binding flow (getPhoneNumber API), loading/error states, conditional compilation `#ifdef MP-WEIXIN`
- [x] T054 [P] [US1] Create mini-program auth composable in `miniapp/shared/composables/useAuth.ts` — login flow, token management (uni.setStorageSync), user state

### Frontend - Admin (Vue 3)

- [x] T055 [P] [US1] Create admin login page at `admin-web/src/views/login.vue` — username+password form, remember me checkbox, loading state, redirect to original page after login
- [x] T056 [US1] Create admin auth store in `admin-web/src/stores/auth.ts` — login/logout, token management, user info, permission list, dynamic menu generation

**Checkpoint**: User can register via phone, login via WeChat, complete real-name verification, manage frequent travellers on Web and Mini Program. Admin can login to backend.

---

## Phase 4: US2 - Product Search & Browse (Priority: P1)

**Goal**: Users can browse products with filters, view product details with itinerary/pricing/calendar

**Independent Test**: User searches products, applies filters, views detail page with all information

### Backend API

- [x] T057 [US2] Implement category and destination repository in `internal/product/repository/category_repo.go` — tree structure queries, FindAll, FindByParentID
- [x] T058 [US2] Implement product repository in `internal/product/repository/product_repo.go` — FindWithFilters (destination/city/days/price/status), FindByID with preloads (itinerary/departures/reviews), full-text search with PostgreSQL
- [x] T059 [US2] Implement product service in `internal/product/service/product.go` — ListProducts (filters/sort/pagination), GetProductDetail, GetDepartureCalendar, Search autocomplete
- [x] T060 [US2] Implement product handler in `internal/product/handler/product.go` — GET /api/v1/products (list with filters), GET /api/v1/products/:id (detail), GET /api/v1/products/:id/departures (calendar), GET /api/v1/products/:id/itinerary, GET /api/v1/products/:id/reviews, GET /api/v1/search/autocomplete per product-api.yaml
- [x] T061 [US2] Implement product review repository and service in `internal/product/repository/review_repo.go` and `internal/product/service/review.go` — list reviews by product with pagination, rating statistics
- [x] T062 [US2] Implement homepage data API in `internal/product/handler/homepage.go` — GET /api/v1/homepage returning banner list, popular destinations, recommended products (rule-based: hot products for new users, same-destination for returning users)

### Frontend - Web (Nuxt.js 3)

- [x] T063 [US2] Create homepage at `web/pages/index.vue` — search box with autocomplete, 金刚区 icon grid (8 entries), Banner carousel (3-5s auto-rotate), popular destinations tabs, "猜你喜欢" recommendation section, loading skeleton
- [x] T064 [US2] Create product list page at `web/pages/products/index.vue` — filter bar (top tags: destination/city/days/price + side drawer: accommodation/theme/grade/transport), sort dropdown (6 options), product card grid/list toggle, pagination, empty state, loading skeleton
- [x] T065 [US2] Create product card component in `web/components/ProductCard.vue` — product image, name, price (起), days, departure cities, satisfaction score, tags (热销/新品/特价), sold-out overlay
- [x] T066 [US2] Create product detail page at `web/pages/products/[id].vue` — image carousel (lazy load), product info, itinerary timeline (Day 1/2/3 with spots/meals/hotel/transport), fee included/excluded, crowd pricing (adult/child/infant), single supplement note, cancellation policy (always visible, not collapsible), review section (5-star + dimension scores + filter), departure calendar (heatmap: green=adequate/orange=tight/grey=sold-out), fixed bottom booking bar
- [x] T067 [US2] Create departure calendar component in `web/components/DepartureCalendar.vue` — 3-month calendar grid, daily adult price, stock status indicator (adequate/tight/sold-out), date selection handler, past dates disabled
- [x] T068 [US2] Create product detail composable in `web/composables/useProduct.ts` — fetch product detail, departure calendar, reviews with caching (@tanstack/vue-query)

### Frontend - Mini Program (Uni-App)

- [x] T069 [P] [US2] Create mini-program homepage at `miniapp/pages/index.vue` — search bar, 金刚区 icons, banner swiper, product recommendations, loading state
- [x] T070 [P] [US2] Create mini-program product list at `miniapp/pages/products/list.vue` — filter tabs, product card list, pull-down refresh, infinite scroll, empty state
- [x] T071 [US2] Create mini-program product detail at `miniapp/pages/products/detail.vue` — swiper images, itinerary, pricing, departure calendar, booking button, share capability

### Frontend - Admin (Vue 3)

- [x] T072 [P] [US2] Create admin product list page at `admin-web/src/views/product/ProductList.vue` — product table (name/destination/status/supplier), filters (status/destination/supplier), actions (edit/review/suspend), pagination
- [x] T073 [US2] Create admin homepage config page at `admin-web/src/views/config/HomepageConfig.vue` — banner management (image/upload/link/position/sort/status), popular destination config

**Checkpoint**: Homepage displays content, product list with filters works, product detail shows all information, departure calendar shows prices and stock. Works on both Web and Mini Program.

---

## Phase 5: US3 - Booking & Payment (Priority: P1)

**Goal**: Users complete booking flow (select departure → fill travellers → addons → confirm → pay), 30min timeout auto-cancel

**Independent Test**: User completes full booking from product detail to payment success

### Backend API

- [x] T074 [US3] Implement inventory service in `internal/product/service/inventory.go` — LockStock (Redis DECR + DB SELECT FOR UPDATE), ReleaseStock (Redis INCR), GetAvailableStock, stock warning levels (adequate/tight/full) per research.md
- [x] T075 [US3] Implement order repository in `internal/order/repository/order_repo.go` — Create, FindByID, FindByUserID (with filters), UpdateStatus, CreateStatusLog
- [x] T076 [US3] Implement order service in `internal/order/service/order.go` — CreateOrder (validate real-name, lock inventory, calculate price with single room supplement/child pricing, create order+travellers+status log), CancelOrder, GetOrderList, GetOrderDetail
- [x] T077 [US3] Implement single room supplement calculation in `internal/order/service/pricing.go` — auto-add when adult count is odd, per-date pricing from departure_date.single_supplement field per PRD §4.2.5
- [x] T078 [US3] Implement child pricing rules in `internal/order/service/pricing.go` — child (2-12yr, no bed) uses child_price, infant (<2yr) uses infant_price, child must link to adult, max 1 infant per adult per PRD §4.2.5
- [x] T079 [US3] Implement order handler in `internal/order/handler/order.go` — POST /api/v1/orders (create), GET /api/v1/orders (list), GET /api/v1/orders/:id (detail), POST /api/v1/orders/:id/cancel per order-api.yaml
- [x] T080 [US3] Implement payment repository in `internal/payment/repository/payment_repo.go` — Create, FindByID, FindByOrderID, UpdateStatus
- [x] T081 [US3] Implement Alipay payment integration in `internal/payment/service/alipay.go` — using smartwalle/alipay/v3 SDK, CreatePayment (page pay + wap pay), VerifyNotification, QueryOrder per payment-api.yaml and PRD §5.1.1
- [x] T082 [US3] Implement WeChat payment integration in `internal/payment/service/wechat.go` — using wechatpay-go SDK, CreatePayment (Native + JSAPI), VerifyNotification, QueryOrder per payment-api.yaml and PRD §5.1.2
- [x] T083 [US3] Implement payment service in `internal/payment/service/payment.go` — CreatePayment (route to channel), HandleCallback (idempotent with DB unique constraint + Redis dedup per research.md), QueryPaymentStatus
- [x] T084 [US3] Implement payment handler in `internal/payment/handler/payment.go` — POST /api/v1/orders/:id/payment (create), POST /api/v1/payment/alipay/notify (callback), POST /api/v1/payment/wechat/notify (callback), GET /api/v1/payment/:id/status (query) per payment-api.yaml
- [x] T085 [US3] Implement order auto-cancel task in `internal/order/service/timeout.go` — using Asynq delayed task (30min), cancel order + release inventory + update status per research.md
- [x] T086 [US3] Implement payment success flow in `internal/order/service/payment_callback.go` — update order status to paid_full, send confirmation notification (SMS + in-app), record payment transaction

### Frontend - Web (Nuxt.js 3)

- [x] T087 [US3] Create booking wizard page at `web/pages/booking/[productId].vue` — 4-step wizard (departure→travellers→addons→confirm), step progress bar, real-time price summary footer, back/next navigation
- [x] T088 [US3] Create departure selection step component in `web/components/booking/DepartureStep.vue` — departure calendar, adult/child/infant counters with min/max limits, live price calculation (single supplement auto-add), available seats display, group size hint
- [x] T089 [US3] Create traveller form step component in `web/components/booking/TravellerStep.vue` — per-traveller form (name/id_card/phone/birth/gender), "select from frequent travellers" button, real-time ID card validation, child-adult linking, form validation
- [x] T090 [US3] Create addon selection step component in `web/components/booking/AddonStep.vue` — insurance/transfer checkboxes with prices, price updates on toggle
- [x] T091 [US3] Create order confirmation step component in `web/components/booking/ConfirmStep.vue` — product summary, traveller list, fee breakdown (product+supplement+addons=total), cancellation policy summary, "agree to policy" checkbox (required), submit button
- [x] T092 [US3] Create payment page at `web/pages/payment/[orderId].vue` — Alipay/WeChat payment selection, 30-minute countdown timer (mm:ss), payment status polling, success redirect, timeout alert
- [x] T093 [US3] Create payment countdown component in `web/components/PaymentCountdown.vue` — countdown timer with mm:ss display, color change at 5min warning, timeout callback

### Frontend - Mini Program (Uni-App)

- [x] T094 [US3] Create mini-program booking flow at `miniapp/pages/booking/index.vue` — 4-step wizard adapted for mini-program, same business logic as web
- [x] T095 [US3] Create mini-program payment at `miniapp/pages/payment/index.vue` — wx.requestPayment integration for WeChat pay, countdown timer, conditional compilation

### Frontend - Admin (Vue 3)

*US3 (Booking & Payment) is C端-only, no admin frontend tasks.*

**Checkpoint**: User can complete booking from product detail through payment. 30-minute timeout works. Single room supplement and child pricing calculate correctly. Payment callbacks update order status.

---

## Phase 6: US4 - Order Management (Priority: P2)

**Goal**: Users can view orders, check status, request refunds with tiered cancellation rules

**Independent Test**: User views order list, checks detail, submits refund request, refund calculated correctly

### Backend API

- [x] T097 [US4] Implement cancellation rule engine in `internal/order/service/cancellation.go` — load refund_rule by product, match days_before_departure to tier, calculate refund amount per PRD §6.2.4 table 6-6 formula: refund = paid_amount - occurred_fees - cancellation_fee - non_refundable
- [x] T098 [US4] Implement refund service in `internal/order/service/refund.go` — CreateRefundRequest (calculate amount, create refund_record), process refund via payment channel (original route back per PRD §5.3), update order status
- [x] T099 [US4] Implement refund handler in `internal/order/handler/refund.go` — POST /api/v1/orders/:id/refund (request), GET /api/v1/orders/:id/refund-status per order-api.yaml
- [x] T100 [US4] Implement order status auto-transition tasks in `internal/order/service/status_transition.go` — PENDING_TRAVEL on departure date, IN_TRAVEL on trip start, COMPLETED on return date+1, using Asynq scheduled tasks

### Frontend - Web (Nuxt.js 3)

- [x] T101 [US4] Create order list page at `web/pages/user/orders.vue` — status tabs (全部/待付款/待出行/退款中/已完成/已取消), order cards (image/name/date/amount/status/action buttons), search by product name/order no, pagination, empty state
- [x] T102 [US4] Create order detail page at `web/pages/user/order-[id].vue` — product info (with itinerary summary), traveller list, fee breakdown, payment records, cancellation policy, action buttons by status (待付款: pay+cancel, 待出行: refund, 已完成: review)
- [x] T103 [US4] Create refund request component in `web/components/RefundRequest.vue` — refund reason selector, refund amount preview (auto-calculated by cancellation rules), submit with confirmation dialog

### Frontend - Mini Program (Uni-App)

- [x] T104 [P] [US4] Create mini-program order list at `miniapp/pages/orders/list.vue` — status tabs, order cards, pull-down refresh, empty state
- [x] T105 [US4] Create mini-program order detail at `miniapp/pages/orders/detail.vue` — same info as web, adapted for mobile layout
- [x] T105b [US4] Implement review submission API in `internal/product/handler/review.go` and `internal/product/service/review.go` — POST /api/v1/products/:id/reviews (submit review with rating 1-5 stars + dimension scores for guide/itinerary/hotel/transport/food + text content + optional images), validate user has completed order for this product, auto-publish without moderation per spec assumptions
- [x] T105c [US4] Create review submission component in `web/components/ReviewForm.vue` — 5-star rating selector, dimension score inputs (导游/行程/住宿/交通/餐饮), text content (min 10 chars), image upload (max 5 photos), submit with loading state. Display on order detail page for completed orders

**Checkpoint**: User can view orders by status, see order details, submit refund request, submit reviews for completed orders. Refund amount calculated correctly by cancellation rules. Works on both Web and Mini Program.

---

## Phase 7: US5 - Admin Product Management (Priority: P2)

**Goal**: Suppliers can create products, operators can review/approve, manage price calendar and inventory

**Independent Test**: Supplier creates product → submits review → operator approves → product visible on C端

### Backend API

- [x] T106 [US5] Implement admin product repository in `internal/admin/repository/product_repo.go` — Create, Update, FindByID (with all relations), FindWithFilters, UpdateStatus
- [x] T107 [US5] Implement admin product service in `internal/admin/service/product.go` — CreateProduct (validates required fields, generates product_no DOM-{code}-{date}-{seq}), UpdateProduct (triggers re-review for key fields per PRD §6.1.1), SubmitForReview, ApproveProduct, RejectProduct
- [x] T108 [US5] Implement itinerary service in `internal/admin/service/itinerary.go` — SaveItinerary (per-day cards with spots/meals/hotel/transport), load from template
- [x] T109 [US5] Implement price calendar service in `internal/admin/service/price_calendar.go` — SetDailyPrice (adult/child/infant/single_supplement), BatchUpdatePrices (5 modes: fixed/percent/amount/formula/follow per PRD §6.1.9), holiday template management
- [x] T110 [US5] Implement departure/inventory service in `internal/admin/service/departure.go` — CreateDeparture, UpdateStock, GetStockStatus, manual stock adjustment with reason
- [x] T111 [US5] Implement admin product handler in `internal/admin/handler/product.go` — CRUD endpoints + review workflow + departure management + batch pricing per admin-api.yaml
- [x] T112 [US5] Implement product review workflow in `internal/admin/service/review.go` — submit review, approve (status→approved), reject (require reason), change review for key field edits

### Frontend - Admin (Vue 3)

- [x] T113 [US5] Create product form page at `admin-web/src/views/product/ProductForm.vue` — multi-step form (基础信息→行程编辑→价格配置→退改规则→库存设置), step validation, save draft
- [x] T114 [US5] Create itinerary editor component in `admin-web/src/components/ItineraryEditor.vue` — day cards (auto-generated from days count), per-day form (title/description/spots selector/meals/hotel/transport/images), drag-to-reorder, template save/load
- [x] T115 [US5] Create price calendar component in `admin-web/src/components/PriceCalendar.vue` — month grid view, per-cell display (adult price/stock status/special marker), inline edit, batch update dialog (5 modes), holiday template application
- [x] T116 [US5] Create product review page at `admin-web/src/views/product/ProductReview.vue` — review queue list, product detail preview, approve/reject actions with reason input

**Checkpoint**: Supplier can create product with itinerary and pricing, submit for review. Operator can review and approve. Price calendar and batch pricing work. Approved product visible on C端.

---

## Phase 8: US6 - Admin Order Management (Priority: P2)

**Goal**: Operators can query orders, review refund requests with tiered approval, configure cancellation rules

**Independent Test**: Operator searches orders, reviews refund with correct tiered approval, configures cancellation rule template

### Backend API

- [x] T117 [US6] Implement admin order service in `internal/admin/service/order.go` — ListOrders (multi-dimension filters), GetOrderDetail (with all relations), ExportOrders (async for >1000 rows)
- [x] T118 [US6] Implement admin refund review service in `internal/admin/service/refund_review.go` — ListRefundRequests, ReviewRefund (approve/reject), tiered approval logic (≤1000: operator, 1000-5000: finance supervisor, >5000: director per spec clarification)
- [x] T119 [US6] Implement cancellation rule template service in `internal/admin/service/cancellation_rule.go` — CRUD for refund_rule templates, assign template to product
- [x] T120 [US6] Implement admin order handler in `internal/admin/handler/order.go` — order list/detail/export, refund list/approve/reject, cancellation rule CRUD per admin-api.yaml

### Frontend - Admin (Vue 3)

- [x] T121 [US6] Create admin order list page at `admin-web/src/views/order/OrderList.vue` — order table with multi-dimension filters (order no/phone/status/date/product/supplier), sort, pagination, export button
- [x] T122 [US6] Create admin order detail page at `admin-web/src/views/order/OrderDetail.vue` — full order info (product/travellers/fees/payments/status log), refund section
- [x] T123 [US6] Create refund review page at `admin-web/src/views/order/RefundReview.vue` — refund request list, detail view (refund amount calculation/cancellation rule match/occurred fees), approve/reject with reason, tiered approval indicator
- [x] T124 [US6] Create cancellation rule editor at `admin-web/src/views/config/CancellationRule.vue` — tiered rate editor (days range → refund percentage rows), template save/load, assign to products

**Checkpoint**: Operator can search orders, review refunds with correct approval tiers, configure cancellation rule templates.

---

## Phase 9: US7 - RBAC & Admin User Management (Priority: P3)

**Goal**: Admin can manage users, roles, permissions with function+data+field level control

**Independent Test**: Admin creates user, assigns role, user sees only authorized menus and data

### Backend API

- [ ] T125 [US7] Implement admin user repository in `internal/admin/repository/admin_user_repo.go` — CRUD, FindByUsername, UpdatePassword (Argon2id), UpdateStatus
- [ ] T126 [US7] Implement role repository in `internal/admin/repository/role_repo.go` — CRUD, AssignPermissions, AssignMenus
- [ ] T127 [US7] Implement permission repository in `internal/admin/repository/permission_repo.go` — FindAll (tree structure), FindByRoleID
- [ ] T128 [US7] Implement RBAC service in `internal/admin/service/rbac.go` — CreateUser (with initial password, force change on first login), CreateRole, AssignPermissions, GetUserPermissions, GetMenuTree (filtered by user role)
- [ ] T129 [US7] Implement admin user/role/permission handler in `internal/admin/handler/rbac.go` — user CRUD, role CRUD, permission list, menu tree per admin-api.yaml
- [ ] T130 [US7] Implement supplier data isolation in `internal/common/middleware/data_permission.go` — middleware that filters queries by supplier_id for supplier role users

### Frontend - Admin (Vue 3)

- [ ] T131 [US7] Create user management page at `admin-web/src/views/system/UserManage.vue` — user list (username/phone/role/status), create user dialog (info+role selection), freeze/unfreeze actions, password reset
- [ ] T132 [US7] Create role management page at `admin-web/src/views/system/RoleManage.vue` — role list, create/edit role dialog, permission assignment (tree checkbox for function+button+API permissions)
- [ ] T133 [US7] Create permission tree editor component in `admin-web/src/components/PermissionTree.vue` — tree structure with checkboxes, group by menu/button/API type, select all/none
- [ ] T133b [US7] Create MFA enrollment and verification components in `admin-web/src/components/MfaSetup.vue` and `admin-web/src/components/MfaVerify.vue` — TOTP QR code display for enrollment, 6-digit code input for verification, used before sensitive operations (refund approval, permission changes, data export) per FR-030

**Checkpoint**: Admin can create users with roles, users see only authorized menus, supplier data isolation works.

---

## Phase 10: Frontend Enhancement

**Purpose**: Polish UI, add homepage content, improve UX

- [ ] T134 [P] Implement homepage banner management API in `internal/admin/handler/banner.go` — CRUD for banners with image upload, link target, position, sort order, status, expiry date
- [ ] T135 [P] Implement popular destination recommendation API in `internal/product/handler/destination.go` — list destinations by category (domestic/popular), with product count and starting price
- [ ] T136 [US2] Enhance homepage at `web/pages/index.vue` — connect banner carousel to API, connect popular destinations to API, add search autocomplete with debounce
- [ ] T137 [US4] Enhance personal center at `web/pages/user/index.vue` — member level display, order count badges, quick actions (my orders/my travellers/real-name)
- [ ] T138 [P] Implement image upload service in `internal/common/service/upload.go` — generate STS token for OSS upload, file format validation (jpg/png, ≤5MB), thumbnail generation
- [ ] T139 [P] Add responsive design adjustments to web pages for mobile browser compatibility in `web/assets/responsive.css` and component-level media queries

**Checkpoint**: Homepage shows dynamic content, search autocomplete works, personal center is complete, responsive on mobile

---

## Phase 11: Integration Testing & Security Hardening

**Purpose**: End-to-end verification, security compliance, performance baseline

- [ ] T140 Write integration test for user registration→login→real-name flow in `tests/integration/user_test.go`
- [ ] T141 Write integration test for product listing→detail→booking→payment flow in `tests/integration/booking_test.go`
- [ ] T142 Write integration test for order→refund flow with cancellation rule calculation in `tests/integration/refund_test.go`
- [ ] T143 Write integration test for payment callback idempotency in `tests/integration/payment_test.go`
- [ ] T144 [P] Verify TLS 1.3 configuration in Traefik config, ensure all HTTP redirects to HTTPS
- [ ] T145 [P] Verify AES-256-GCM field encryption for id_card_no and phone in user_account and order_traveller tables — check encrypted storage and masked API responses
- [ ] T146 [P] Verify audit log coverage — ensure all POST/PUT/DELETE operations generate audit_log entries with correct fields
- [ ] T147 [P] Verify password policy — Argon2id hashing, 8+ chars with complexity, 90-day expiry, 5-failure lockout
- [ ] T148 Run quickstart.md validation scenarios VS1-VS8 and verify all pass

**Checkpoint**: All integration tests pass, security checklist complete, quickstart scenarios verified

---

## Phase 12: Deployment & Operations

**Purpose**: Production deployment setup

- [ ] T149 Create WinSW service configuration `deploy/winsw/travel-api-service.xml` with auto-start, auto-restart on failure, log capture
- [ ] T150 Create Traefik configuration `deploy/traefik/traefik.yml` with TLS 1.3, route rules for API/Web/Admin, rate limiting, health checks
- [ ] T151 Create GitHub Actions CI/CD pipeline `.github/workflows/ci.yml` — Go build (CGO_ENABLED=0), lint (golangci-lint), test, frontend build, artifact upload
- [ ] T152 Create deployment script `deploy/scripts/deploy.ps1` — download artifact, stop service, replace binary, start service, health check verification
- [ ] T153 [P] Create Prometheus metrics endpoint in `internal/common/middleware/metrics.go` — HTTP request duration, count, error rate, business metrics (order count, payment success rate)
- [ ] T154 [P] Create database backup script `deploy/scripts/backup.ps1` — pg_basebackup daily, WAL archival every 15min, encrypted storage
- [ ] T155 [P] Create Grafana dashboard configuration `deploy/grafana/dashboards/` — infrastructure dashboard (CPU/memory/disk/network per node), application dashboard (QPS/P99 latency/error rate/goroutine count per service), business dashboard (order volume/payment success rate/refund rate/active users) per PRD §10.4.2
- [ ] T156 [P] Create Prometheus alerting rules `deploy/prometheus/alert-rules.yml` — P1 alerts (DB connection >80%, replication lag >5s, error rate >1% for 5min, payment success <95% for 10min) with phone+SMS+DingTalk notification; P2 alerts (CPU >80% for 10min, memory >85% for 10min, disk >80%) with SMS+DingTalk notification per PRD §10.4.2 table 9-5
- [ ] T157 [P] Create Windows Exporter installation script `deploy/scripts/setup-exporter.ps1` — install Windows Exporter (MSI, port 9182) for system-level metrics collection by Prometheus

**Checkpoint**: Service runs as Windows service, Traefik routes correctly, CI/CD pipeline builds and deploys, monitoring operational

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 - BLOCKS all user stories
- **Phase 3-9 (User Stories)**: All depend on Phase 2 completion
  - US1 (Login) should complete first as other stories depend on auth
  - US2 (Products) can start after US1 auth is ready
  - US3 (Booking) depends on US1 (auth) + US2 (products)
  - US4 (Orders) depends on US3 (booking creates orders)
  - US5-US7 (Admin) can proceed in parallel after Phase 2
- **Phase 10 (Enhancement)**: Depends on US2 (homepage) + US4 (personal center)
- **Phase 11 (Testing)**: Depends on all user stories complete
- **Phase 12 (Deployment)**: Depends on Phase 11

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2 — No other story dependencies
- **US2 (P1)**: Can start after Phase 2 — Uses auth from US1
- **US3 (P1)**: Depends on US1 (auth) + US2 (product detail for booking entry)
- **US4 (P2)**: Depends on US3 (orders created by booking)
- **US5 (P2)**: Can start after Phase 2 — Independent admin functionality
- **US6 (P2)**: Can start after Phase 2 — Uses orders from US3
- **US7 (P3)**: Can start after Phase 2 — Independent admin functionality

### Parallel Opportunities

- Phase 1: T002-T006 can all run in parallel; T013-T015 can all run in parallel
- Phase 2: T022-T026 can all run in parallel; T028-T036 can all run in parallel
- Phase 3: T048-T049, T053-T054, T055-T056 can run in parallel (different frontends)
- Phase 4: T069-T070, T072 can run in parallel (different frontends)
- Phase 9 (US7) can run in parallel with Phase 5-8 (different domain)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: US1 User Registration & Login
4. **STOP and VALIDATE**: User can register, login, verify identity on Web + Mini Program
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (Login) → Test → Deploy (users can register/login)
3. Add US2 (Products) → Test → Deploy (users can browse products)
4. Add US3 (Booking) → Test → Deploy (users can book and pay) — **MVP COMPLETE**
5. Add US4 (Orders) → Test → Deploy (users can manage orders)
6. Add US5-US7 (Admin) → Test → Deploy (full admin capabilities)
7. Enhancement + Testing + Deployment → Production ready

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Each user story includes both backend and frontend tasks (non-negotiable)
- All amounts stored as integers (cents) to avoid floating-point issues
- Sensitive fields (id_card_no, phone) encrypted with AES-256-GCM before storage
- API responses mask sensitive fields (phone: 138****8000, id_card: 110***********1234)
- Payment callbacks must be idempotent (DB unique constraint + Redis dedup)
- Inventory uses Redis atomic decrement + PostgreSQL SELECT FOR UPDATE
