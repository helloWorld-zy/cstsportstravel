# Tasks: 一期扩展 — 出境游 + 供应商开放平台 + 分销体系

**Input**: Design documents from `specs/002-outbound-supplier-distribution/`

**Prerequisites**: plan.md (required), spec.md (required), data-model.md, contracts/, research.md

**Organization**: Tasks grouped by phase/user story per plan.md §Implementation Phases

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to

### User Story Tag Mapping (spec.md → tasks.md)

| spec.md User Story | tasks.md Tag | Phase |
|-------------------|--------------|-------|
| US1 出境游产品浏览与签证信息查询 | US1 | Phase 2 |
| US2 出境游预订与签证代办 | US1 | Phase 2 |
| US3 供应商入驻与管理 | US2 | Phase 3 |
| US4 二级分销体系 | US3 | Phase 4 |
| US5 支付扩展（银联+定金尾款） | US4 | Phase 5 |
| US6 营销系统（优惠券+促销活动） | US5 | Phase 6 |
| US7 财务管理增强 | — | Phase 5/6 (embedded) |
| US8 抖音小程序 | — | Phase 8 |

---

## Phase 1: 服务拆分基础设施 (Week 1-2)

**Purpose**: 从 Gin 单体拆分为 5 个微服务，搭建 NATS/Consul/Meilisearch/Traefik

- [X] T001 [P] Deploy Consul dev cluster and create Go client wrapper in `internal/shared/consul/client.go`
- [X] T002 [P] Deploy NATS with JetStream enabled and create Go client wrapper in `internal/shared/nats/client.go`
- [X] T003 [P] Deploy Meilisearch and create Go client wrapper in `internal/shared/meili/client.go`
- [X] T004 Create Traefik config with dynamic service routing in `infra/traefik/traefik.yml` and `infra/traefik/dynamic.yml`
- [X] T005 [P] Define NATS subjects, event DTOs, and pub/sub wrappers in `internal/shared/event/events.go`
- [X] T006 [P] Create database migration for outbound product extensions in `migrations/002_outbound_tables.sql` (country, visa_material_template, product columns)
- [X] T007 [P] Create database migration for supplier tables in `migrations/003_supplier_tables.sql` (supplier, supplier_qualification, settlement_statement, commission_rule)
- [X] T008 [P] Create database migration for distribution tables in `migrations/004_distribution_tables.sql` (distributor, distributor_relation, promotion_link, commission_detail, withdrawal_record, promotion_click)
- [X] T009 [P] Create database migration for visa tables in `migrations/005_visa_tables.sql` (visa_order, visa_material, visa_progress)
- [X] T010 [P] Create database migration for marketing tables in `migrations/006_marketing_tables.sql` (coupon, coupon_claim, promotion_activity)
- [X] T011 [P] Create database migration for payment extension columns in `migrations/007_payment_extension.sql` (main_order/payment_transaction new columns)
- [X] T012 Split user-service from monolith: create `cmd/user-service/main.go` with WinSW XML config in `infra/winsw/user-service.xml`
- [X] T013 Split product-service from monolith: create `cmd/product-service/main.go` with WinSW XML config in `infra/winsw/product-service.xml`
- [X] T014 Split order-service from monolith: create `cmd/order-service/main.go` with WinSW XML config in `infra/winsw/order-service.xml`
- [X] T015 Split payment-service from monolith: create `cmd/payment-service/main.go` with WinSW XML config in `infra/winsw/payment-service.xml`
- [X] T016 Create distribution-service skeleton: `cmd/distribution-service/main.go` with health check endpoint and WinSW config
- [X] T017 Create OpenAPI v2 aggregation document in `api/openapi/v2.yaml` referencing all sub-specs
- [X] T018 Create shared middleware package (auth, rate-limit, audit, tenant isolation) in `internal/shared/middleware/`
- [X] T019 Create unified error codes package in `internal/shared/errors/codes.go`
- [X] T020 Create AES-256-GCM field encryption utility in `internal/shared/encryption/aes_gcm.go`

**Checkpoint**: All 5 services can start independently, register with Consul, and communicate via NATS

---

## Phase 2: 出境游产品与预订 (Week 3-5) [US1]

**Goal**: 出境游产品浏览、预订五步向导、签证服务闭环（15个功能点）

**Independent Test**: 用户可从出境游产品列表→详情→预订→支付→签证材料提交→进度查询全流程

### Implementation

