# Quickstart: 境内跟团游 MVP

## Prerequisites

- Go 1.24+ (1.26+ recommended for development)
- PostgreSQL 17.x
- Redis 7.2+
- Node.js 20+ (for frontend projects)
- Git

## Setup

### Backend

```bash
# Clone and setup
git clone <repo-url>
cd cstsportstravel
cp configs/config.yaml configs/config.local.yaml
# Edit configs/config.local.yaml with your DB/Redis credentials

# Create database
createdb travel_booking_dev

# Run migrations
go run cmd/server/main.go migrate up

# Start the server
go run cmd/server/main.go
# Server starts on https://localhost:8443 (TLS required)
```

### Frontend - Web Sales Platform

```bash
cd web
npm install
npm run dev
# Runs on http://localhost:3000
```

### Frontend - Admin System

```bash
cd admin-web
npm install
npm run dev
# Runs on http://localhost:5173
```

### Frontend - WeChat Mini Program

```bash
cd miniapp
npm install
# Use WeChat Developer Tools to open the project directory
# Configure appid in manifest.json
```

## Seed Data

```bash
# Seed initial admin user (admin/admin123)
go run cmd/server/main.go seed

# Seed sample product data for testing
go run cmd/server/main.go seed --products
```

## Validation Scenarios

### VS1: User Registration and Login

Validates FR-001, FR-006. User can register via phone + SMS code and login.

```bash
# Step 1: Request SMS code
curl -X POST https://localhost:8443/api/v1/auth/sms-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'
# Expected: {"code": 0, "message": "success", "data": {"expires_in": 300}}

# Step 2: Login with code (use code from SMS or test mode returns it directly)
curl -X POST https://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000", "code": "123456"}'
# Expected: {"code": 0, "data": {"user": {...}, "access_token": "eyJ...", "refresh_token": "eyJ..."}}

# Step 3: Verify token works
curl https://localhost:8443/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
# Expected: {"code": 0, "data": {"id": 1, "phone": "138****8000", ...}}
```

### VS2: Real-Name Verification

Validates FR-003. User submits real-name info for verification.

```bash
# Submit real-name verification
curl -X POST https://localhost:8443/api/v1/users/me/real-name \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"real_name": "张三", "id_card_no": "110101199001011234"}'
# Expected: {"code": 0, "data": {"status": "verified"}}

# Verify user profile shows verified status
curl https://localhost:8443/api/v1/users/me \
  -H "Authorization: Bearer <token>"
# Expected: real_name_status = "verified"
```

### VS3: Product Browsing

Validates FR-010, FR-011, FR-012. User can search and view products.

```bash
# Step 1: List products with filters
curl "https://localhost:8443/api/v1/products?destination=云南&days_min=5&days_max=7&sort=price_asc&page=1&page_size=10"
# Expected: {"code": 0, "data": {"items": [...], "total": 25, "page": 1}}

# Step 2: View product detail
curl https://localhost:8443/api/v1/products/1
# Expected: Full product info with itinerary, fees, cancellation rules

# Step 3: View departure calendar
curl "https://localhost:8443/api/v1/products/1/departures?month=2026-07"
# Expected: {"code": 0, "data": [{"date": "2026-07-01", "adult_price": 399900, "status": "open", ...}, ...]}

# Step 4: Search autocomplete
curl "https://localhost:8443/api/v1/products/search/suggest?q=丽江"
# Expected: {"code": 0, "data": ["丽江古城", "丽江玉龙雪山", ...]}
```

### VS4: Complete Booking Flow

Validates FR-014, FR-015, FR-016. End-to-end booking from product to payment.

```bash
# Step 1: Create order (select departure + travellers)
curl -X POST https://localhost:8443/api/v1/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "departure_id": 10,
    "adult_count": 2,
    "child_count": 1,
    "infant_count": 0,
    "travellers": [
      {"real_name": "张三", "id_card_no": "110101199001011234", "phone": "13800138000", "birth_date": "1990-01-01", "gender": "male"},
      {"real_name": "李四", "id_card_no": "110101199202022345", "phone": "13900139000", "birth_date": "1992-02-02", "gender": "female"},
      {"real_name": "张小三", "id_card_no": "110101202001013456", "birth_date": "2020-01-01", "gender": "male", "is_child": true, "linked_adult_traveller_index": 0}
    ],
    "contact_name": "张三",
    "contact_phone": "13800138000"
  }'
# Expected: {"code": 0, "data": {"order_id": 1, "order_no": "ORD-20260619-143022-0001", "payable_amount": 799800, "expire_at": "2026-06-19T15:00:22Z"}}

# Step 2: Create payment
curl -X POST https://localhost:8443/api/v1/payments/create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"order_id": 1, "channel": "alipay", "method": "h5"}'
# Expected: {"code": 0, "data": {"payment_id": 1, "pay_url": "https://openapi.alipay.com/..."}}

# Step 3: Simulate payment callback (test mode)
curl -X POST https://localhost:8443/api/v1/test/payments/simulate-callback \
  -H "Content-Type: application/json" \
  -d '{"payment_id": 1, "status": "paid"}'
# Expected: Payment processed, order status updated

# Step 4: Verify order status
curl https://localhost:8443/api/v1/orders/1 \
  -H "Authorization: Bearer <token>"
# Expected: order_status = "paid_full" or "pending_travel"
```

