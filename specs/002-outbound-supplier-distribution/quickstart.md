# Quickstart Validation Guide: 一期扩展 — 出境游 + 供应商开放平台 + 分销体系

**Date**: 2026-06-30
**Feature**: specs/002-outbound-supplier-distribution/spec.md

## Prerequisites

- Go 1.26+ 开发环境
- PostgreSQL 18+ 数据库（已执行全部迁移脚本）
- Redis/Memurai 7.2+
- NATS 2.11+（JetStream 已启用）
- Meilisearch 1.19+
- Consul 1.22+
- Traefik 3.x+
- Node.js 18+（前端构建）
- 银联测试商户号（77 开头）+ 测试证书
- 抖音开放平台 AppID（抖音小程序测试）
- 百度/腾讯 OCR API Key（护照识别）

## Setup

```bash
# 1. 启动基础设施
consul agent -dev &
nats-server -js &
meilisearch &

# 2. 执行数据库迁移
psql -U postgres -d travel_booking -f backend/migrations/002_outbound_tables.sql
psql -U postgres -d travel_booking -f backend/migrations/003_supplier_tables.sql
psql -U postgres -d travel_booking -f backend/migrations/004_distribution_tables.sql
psql -U postgres -d travel_booking -f backend/migrations/005_visa_tables.sql
psql -U postgres -d travel_booking -f backend/migrations/006_marketing_tables.sql
psql -U postgres -d travel_booking -f backend/migrations/007_payment_extension.sql

# 3. 启动后端服务
cd backend
go run cmd/user-service/main.go &
go run cmd/product-service/main.go &
go run cmd/order-service/main.go &
go run cmd/payment-service/main.go &
go run cmd/distribution-service/main.go &

# 4. 启动前端
cd frontend/web && npm run dev &
cd frontend/admin && npm run dev &
cd frontend/miniprogram && npm run dev:mp-weixin &
```

## Validation Scenarios

### Scenario 1: 出境游产品浏览与签证信息

**Objective**: 验证出境游产品列表筛选、签证信息卡片展示

**Steps**:
1. 创建出境游产品（含签证信息配置）
2. 审核通过并上架
3. 访问出境游产品列表 API（带筛选参数）
4. 访问产品详情 API（验证签证信息卡片）

**Commands**:
```bash
# 创建出境游产品
curl -X POST http://localhost:8080/api/v2/supplier/products \
  -H "Authorization: Bearer $SUPPLIER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "productName": "日本东京6日游",
    "productType": "outbound_group",
    "destinationCountryId": 1,
    "days": 6,
    "visaInfo": {
      "visaType": "visa_required",
      "processingDays": 7,
      "fee": 500,
      "materialPreview": "护照原件、照片、在职证明..."
    }
  }'

# 审核通过
curl -X POST http://localhost:8080/api/v2/admin/products/$PRODUCT_ID/approve \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# 查询出境游产品列表（按签证类型筛选）
curl "http://localhost:8080/api/v2/products/outbound?visaType=visa_required&page=1&pageSize=10" \
  -H "Authorization: Bearer $USER_TOKEN"

# 查询产品详情（含签证信息卡片）
curl "http://localhost:8080/api/v2/products/outbound/$PRODUCT_ID" \
  -H "Authorization: Bearer $USER_TOKEN"
```

**Expected**:
- 产品列表返回签证类型标签（visa_required）
- 产品详情包含签证信息卡片（visaInfo 字段非空）
- 行前信息服务数据可查询

---

### Scenario 2: 出境游预订与签证代办

**Objective**: 验证出境游预订五步向导、护照校验、签证订单创建

**Steps**:
1. 用户选择出境游团期
2. 填写出游人护照信息（含 OCR 识别）
3. 选择签证代办服务
4. 确认订单并支付
5. 验证签证订单自动创建