- [ ] T021 [P] [US1] Create Country domain model in `backend/internal/product/domain/country.go` with visa_type enum and hierarchy
- [ ] T022 [P] [US1] Create VisaInfo domain model in `backend/internal/product/domain/visa_info.go` with visa types, processing days, material preview
- [ ] T023 [P] [US1] Create VisaMaterialTemplate domain model in `backend/internal/product/domain/visa_material_template.go` with occupation-based templates
- [ ] T024 [US1] Extend Product model with outbound fields (destination_country_id, visa_info, insurance_requirements) in `backend/internal/product/domain/outbound.go`
- [ ] T025 [P] [US1] Create Country repository with CRUD and hierarchy query in `backend/internal/product/repository/country_repo.go`
- [ ] T026 [P] [US1] Create VisaMaterialTemplate repository with country+occupation query in `backend/internal/product/repository/visa_template_repo.go`
- [ ] T027 [US1] Create outbound product listing API with continent/country/visa_type/city/days filters in `backend/internal/product/handler/outbound_handler.go`
- [ ] T028 [US1] Create outbound product detail API with visa info card in `backend/internal/product/handler/outbound_handler.go`
- [ ] T029 [US1] Create pre-trip service API (entry policy, materials, cash rules, customs guide, emergency contacts) in `backend/internal/product/handler/pretrip_handler.go`
- [ ] T030 [P] [US1] Create PassportInfo domain model with expiry validation (≥6 months after return) in `backend/internal/order/domain/passport.go`
- [ ] T031 [P] [US1] Create OCR adapter interface and Baidu OCR implementation for passport recognition in `backend/internal/order/service/ocr_adapter.go`
- [ ] T032 [US1] Create passport management API (CRUD + OCR + expiry validation) in `backend/internal/order/handler/passport_handler.go`
- [ ] T033 [US1] Create outbound booking API (5-step wizard: select departure→fill passport→visa service→addons→confirm) in `backend/internal/order/handler/outbound_booking_handler.go`
- [ ] T034 [P] [US1] Create VisaOrder domain model with 5-node state machine (pending_submit→reviewing→submitted→approved/rejected) in `backend/internal/order/domain/visa_order.go`
- [ ] T035 [P] [US1] Create VisaMaterial domain model in `backend/internal/order/domain/visa_material.go`
- [ ] T036 [P] [US1] Create VisaProgress domain model in `backend/internal/order/domain/visa_progress.go`
- [ ] T037 [US1] Create VisaOrder repository with status machine transitions in `backend/internal/order/repository/visa_order_repo.go`
- [ ] T038 [US1] Create VisaMaterial repository with file storage integration in `backend/internal/order/repository/visa_material_repo.go`
- [ ] T039 [US1] Create visa material management API (upload ≤10MB, occupation-based checklist, completeness check) in `backend/internal/order/handler/visa_material_handler.go`
- [ ] T040 [US1] Create visa progress tracking API with NATS event publishing on status change in `backend/internal/order/handler/visa_progress_handler.go`
- [ ] T041 [US1] Create visa notification NATS consumer (SMS + in-app for status change/approval/rejection/logistics) in `backend/internal/order/service/visa_notification_consumer.go`
- [ ] T042 [US1] Create visa history query API and visa expiry reminder (Asynq cron job, 90 days before expiry) in `backend/internal/order/service/visa_reminder.go`
- [ ] T043 [US1] Create outbound product list page with visa filters in `frontend/web/pages/outbound/index.vue`
- [ ] T044 [US1] Create outbound product detail page with visa info card in `frontend/web/pages/outbound/[id].vue`
- [ ] T045 [US1] Create outbound booking 5-step wizard page (passport form, OCR, visa service, addons, confirm) in `frontend/web/pages/outbound/booking.vue`
- [ ] T046 [US1] Create visa progress page with 5-node progress bar and material upload in `frontend/web/components/visa/VisaProgress.vue`
- [ ] T047 [US1] Create pre-trip service page (entry policy, entry card templates, flight tracking, checklist) in `frontend/web/pages/outbound/pretrip.vue`
- [ ] T048 [P] [US1] Create mini-program outbound list page in `frontend/miniprogram/pages/outbound/list.vue`
- [ ] T049 [P] [US1] Create mini-program outbound detail page in `frontend/miniprogram/pages/outbound/detail.vue`
- [ ] T050 [P] [US1] Create mini-program outbound booking page in `frontend/miniprogram/pages/outbound/booking.vue`
- [ ] T051 [P] [US1] Create mini-program visa progress page in `frontend/miniprogram/pages/visa/progress.vue`
- [ ] T052 [P] [US1] Create mini-program visa materials upload page in `frontend/miniprogram/pages/visa/materials.vue`
- [ ] T053 [US1] Create admin visa order list page in `frontend/admin/views/visa/VisaOrderList.vue`
- [ ] T054 [US1] Create admin visa material audit page in `frontend/admin/views/visa/VisaMaterialAudit.vue`