### VS5: Payment Timeout Auto-Cancel

Validates FR-016. Order auto-cancels after 30 minutes of non-payment.

```bash
# Create order but do not pay
curl -X POST https://localhost:8443/api/v1/orders \
  -H "Authorization: Bearer <token>" \
  -d '{...}'
# Wait 30 minutes (or use test mode to fast-forward)

# Check order status
curl https://localhost:8443/api/v1/orders/2 \
  -H "Authorization: Bearer <token>"
# Expected: order_status = "cancelled", cancel_reason = "payment_timeout"

# Verify inventory released - departure available stock should be restored
curl "https://localhost:8443/api/v1/products/1/departures?month=2026-07"
# Expected: Stock count reflects released seats
```

### VS6: Refund Flow

Validates FR-018, FR-019, FR-020, FR-021. User requests refund, admin approves.

```bash
# Step 1: User submits refund request
curl -X POST https://localhost:8443/api/v1/orders/1/refund \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"reason": "行程变更", "description": "因个人原因无法出行"}'
# Expected: {"code": 0, "data": {"refund_id": 1, "refund_amount": 639840, "status": "pending"}}

# Step 2: Admin views pending refunds
curl "https://localhost:8443/api/v1/admin/refunds?status=pending" \
  -H "Authorization: Bearer <admin_token>"
# Expected: List of pending refunds with calculated amounts

# Step 3: Admin approves refund (amount <= 1000, operator can approve directly)
curl -X PUT https://localhost:8443/api/v1/admin/refunds/1/approve \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"note": "审核通过"}
# Expected: {"code": 0, "data": {"status": "approved"}}

# Step 4: Verify refund processed
curl https://localhost:8443/api/v1/orders/1 \
  -H "Authorization: Bearer <token>"
# Expected: order_status = "refunded"
```

### VS7: Admin Product Management

Validates FR-007, FR-008. Supplier creates product, operator reviews.

```bash
# Step 1: Supplier creates product
curl -X POST https://localhost:8443/api/v1/admin/products \
  -H "Authorization: Bearer <supplier_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "云南丽江大理5日游",
    "category_id": 1,
    "origin_city": "上海",
    "destination_cities": ["丽江", "大理"],
    "days": 5,
    "nights": 4,
    "transport_mode": "flight",
    "product_grade": "comfort",
    "fee_included": "往返机票、酒店住宿、景点门票、导游服务",
    "fee_excluded": "个人消费、自费项目"
  }'
# Expected: {"code": 0, "data": {"id": 1, "product_no": "DOM-DOM-20260619-0001", "status": "draft"}}

# Step 2: Supplier submits for review
curl -X POST https://localhost:8443/api/v1/admin/products/1/submit-review \
  -H "Authorization: Bearer <supplier_token>"
# Expected: {"code": 0, "data": {"status": "pending_review"}}

# Step 3: Operator approves product
curl -X PUT https://localhost:8443/api/v1/admin/products/1/approve \
  -H "Authorization: Bearer <operator_token>" \
  -H "Content-Type: application/json" \
  -d '{"note": "产品信息完整，审核通过"}'
# Expected: {"code": 0, "data": {"status": "approved"}}

# Step 4: Verify product visible on C-side
curl "https://localhost:8443/api/v1/products?keyword=丽江"
# Expected: Product appears in search results
```

### VS8: RBAC Permission Control

Validates FR-026, FR-027. Different roles see different menus and data.

```bash
# Step 1: Admin creates supplier account with limited role
curl -X POST https://localhost:8443/api/v1/admin/users \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"username": "supplier01", "real_name": "供应商A", "phone": "13700137000", "role_ids": [3], "supplier_id": 1}'
# Expected: {"code": 0, "data": {"id": 2, "username": "supplier01", "must_change_password": true}}

# Step 2: Supplier logs in and sees only their products
curl -X POST https://localhost:8443/api/v1/auth/admin/login \
  -d '{"username": "supplier01", "password": "initial_password"}'
# Then query products - should only see supplier_id=1 products
curl "https://localhost:8443/api/v1/admin/products" \
  -H "Authorization: Bearer <supplier_token>"
# Expected: Only products with supplier_id = 1

# Step 3: Supplier tries to access unauthorized endpoint
curl "https://localhost:8443/api/v1/admin/users" \
  -H "Authorization: Bearer <supplier_token>"
# Expected: {"code": 403, "message": "permission denied"}
```

## Important Notes

1. **TLS Required**: The backend server enforces TLS 1.3. For local development, a self-signed certificate is auto-generated. Use `--insecure` flag with curl or configure your HTTP client to skip certificate verification.

2. **Test Mode**: When `app.env=test` in config, SMS codes are returned directly in the API response (no actual SMS sent), and payment callbacks can be simulated via the test endpoint.

3. **Amount Format**: All amounts in API requests/responses are in **yuan with 2 decimal places** (e.g., `199.99`). Internally, the system stores amounts as integer cents. The API layer handles conversion.

4. **Date Format**: All dates follow ISO 8601 format: `YYYY-MM-DD` for dates, `YYYY-MM-DDTHH:mm:ssZ` for timestamps.

5. **Pagination**: List endpoints support `page` (1-based) and `page_size` (default 20, max 100) parameters. Response includes `total` count and `items` array.