**Commands**:
```bash
# 提交出境游预订（含护照信息）
curl -X POST http://localhost:8080/api/v2/orders/outbound \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "productId": '$PRODUCT_ID',
    "departureId": '$DEPARTURE_ID',
    "adultCount": 2,
    "childCount": 0,
    "travellers": [
      {
        "name": "张三",
        "namePinyin": "ZHANG SAN",
        "passportNumber": "E12345678",
        "passportExpiry": "2028-06-30",
        "passportIssuePlace": "北京",
        "nationality": "中国"
      }
    ],
    "visaServiceRequired": true,
    "contactName": "张三",
    "contactPhone": "13800138000"
  }'

# 支付订单
curl -X POST http://localhost:8080/api/v2/payments/create \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "orderNo": "'$ORDER_NO'",
    "channel": "alipay",
    "method": "pc"
  }'

# 模拟支付回调
curl -X POST http://localhost:8080/api/v2/payments/notify/alipay \
  -d "out_trade_no=$ORDER_NO&trade_status=TRADE_SUCCESS&sign=..."

# 查询签证订单
curl "http://localhost:8080/api/v2/orders/$ORDER_NO/visa" \
  -H "Authorization: Bearer $USER_TOKEN"
```

**Expected**:
- 护照有效期不足时返回错误（code: PASSPORT_EXPIRY_INSUFFICIENT）
- 支付成功后自动创建签证订单（status: pending_submit）
- 签证订单关联主订单

---

### Scenario 3: 供应商入驻全流程

**Objective**: 验证供应商入驻申请→审核→合同签署→工作台开通

**Steps**:
1. 供应商提交入驻申请
2. 运营专员初审通过
3. 运营主管复审通过
4. 电子合同签署
5. 供应商登录工作台

**Commands**:
```bash
# 提交入驻申请
curl -X POST http://localhost:8080/api/v2/suppliers/apply \
  -F "companyName=测试旅行社" \
  -F "creditCode=91110000MA12345678" \
  -F "businessLicense=@license.pdf" \
  -F "legalPersonName=李四" \
  -F "legalPersonIdCard=110101199001011234" \
  -F "contactName=王五" \
  -F "contactPhone=13900139000"

# 查询申请进度
curl "http://localhost:8080/api/v2/suppliers/apply/APP-20260630-0001"

# 初审通过
curl -X POST http://localhost:8080/api/v2/admin/suppliers/applications/$APP_ID/audit \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}'

# 复审通过
curl -X POST http://localhost:8080/api/v2/admin/suppliers/applications/$APP_ID/audit \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}'

# 供应商登录工作台
curl -X POST http://localhost:8080/api/v2/auth/supplier/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13900139000", "code": "123456"}'
```

**Expected**:
- 申请提交后返回申请编号（APP-YYYYMMDD-NNNN）
- 初审/复审状态正确流转
- 复审通过后供应商可登录工作台

---

### Scenario 4: 二级分销全流程

**Objective**: 验证分销商入驻→推广链接生成→佣金计算→提现

**Steps**:
1. 分销商提交入驻申请
2. 审核通过并签署协议
3. 生成推广链接
4. 消费者通过推广链接下单
5. 验证佣金计算
6. 佣金解冻后申请提现

**Commands**:
```bash
# 分销商入驻申请
curl -X POST http://localhost:8080/api/v2/distributors/apply \
  -F "distributorType=personal" \
  -F "realName=赵六" \
  -F "idCardNumber=110101199002021234" \
  -F "idCardFront=@front.jpg" \
  -F "idCardBack=@back.jpg" \
  -F "phone=13700137000" \
  -F "bankName=中国工商银行" \
  -F "bankAccountName=赵六" \
  -F "bankAccountNumber=6222021234567890123"

# 审核通过
curl -X POST http://localhost:8080/api/v2/admin/distributors/$DIST_ID/audit \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}'

# 分销商签署协议后登录
curl -X POST http://localhost:8080/api/v2/auth/distributor/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13700137000", "code": "123456"}'

# 生成推广链接
curl -X POST http://localhost:8080/api/v2/distributor/promotion-links \
  -H "Authorization: Bearer $DIST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"productId": '$PRODUCT_ID'}'

# 模拟消费者通过推广链接下单
curl -X POST http://localhost:8080/api/v2/orders \
  -H "Authorization: Bearer $CONSUMER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "productId": '$PRODUCT_ID',
    "departureId": '$DEPARTURE_ID',
    "adultCount": 2,
    "distributorCode": "'$DIST_CODE'",
    ...
  }'

# 查询佣金明细
curl "http://localhost:8080/api/v2/distributor/commissions" \
  -H "Authorization: Bearer $DIST_TOKEN"
```