- [ ] T055 [US1] Create visa application form dynamic field generation service (fields vary by visa type/country) in `backend/internal/order/service/visa_form_service.go`

**Checkpoint**: 出境游产品浏览→预订→签证全流程可独立测试

---

## Phase 3: 供应商入驻与管理 (Week 4-6) [US2]

**Goal**: 供应商入驻申请→审核→合同签署→工作台产品/订单/结算全流程

**Independent Test**: 供应商可提交入驻申请，运营审核通过后登录工作台发布产品、查看订单、查看结算单

### Implementation

- [ ] T055 [P] [US2] Create Supplier domain model with status machine (pending/reviewing/active/suspended/terminated) in `backend/internal/supplier/domain/supplier.go`
- [ ] T056 [P] [US2] Create SupplierQualification domain model in `backend/internal/supplier/domain/supplier_qualification.go`
- [ ] T057 [P] [US2] Create SettlementStatement domain model with 5-step status (pending/confirmed/disputed/paid/archived) in `backend/internal/supplier/domain/settlement.go`
- [ ] T058 [P] [US2] Create CommissionRule domain model with scope types (global/category/supplier/product) in `backend/internal/supplier/domain/commission_rule.go`
- [ ] T059 [US2] Create Supplier repository with RLS tenant isolation in `backend/internal/supplier/repository/supplier_repo.go`
- [ ] T060 [US2] Create SettlementStatement repository in `backend/internal/supplier/repository/settlement_repo.go`
- [ ] T061 [US2] Create supplier application submission API (multipart form, file validation, application number generation APP-YYYYMMDD-NNNN) in `backend/internal/supplier/handler/application_handler.go`
- [ ] T062 [US2] Create supplier application status query API in `backend/internal/supplier/handler/application_handler.go`
- [ ] T063 [US2] Create supplier 2-level audit API (first review 2d, second review 1d, timeout alert via Asynq) in `backend/internal/supplier/handler/audit_handler.go`
- [ ] T064 [US2] Create e-contract generation service (template management, PDF generation, CA signing adapter, 7-day reminder + 30-day auto-close timeout via Asynq) in `backend/internal/supplier/service/contract_service.go`
- [ ] T065 [US2] Create supplier data isolation middleware (tenant_id + supplier_id, RLS policies) in `backend/internal/shared/middleware/supplier_isolation.go`
- [ ] T066 [US2] Create supplier product management API (publish/edit/toggle/status tracking) in `backend/internal/supplier/handler/workspace_product_handler.go`
- [ ] T067 [US2] Create supplier order handling API (list/confirm/reject/detail/refund) in `backend/internal/supplier/handler/workspace_order_handler.go`
- [ ] T068 [US2] Create supplier settlement 5-step flow service (auto-generate→review→confirm→pay→archive) in `backend/internal/supplier/service/settlement_service.go`
- [ ] T069 [US2] Create supplier settlement API (list/detail/confirm/dispute) in `backend/internal/supplier/handler/settlement_handler.go`
- [ ] T070 [US2] Create supplier withdrawal API (apply/approve/reject) in `backend/internal/supplier/handler/withdrawal_handler.go`
- [ ] T071 [US2] Create supplier commission rule configuration API (category/supplier level, tiered rates, approval flow) in `backend/internal/supplier/handler/commission_handler.go`
- [ ] T072 [US2] Create supplier statistics service (sales/revenue/review, Asynq pre-aggregation) in `backend/internal/supplier/service/statistics_service.go`
- [ ] T073 [P] [US2] Create supplier application page (multi-step form, auto-save) in `frontend/web/pages/supplier/apply.vue`
- [ ] T074 [P] [US2] Create admin supplier audit list page in `frontend/admin/views/supplier/AuditList.vue`
- [ ] T075 [P] [US2] Create admin supplier list page in `frontend/admin/views/supplier/SupplierList.vue`
- [ ] T076 [P] [US2] Create admin supplier detail page in `frontend/admin/views/supplier/SupplierDetail.vue`

**Checkpoint**: 供应商入驻→审核→工作台全流程可独立测试

---

## Phase 4: 二级分销体系 (Week 5-7) [US3]

**Goal**: 分销商入驻→推广链接→佣金计算→提现全流程，含防薅羊毛规则

**Independent Test**: 分销商可完成入驻→审核→协议签署→推广→佣金→提现全流程

### Implementation

- [ ] T077 [P] [US3] Create Distributor domain model with status (pending/active/frozen/cancelled/deactivated) and grade (normal/senior) in `backend/internal/distribution/domain/distributor.go`
- [ ] T078 [P] [US3] Create DistributorRelation domain model (parent_id, level 1 or 2) in `backend/internal/distribution/domain/distributor_relation.go`
- [ ] T079 [P] [US3] Create CommissionDetail domain model with status machine (pending→frozen→withdrawable→withdrawn/recovered) in `backend/internal/distribution/domain/commission.go`
- [ ] T080 [P] [US3] Create PromotionLink domain model with click tracking in `backend/internal/distribution/domain/promotion_link.go`
- [ ] T081 [P] [US3] Create WithdrawalRecord domain model in `backend/internal/distribution/domain/withdrawal.go`
- [ ] T082 [P] [US3] Create PromotionClick domain model with IP/device fingerprint in `backend/internal/distribution/domain/promotion_click.go`
- [ ] T083 [US3] Create Distributor repository with all CRUD and relationship queries in `backend/internal/distribution/repository/distributor_repo.go`
- [ ] T084 [US3] Create CommissionDetail repository with freeze/thaw batch operations in `backend/internal/distribution/repository/commission_repo.go`
- [ ] T085 [US3] Create PromotionLink repository with click statistics in `backend/internal/distribution/repository/promotion_link_repo.go`
- [ ] T086 [US3] Create distributor application API (personal/enterprise, auto-validate ID card/bank card/business license) in `backend/internal/distribution/handler/application_handler.go`
- [ ] T087 [US3] Create distributor audit API (approve/reject/supplement, distributor code generation 8-char) in `backend/internal/distribution/handler/audit_handler.go`
- [ ] T088 [US3] Create agreement signing API (record sign time + IP, status transition to active, 15-day activation timeout auto-reject via Asynq) in `backend/internal/distribution/handler/agreement_handler.go`
- [ ] T089 [US3] Create invitation mechanism API (invite link/invite code generation, parent_id binding) in `backend/internal/distribution/handler/invitation_handler.go`
- [ ] T090 [US3] Create promotion link generation API (short link + QR code with logo, 3 sizes) in `backend/internal/distribution/handler/promotion_handler.go`
- [ ] T091 [US3] Create promotion tracking service (URL param + Cookie 30-day dual tracking) in `backend/internal/distribution/service/tracking_service.go`
- [ ] T092 [US3] Create distribution order tracking (record distributor_id_l1/l2, promotion_code, track_source on order creation) in `backend/internal/distribution/handler/order_tracking_handler.go`
- [ ] T093 [US3] Create commission rule configuration API (3-level priority: product>category>global, level1/level2 rates, 5min cache refresh) in `backend/internal/distribution/handler/commission_rule_handler.go`
- [ ] T094 [US3] Create commission calculation engine (NATS consumer, base/ratio/attribution/cap rules, 50% cap enforcement) in `backend/internal/distribution/service/commission_service.go`
- [ ] T095 [US3] Create commission freeze/thaw service (T+N: domestic 7d, outbound 15d, cruise 15d; auto-thaw Asynq job) in `backend/internal/distribution/service/freeze_service.go`
- [ ] T096 [US3] Create commission refund recovery service (full/partial, freeze-in/out handling) in `backend/internal/distribution/service/recovery_service.go`
- [ ] T097 [US3] Create distributor withdrawal API (min 100 CNY, review, senior T+3 accelerated) in `backend/internal/distribution/handler/withdrawal_handler.go`
- [ ] T098 [US3] Create anti-fraud engine (self-purchase ban, identity isolation, device association 30d, IP rate limit 10/h, violation punishment) in `backend/internal/distribution/service/anti_fraud_service.go`
- [ ] T099 [US3] Create distributor grade service (auto upgrade/downgrade 90-day review Asynq job) in `backend/internal/distribution/service/grade_service.go`
- [ ] T100 [US3] Create distributor overview API (total/withdrawable/frozen commission, today stats, announcements) in `backend/internal/distribution/handler/overview_handler.go`
- [ ] T101 [US3] Create my-promotion API (product list, link management, click/order stats) in `backend/internal/distribution/handler/promotion_stats_handler.go`
- [ ] T102 [US3] Create my-team API (member list, team summary, invite, leaderboard, L1 only) in `backend/internal/distribution/handler/team_handler.go`
- [ ] T103 [US3] Create commission detail list API (filter by time/category/status, export Excel) in `backend/internal/distribution/handler/commission_detail_handler.go`
- [ ] T104 [US3] Create performance dashboard API (trend charts, product ranking, channel analysis) in `backend/internal/distribution/handler/performance_handler.go`
- [ ] T105 [US3] Create admin distributor list/detail API (filter by type/grade/status, freeze/unfreeze/cancel) in `backend/internal/distribution/handler/admin_distributor_handler.go`
- [ ] T106 [US3] Create admin commission settlement audit API (list/approve/reject/batch approve) in `backend/internal/distribution/handler/admin_withdrawal_handler.go`
- [ ] T107 [US3] Create admin distribution rule configuration API (commission rates, settlement rules, change log) in `backend/internal/distribution/handler/admin_rule_handler.go`
- [ ] T108 [US3] Create admin distribution report API (order stats, commission spend, activity analysis, export) in `backend/internal/distribution/handler/admin_report_handler.go`