**Expected**:
- 推广链接生成成功（含短链接和二维码）
- 消费者下单后佣金自动计算（status: pending）
- 佣金冻结期满后自动变为可提现

---

### Scenario 5: 银联支付

**Objective**: 验证银联网关支付和回调处理

**Steps**:
1. 创建订单
2. 选择银联支付
3. 模拟银联回调
4. 验证订单状态更新

**Commands**:
```bash
# 创建银联支付订单
curl -X POST http://localhost:8080/api/v2/payments/create \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "orderNo": "'$ORDER_NO'",
    "channel": "unionpay",
    "method": "gateway"
  }'

# 模拟银联回调（后台通知）
curl -X POST http://localhost:8080/api/v2/payments/notify/unionpay \
  -d "orderId=$ORDER_NO&txnAmt=50000&respCode=00&sign=..."
```

**Expected**:
- 支付创建返回银联网关跳转 URL
- 回调处理后订单状态更新为已付款

---

### Scenario 6: 定金+尾款支付

**Objective**: 验证定金支付→尾款提醒→尾款支付→逾期取消

**Steps**:
1. 创建支持定金+尾款的订单
2. 支付定金
3. 验证尾款提醒
4. 支付尾款
5. 测试逾期取消

**Commands**:
```bash
# 创建定金+尾款订单
curl -X POST http://localhost:8080/api/v2/orders/outbound \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "productId": '$PRODUCT_ID',
    "departureId": '$DEPARTURE_ID',
    "adultCount": 2,
    "paymentMode": "deposit",
    ...
  }'

# 支付定金
curl -X POST http://localhost:8080/api/v2/payments/create \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"orderNo": "'$ORDER_NO'", "channel": "alipay", "method": "pc", "paymentType": "deposit"}'

# 支付尾款
curl -X POST http://localhost:8080/api/v2/payments/create \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"orderNo": "'$ORDER_NO'", "channel": "alipay", "method": "pc", "paymentType": "balance"}'
```

**Expected**:
- 定金支付成功后订单状态为"已付定金"
- 尾款支付成功后订单状态为"已付全款"
- 逾期未付尾款自动取消

---

### Scenario 7: 优惠券领取与使用

**Objective**: 验证优惠券创建→领取→下单使用→退款退回

**Commands**:
```bash
# 创建优惠券
curl -X POST http://localhost:8080/api/v2/admin/marketing/coupons \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "couponName": "新用户满减券",
    "couponType": "full_reduction",
    "discountAmount": 100,
    "minConsumption": 500,
    "totalStock": 1000,
    "perUserLimit": 1,
    "validityType": "fixed",
    "validFrom": "2026-07-01T00:00:00Z",
    "validTo": "2026-12-31T23:59:59Z",
    "applicableScope": "all"
  }'

# 用户领取优惠券
curl -X POST http://localhost:8080/api/v2/coupons/$COUPON_ID/claim \
  -H "Authorization: Bearer $USER_TOKEN"

# 下单时使用优惠券
curl -X POST http://localhost:8080/api/v2/orders \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "couponClaimId": '$CLAIM_ID',
    ...
  }'
```

**Expected**:
- 优惠券领取后状态变为 available
- 下单使用后状态变为 used
- 退款后优惠券状态变为 returned

---

## Performance Validation

```bash
# 使用 wrk 或 k6 进行压测
# 产品列表接口
wrk -t12 -c400 -d30s http://localhost:8080/api/v2/products/outbound

# 订单确认接口
wrk -t12 -c400 -d30s -s order_create.lua http://localhost:8080/api/v2/orders
```

**Expected**:
- 产品列表 P99 ≤ 1.0s
- 订单确认 P99 ≤ 500ms
- 全站 QPS ≥ 10,000

---

## References

- [Feature Spec](./spec.md)
- [Data Model](./data-model.md)
- [Supplier API Contract](./contracts/supplier-api.yaml)
- [Distribution API Contract](./contracts/distribution-api.yaml)
- [Research Notes](./research.md)