**Checkpoint**: 分销体系后端全部完成，佣金计算和防薅羊毛规则可独立测试

---

## Phase 5: 支付扩展 - 银联+定金尾款 (Week 6-8) [US4]

**Goal**: 银联支付接入、定金+尾款模式、部分退款

**Independent Test**: 消费者可选择银联支付、定金+尾款模式完成订单

### Implementation

- [ ] T109 [P] [US4] Create UnionPay gateway adapter (smartwalle/unionpay, gateway + WAP payment) in `backend/internal/payment/gateway/unionpay.go`
- [ ] T110 [P] [US4] Create UnionPay callback handler (backUrl for confirmation, frontUrl for display only) in `backend/internal/payment/handler/unionpay_notify.go`
- [ ] T111 [P] [US4] Create UnionPay refund adapter (cancel for same-day, refund for next-day) in `backend/internal/payment/gateway/unionpay_refund.go`
- [ ] T112 [P] [US4] Create DepositOrder domain model (deposit_amount, balance_amount, balance_deadline) in `backend/internal/order/domain/deposit_order.go`
- [ ] T113 [US4] Create deposit payment flow service (create deposit payment, status transition to paid_deposit) in `backend/internal/payment/service/deposit_service.go`
- [ ] T114 [US4] Create balance payment flow service (create balance payment, reminder 3 days before via Asynq) in `backend/internal/payment/service/balance_service.go`
- [ ] T115 [US4] Create balance overdue handler (24h grace period, auto-cancel, inventory release, deposit refund) in `backend/internal/payment/service/overdue_service.go`
- [ ] T116 [US4] Create partial refund API (amount validation, original-channel return, cumulative check) in `backend/internal/payment/handler/partial_refund_handler.go`
- [ ] T117 [US4] Create payment status proactive query service (30s trigger, 60s retry, all 3 channels) in `backend/internal/payment/service/proactive_query_service.go`
- [ ] T118 [US4] Extend reconciliation system for UnionPay (file download, parsing, 3-way matching) in `backend/internal/payment/service/reconciliation_service.go`
- [ ] T119 [US4] Add UnionPay payment option and deposit/balance selection to web payment page in `frontend/web/components/payment/PaymentMethodSelector.vue`
- [ ] T120 [US4] Create balance payment reminder page (paid deposit, balance due, countdown) in `frontend/web/components/payment/BalanceReminder.vue`
- [ ] T121 [US4] Create partial refund application page in `frontend/web/components/order/PartialRefund.vue`
- [ ] T122 [US4] Create Douyin payment adapter (tt.pay API, conditional compilation) in `frontend/miniprogram/utils/payment.js`

**Checkpoint**: 银联支付和定金+尾款模式可独立测试

---

## Phase 6: 营销系统 (Week 7-9) [US5]

**Goal**: 优惠券和促销活动完整生命周期

**Independent Test**: 运营可创建优惠券/促销活动，消费者可领取优惠券并下单使用

### Implementation

- [ ] T123 [P] [US5] Create Coupon domain model (4 types: full_reduction/discount/cash/exchange) in `backend/internal/marketing/domain/coupon.go`
- [ ] T124 [P] [US5] Create CouponClaim domain model with status machine in `backend/internal/marketing/domain/coupon_claim.go`
- [ ] T125 [P] [US5] Create PromotionActivity domain model (flash_sale/full_reduction/early_bird) in `backend/internal/marketing/domain/promotion_activity.go`
- [ ] T126 [US5] Create coupon CRUD API (create with all config params, list, detail) in `backend/internal/marketing/handler/coupon_handler.go`
- [ ] T127 [US5] Create coupon distribution API (6 methods: push/center/product/activity/share/exchange) in `backend/internal/marketing/handler/coupon_distribution_handler.go`
- [ ] T128 [US5] Create coupon usage API (validate on order, occupy on confirm, use on pay, return on refund) in `backend/internal/marketing/handler/coupon_usage_handler.go`
- [ ] T129 [US5] Create promotion activity engine (flash sale with isolated stock, tiered reduction, early bird auto-match) in `backend/internal/marketing/service/activity_engine.go`
- [ ] T130 [US5] Create coupon analytics API (distributed/claimed/used/rate/GMV) in `backend/internal/marketing/handler/coupon_analytics_handler.go`
- [ ] T131 [P] [US5] Create coupon center page (claimable list, one-click claim) in `frontend/web/pages/coupon/index.vue`
- [ ] T132 [P] [US5] Create my coupons page (available/used/expired tabs) in `frontend/web/pages/coupon/mine.vue`
- [ ] T133 [US5] Create coupon selector component for order confirmation (sorted by discount, real-time calc) in `frontend/web/components/coupon/CouponSelector.vue`
- [ ] T134 [P] [US5] Create mini-program coupon center page in `frontend/miniprogram/pages/coupon/index.vue`
- [ ] T135 [P] [US5] Create mini-program my coupons page in `frontend/miniprogram/pages/coupon/mine.vue`
- [ ] T136 [US5] Create admin coupon management page (create/list/analytics) in `frontend/admin/views/marketing/CouponManage.vue`
- [ ] T137 [US5] Create admin coupon analytics page in `frontend/admin/views/marketing/CouponAnalytics.vue`
- [ ] T138 [US5] Create admin promotion activity management page in `frontend/admin/views/marketing/PromotionActivity.vue`

**Checkpoint**: 优惠券和促销活动可独立测试

---

## Phase 7: Meilisearch 搜索集成 (Week 7-8)

**Goal**: 产品搜索从数据库全文搜索迁移至 Meilisearch

**Independent Test**: 产品搜索响应 <50ms，支持多维度过滤和 typo 容错

### Implementation

- [ ] T139 [P] Design Meilisearch product index schema (searchable/filterable/sortable attributes for domestic+outbound+cruise) in `backend/internal/shared/meili/index_schema.go`
- [ ] T140 [P] Create Meilisearch index initialization script in `backend/scripts/meili_init.go`
- [ ] T141 Create product sync service (Asynq task, real-time push on product CRUD, <5s delay) in `backend/internal/product/service/search_sync_service.go`
- [ ] T142 Create outbound product search handler with Meilisearch (continent/country/visa_type/city/days filters) in `backend/internal/product/handler/search_handler.go`
- [ ] T143 Create search suggestion API (hot destinations → product names → attractions) in `backend/internal/product/handler/search_suggest_handler.go`
- [ ] T144 Create database fallback search (Meilisearch unavailable → PostgreSQL tsvector) in `backend/internal/product/repository/search_fallback_repo.go`

**Checkpoint**: 搜索功能迁移至 Meilisearch，响应时间 <50ms

---

## Phase 8: 抖音小程序 (Week 8-9)

**Goal**: 抖音小程序核心页面，与微信小程序共享代码基

**Independent Test**: 用户可在抖音小程序完成登录→浏览→下单→支付→查看订单全流程

### Implementation

- [ ] T145 [P] Configure Uni-App conditional compilation for Douyin (`#ifdef MP-TOUTIAO`) in `frontend/miniprogram/utils/platform.js`
- [ ] T146 [P] Create Douyin login adapter (tt.login, OpenID acquisition, account binding) in `frontend/miniprogram/pages-douyin/login.vue`
- [ ] T147 [P] Create Douyin payment adapter (tt.pay API integration) in `frontend/miniprogram/utils/douyin_payment.js`
- [ ] T148 Create core page conditional compilation adaptations (product list/detail/booking/order/personal) in `frontend/miniprogram/utils/douyin_adapters.js`
- [ ] T149 Create Douyin mini-program manifest and config in `frontend/miniprogram/manifest.json` (MP-TOUTIAO section)
- [ ] T150 Submit Douyin mini-program for review (7-14 day cycle)

**Checkpoint**: 抖音小程序可提审

---

## Phase 9: 供应商工作台前端 (Week 6-8)

**Goal**: 供应商工作台完整前端页面

**Independent Test**: 供应商可登录工作台管理产品、处理订单、查看结算

### Implementation

- [ ] T151 [P] Create supplier workspace login page with phone+code auth in `frontend/admin/views/supplier-workspace/Login.vue`
- [ ] T152 [P] Create supplier workspace layout with independent menu system in `frontend/admin/views/supplier-workspace/Layout.vue`
- [ ] T153 [P] Create supplier workspace router config with supplier_id isolation in `frontend/admin/router/supplier.ts`
- [ ] T154 [P] Create supplier product management page (list/filter/create multi-step form) in `frontend/admin/views/supplier-workspace/ProductManage.vue`
- [ ] T155 [P] Create supplier product publish form (basic info→itinerary→price→refund rules→inventory) in `frontend/admin/views/supplier-workspace/ProductPublish.vue`
- [ ] T156 [P] Create supplier order handling page (list/confirm/reject/detail/refund) in `frontend/admin/views/supplier-workspace/OrderHandle.vue`
- [ ] T157 [P] Create supplier settlement page (list/detail/confirm/dispute) in `frontend/admin/views/supplier-workspace/Settlement.vue`
- [ ] T158 [P] Create supplier withdrawal page (apply/history) in `frontend/admin/views/supplier-workspace/Withdraw.vue`
- [ ] T159 [P] Create supplier income/expense detail page in `frontend/admin/views/supplier-workspace/IncomeDetail.vue`
- [ ] T160 [P] Create supplier statistics page (sales/revenue/trends/reviews) in `frontend/admin/views/supplier-workspace/Statistics.vue`

**Checkpoint**: 供应商工作台前端全部完成

---

## Phase 10: 分销商中心前端 (Week 7-9)

**Goal**: 分销商中心 Web + 小程序完整前端

**Independent Test**: 分销商可登录中心管理推广、查看佣金、申请提现

### Implementation

- [ ] T161 [P] Create distributor application page (personal/enterprise type selection) in `frontend/web/pages/distributor/apply.vue`
- [ ] T162 [P] Create distributor center login page in `frontend/web/pages/distributor/login.vue`
- [ ] T163 [P] Create distributor center layout with sidebar navigation in `frontend/web/pages/distributor/Layout.vue`
- [ ] T164 [P] Create distributor overview dashboard (total/withdrawable/frozen, quick actions, announcements) in `frontend/web/pages/distributor/index.vue`
- [ ] T165 [P] Create my-promotion page (product list, link/QR management, click/order stats) in `frontend/web/pages/distributor/promote.vue`
- [ ] T166 [P] Create my-team page (member list, team summary, invite link/code, leaderboard) in `frontend/web/pages/distributor/team.vue`
- [ ] T167 [P] Create commission detail page (list, filter, export, L1/L2 tags) in `frontend/web/pages/distributor/commission.vue`
- [ ] T168 [P] Create commission withdrawal page (balance, apply, history) in `frontend/web/pages/distributor/withdraw.vue`
- [ ] T169 [P] Create performance dashboard page (trends, product ranking, channel analysis) in `frontend/web/pages/distributor/performance.vue`
- [ ] T170 [P] Create mini-program distributor center pages (overview/promote/team/commission/withdraw) in `frontend/miniprogram/pages/distributor/`

**Checkpoint**: 分销商中心前端全部完成

---

## Phase 11: 集成测试与安全加固 (Week 9-10)

**Goal**: 端到端测试、安全加固、性能压测

### Implementation

- [ ] T171 [P] Create outbound travel end-to-end test (browse→book→pay→visa→track) in `backend/tests/e2e/outbound_test.go`
- [ ] T172 [P] Create supplier end-to-end test (apply→audit→product→order→settlement) in `backend/tests/e2e/supplier_test.go`
- [ ] T173 [P] Create distribution end-to-end test (apply→promote→order→commission→withdraw) in `backend/tests/e2e/distribution_test.go`
- [ ] T174 [P] Create payment end-to-end test (alipay/wechat/unionpay, full/deposit+balance, partial refund) in `backend/tests/e2e/payment_test.go`
- [ ] T175 [P] Create commission calculation unit tests (base/ratio/attribution/cap/50% cap) in `backend/tests/unit/commission_test.go`
- [ ] T176 [P] Create anti-fraud unit tests (self-purchase/identity/device/IP) in `backend/tests/unit/anti_fraud_test.go`
- [ ] T177 [P] Create visa state machine unit tests (all transitions, invalid transitions) in `backend/tests/unit/visa_state_test.go`
- [ ] T178 Configure MFA (TOTP) for new admin roles (supplier auditor, distribution manager, finance) in `backend/internal/shared/middleware/mfa.go`
- [ ] T178a [P] Add MFA enforcement to supplier workspace audit operations (product approval, settlement approval) in `backend/internal/supplier/handler/audit_handler.go`
- [ ] T178b [P] Add MFA enforcement to distributor withdrawal approval flow in `backend/internal/distribution/handler/admin_withdrawal_handler.go`
- [ ] T178c [P] Extend audit logging to all new supplier/distribution/visa domain operations in `backend/internal/shared/middleware/audit_extend.go`
- [ ] T179 Deploy WAF (reverse proxy rules) and HIDS configuration in `infra/security/waf_rules.yml`
- [ ] T180 Configure cross-region backup (daily full + hourly incremental, RPO <1min) in `infra/backup/backup_strategy.yml`
- [ ] T181 Run MLPS Level 3 compliance checklist validation (32 control points) in `backend/tests/security/mlps_checklist_test.go`
- [ ] T182 Run performance stress test (QPS ≥10000, order TPS ≥500, P99 targets) in `backend/tests/performance/load_test.go`
- [ ] T183 Verify supplier data isolation (cross-supplier data access attempt should fail) in `backend/tests/security/data_isolation_test.go`
- [ ] T184 Verify RLS policies for all supplier/distributor tables in `backend/tests/security/rls_test.go`

**Checkpoint**: 所有测试通过，安全合规达标

---

## Phase 12: 部署与监控 (Week 10)

**Goal**: 生产部署、监控告警配置、灰度上线

### Implementation

- [ ] T185 [P] Register all 5 services as Windows services via WinSW in `infra/winsw/`
- [ ] T186 [P] Configure Prometheus scrape targets for all services in `infra/prometheus/prometheus.yml`
- [ ] T187 [P] Configure Grafana dashboards (system/app/business layers) in `infra/grafana/dashboards/`
- [ ] T188 [P] Configure Jaeger for distributed tracing (1% sampling in prod) in `infra/jaeger/jaeger.yml`
- [ ] T189 [P] Configure log collection (Zap + lumberjack → Loki) in `backend/internal/shared/logging/config.go`
- [ ] T190 Configure alert rules (P1: error rate >1%, QPS drop >50%, DB connection >80%; P2: CPU >80%, refund rate >20%) in `infra/prometheus/alerts.yml`
- [ ] T191 Create business metrics Grafana dashboard (order volume, payment success rate, commission spend, visa completion rate) in `infra/grafana/dashboards/business.yml`
- [ ] T192 Perform rolling deployment with zero-downtime (health check /ready before traffic routing)
- [ ] T193 Execute gray-scale rollout (supplier → outbound → distribution → marketing)

**Checkpoint**: 一期全量上线，监控告警就绪

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (基础设施)**: No dependencies — start immediately
- **Phase 2 (出境游 US1)**: Depends on Phase 1 completion
- **Phase 3 (供应商 US2)**: Depends on Phase 1 completion; can run parallel with Phase 2
- **Phase 4 (分销 US3)**: Depends on Phase 1 completion; can run parallel with Phase 2/3
- **Phase 5 (支付 US4)**: Depends on Phase 1; can run parallel with Phase 2/3/4
- **Phase 6 (营销 US5)**: Depends on Phase 1; can run parallel with Phase 2/3/4/5
- **Phase 7 (Meilisearch)**: Depends on Phase 1; can run parallel with Phase 2-6
- **Phase 8 (抖音小程序)**: Depends on Phase 2 (outbound pages exist); can start after T048-T052
- **Phase 9 (供应商工作台前端)**: Depends on Phase 3 backend APIs
- **Phase 10 (分销商中心前端)**: Depends on Phase 4 backend APIs
- **Phase 11 (集成测试)**: Depends on Phase 2-6 completion
- **Phase 12 (部署监控)**: Depends on Phase 11 completion

### Parallel Opportunities

```bash
# After Phase 1 completes, these can run in parallel:
Phase 2 (US1 出境游)    ← backend team A
Phase 3 (US2 供应商)    ← backend team B
Phase 4 (US3 分销)      ← backend team C
Phase 5 (US4 支付)      ← backend team A (after T021-T054)
Phase 6 (US5 营销)      ← backend team B (after T055-T076)
Phase 7 (Meilisearch)   ← any backend dev

# After backend APIs ready:
Phase 9 (供应商前端)     ← frontend team (after Phase 3)
Phase 10 (分销商前端)    ← frontend team (after Phase 4)
```

---

## Implementation Strategy

### MVP First (Week 1-5)

1. Complete Phase 1 (基础设施) → 5 services running
2. Complete Phase 2 (出境游 US1) → 出境游全流程可用
3. **STOP and VALIDATE**: 出境游产品→预订→签证独立测试
4. 供应商可开始入驻（Phase 3 进行中）

### Incremental Delivery

1. Week 1-2: 基础设施就绪
2. Week 3-5: 出境游上线 + 供应商入驻开放
3. Week 5-7: 分销体系上线
4. Week 6-8: 银联支付 + 营销系统上线
5. Week 8-9: 抖音小程序提审
6. Week 9-10: 全量灰度上线

---

## Notes

- Total tasks: 193
- Phase 1: 20 tasks (infrastructure)
- Phase 2 (US1): 34 tasks (outbound + visa)
- Phase 3 (US2): 22 tasks (supplier)
- Phase 4 (US3): 32 tasks (distribution)
- Phase 5 (US4): 14 tasks (payment)
- Phase 6 (US5): 16 tasks (marketing)
- Phase 7: 6 tasks (Meilisearch)
- Phase 8: 6 tasks (Douyin)
- Phase 9: 10 tasks (supplier frontend)
- Phase 10: 10 tasks (distributor frontend)
- Phase 11: 14 tasks (testing + security)
- Phase 12: 9 tasks (deployment + monitoring)
