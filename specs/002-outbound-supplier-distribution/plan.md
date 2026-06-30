# Implementation Plan: 一期扩展 — 出境游 + 供应商开放平台 + 分销体系

**Branch**: `002-outbound-supplier-distribution` | **Date**: 2026-06-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/002-outbound-supplier-distribution/spec.md`

## Summary

一期在 MVP（境内跟团游）基础上扩展三大核心能力：出境游业务线（含签证服务交易闭环）、供应商开放平台（入驻/审核/结算）、二级分销体系。同时扩展银联支付、定金+尾款支付模式、营销系统（优惠券+促销）、财务管理增强、抖音小程序。架构上从 Gin 单体拆分为 5 个独立微服务，引入 NATS 消息队列、Meilisearch 搜索引擎、Consul 服务发现。新增 89 条功能需求，覆盖 8 个用户故事，预计 10 周交付。

## Technical Context

**Language/Version**: Go 1.26+（后端）、TypeScript/Vue 3（前端）

**Primary Dependencies**: Gin（Web框架）、GORM v2 + pgx（ORM/驱动）、NATS 2.11+（消息队列）、Meilisearch 1.19+（搜索引擎）、Consul 1.22+（服务发现）、Traefik 3.x+（API网关）、Asynq（任务调度）、smartwalle/alipay v3 + wechatpay-go + smartwalle/unionpay（支付SDK）、Uni-App Vue 3（小程序）、Nuxt.js 3（Web SSR）、Vue 3 + Element Plus（后台/工作台）

**Storage**: PostgreSQL 18+（主数据库）、Redis/Memurai 7.2+（缓存/会话）、Meilisearch（搜索索引）、阿里云 OSS（文件存储）

**Testing**: Go 标准 testing + testify、Vue Test Utils（前端）、golangci-lint（代码质量）

**Target Platform**: Windows Server 2022+（服务端部署，WinSW 注册服务）

**Project Type**: Web application（多端：Web SSR + 小程序 + 后台管理 + 供应商工作台 + 分销商中心）

**Performance Goals**: 全站 QPS ≥ 10,000、订单 TPS ≥ 500、首页 P95 ≤ 1.5s、产品列表 P99 ≤ 1.0s、订单确认 P99 ≤ 500ms

**Constraints**: 等保三级合规（TLS 1.3、AES-256-GCM 字段加密、JWT RS256、RBAC + MFA、审计日志 ≥ 6个月）、Windows Server 原生部署（CGO_ENABLED=0 静态编译）、所有服务通过 WinSW 注册为 Windows 服务

**Scale/Scope**: 日均 1,000 单、并发 10,000 用户、50-500 家供应商、5 个微服务 + 4 个前端项目

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原则 | 要求 | 一期合规状态 | 备注 |
|------|------|------------|------|
| I. API-First Design | 所有能力通过 OpenAPI 3.0 契约暴露 | ✅ 通过 | 新增 API 先写契约再实现，沿用 swaggo/swag 自动生成 |
| II. DDD 服务边界 | 按业务域划分服务，跨域通过接口通信 | ✅ 通过 | 新增 distribution-service 独立域，订单/支付服务拆分为独立部署单元 |
| III. Security-by-Design | TLS 1.3、AES-256-GCM、JWT RS256、RBAC、MFA、审计 ≥ 6个月 | ✅ 通过 | 新增供应商/分销商敏感信息（身份证/护照/银行卡）加密存储，沿用 MVP 安全体系 |
| IV. Progressive Delivery | MVP → Phase 2 → Phase 3 递进交付 | ✅ 通过 | 一期即 Phase 2，在 MVP 稳定基础上扩展，不破坏已有功能 |
| V. Code Quality | golangci-lint 零错误、ESLint 零警告、Conventional Commits、核心逻辑测试 ≥ 70% | ✅ 通过 | 新增佣金计算、签证状态机、防薅羊毛规则需 ≥ 70% 覆盖率 |

**技术栈合规检查**：

| 组件 | 宪法要求 | 一期使用 | 合规 |
|------|---------|---------|------|
| 后端语言 | Go 1.26+ | Go 1.26+ | ✅ |
| Web 框架 | Gin | Gin | ✅ |
| ORM | GORM v2 | GORM v2 | ✅ |
| 数据库驱动 | pgx v5 | pgx v5 | ✅ |
| 主数据库 | PostgreSQL 17.x (dev) | PostgreSQL 18+ | ✅ |
| 缓存 | Redis/Memurai 7.2+ | Memurai 7.2+ | ✅ |
| 搜索引擎 | Meilisearch 1.19+ | Meilisearch 1.19+ | ✅ |
| 消息队列 | NATS 2.11+ | NATS 2.11+ | ✅ |
| 任务队列 | Asynq | Asynq | ✅ |
| API 网关 | Traefik 3.x+ | Traefik 3.x+ | ✅ |
| 服务发现 | Consul 1.22+ | Consul 1.22+ | ✅ |
| 前端 Web | Nuxt.js 3 | Nuxt.js 3 | ✅ |
| 前端 Admin | Vue 3 + Element Plus | Vue 3 + Element Plus | ✅ |
| 前端小程序 | Uni-App (Vue 3) | Uni-App (Vue 3) | ✅ |
| 部署 | Windows Server + WinSW | Windows Server + WinSW | ✅ |

**结论：全部通过，无违规项。**

## Project Structure

### Documentation (this feature)

```text
specs/002-outbound-supplier-distribution/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
│   ├── supplier-api.yaml
│   ├── distribution-api.yaml
│   ├── visa-api.yaml
│   ├── payment-extension-api.yaml
│   └── marketing-api.yaml
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
# 后端微服务（从单体拆分）
backend/
├── cmd/                          # 各服务入口
│   ├── user-service/main.go
│   ├── product-service/main.go
│   ├── order-service/main.go
│   ├── payment-service/main.go
│   └── distribution-service/main.go  # 新增
├── internal/                     # 各服务内部实现
│   ├── user/
│   │   ├── domain/               # 领域模型
│   │   ├── service/              # 业务逻辑
│   │   ├── repository/           # 数据访问
│   │   └── handler/              # HTTP 处理器
│   ├── product/
│   │   ├── domain/
│   │   │   ├── outbound.go       # 出境游产品模型（新增）
│   │   │   └── visa_info.go      # 签证信息模型（新增）
│   │   ├── service/
│   │   ├── repository/
│   │   └── handler/
│   ├── order/
│   │   ├── domain/
│   │   │   ├── visa_order.go     # 签证订单模型（新增）
│   │   │   └── deposit_order.go  # 定金+尾款订单模型（新增）
│   │   ├── service/
│   │   ├── repository/
│   │   └── handler/
│   ├── payment/
│   │   ├── domain/
│   │   │   └── unionpay.go       # 银联支付模型（新增）
│   │   ├── service/
│   │   │   ├── unionpay_service.go   # 银联支付服务（新增）
│   │   │   └── deposit_service.go    # 定金+尾款服务（新增）
│   │   ├── gateway/              # 支付渠道网关（适配器模式）
│   │   │   ├── alipay.go
│   │   │   ├── wechat.go
│   │   │   └── unionpay.go       # 新增
│   │   ├── repository/
│   │   └── handler/
│   ├── supplier/                 # 供应商域（新增）
│   │   ├── domain/
│   │   │   ├── supplier.go
│   │   │   ├── settlement.go
│   │   │   └── commission_rule.go
│   │   ├── service/
│   │   ├── repository/
│   │   └── handler/
│   ├── distribution/             # 分销域（新增，独立服务）
│   │   ├── domain/
│   │   │   ├── distributor.go
│   │   │   ├── commission.go
│   │   │   ├── promotion_link.go
│   │   │   └── anti_fraud.go     # 防薅羊毛规则
│   │   ├── service/
│   │   │   ├── distributor_service.go
│   │   │   ├── commission_service.go
│   │   │   ├── promotion_service.go
│   │   │   └── anti_fraud_service.go
│   │   ├── repository/
│   │   └── handler/
│   ├── visa/                     # 签证域（新增，归属 order-service）
│   │   ├── domain/
│   │   │   ├── visa_order.go
│   │   │   ├── visa_material.go
│   │   │   └── visa_progress.go
│   │   ├── service/
│   │   ├── repository/
│   │   └── handler/
│   ├── marketing/                # 营销域（新增或扩展）
│   │   ├── domain/
│   │   │   ├── coupon.go
│   │   │   └── promotion_activity.go
│   │   ├── service/
│   │   ├── repository/
│   │   └── handler/
│   └── shared/                   # 共享内核
│       ├── middleware/            # 中间件（鉴权/限流/审计）
│       ├── event/                 # NATS 事件定义
│       ├── encryption/            # AES-256-GCM 加密
│       └── errors/                # 统一错误码
├── migrations/                   # 数据库迁移
│   ├── 002_outbound_tables.sql
│   ├── 003_supplier_tables.sql
│   ├── 004_distribution_tables.sql
│   ├── 005_visa_tables.sql
│   ├── 006_marketing_tables.sql
│   └── 007_payment_extension.sql
├── api/                          # OpenAPI 契约
│   └── openapi/
│       ├── v2/
│       │   ├── supplier.yaml
│       │   ├── distribution.yaml
│       │   ├── visa.yaml
│       │   └── marketing.yaml
│       └── v2.yaml               # 聚合文档
└── configs/                      # 配置文件

# 前端项目
frontend/
├── web/                          # Web 销售平台（Nuxt.js 3）
│   ├── pages/
│   │   ├── outbound/             # 出境游页面（新增）
│   │   │   ├── index.vue         # 出境游产品列表
│   │   │   ├── [id].vue          # 出境游产品详情
│   │   │   └── booking.vue       # 出境游预订（五步向导）
│   │   ├── coupon/               # 优惠券页面（新增）
│   │   │   ├── index.vue         # 领券中心
│   │   │   └── mine.vue          # 我的优惠券
│   │   └── distributor/          # 分销商中心（新增）
│   │       ├── index.vue         # 数据概览
│   │       ├── promote.vue       # 我的推广
│   │       ├── team.vue          # 我的团队
│   │       ├── commission.vue    # 佣金明细
│   │       ├── withdraw.vue      # 佣金提现
│   │       └── apply.vue         # 入驻申请
│   └── components/
│       ├── outbound/             # 出境游组件（新增）
│       ├── visa/                 # 签证相关组件（新增）
│       ├── coupon/               # 优惠券组件（新增）
│       └── distributor/          # 分销商组件（新增）
├── admin/                        # 后台管理系统（Vue 3 + Element Plus）
│   ├── views/
│   │   ├── supplier/             # 供应商管理（新增）
│   │   │   ├── AuditList.vue     # 入驻审核
│   │   │   ├── SupplierList.vue  # 供应商列表
│   │   │   └── SupplierDetail.vue
│   │   ├── supplier-workspace/   # 供应商工作台（新增）
│   │   │   ├── ProductManage.vue
│   │   │   ├── OrderHandle.vue
│   │   │   ├── Settlement.vue
│   │   │   ├── Withdraw.vue
│   │   │   └── Statistics.vue
│   │   ├── distribution/         # 分销管理（新增）
│   │   │   ├── DistributorList.vue
│   │   │   ├── DistributorAudit.vue
│   │   │   ├── CommissionAudit.vue
│   │   │   ├── CommissionRules.vue
│   │   │   └── DistributionReport.vue
│   │   ├── finance/              # 财务管理（新增/扩展）
│   │   │   ├── PaymentFlow.vue
│   │   │   ├── RefundManage.vue
│   │   │   ├── SupplierSettlement.vue
│   │   │   ├── InvoiceManage.vue
│   │   │   └── FinanceReport.vue
│   │   ├── marketing/            # 营销管理（新增）
│   │   │   ├── CouponManage.vue
│   │   │   ├── CouponAnalytics.vue
│   │   │   └── PromotionActivity.vue
│   │   └── visa/                 # 签证管理（新增）
│   │       ├── VisaOrderList.vue
│   │       └── VisaMaterialAudit.vue
│   └── router/
│       └── modules/              # 路由模块（按权限动态加载）
├── miniprogram/                  # 小程序（Uni-App Vue 3）
│   ├── pages/
│   │   ├── outbound/             # 出境游页面（新增）
│   │   │   ├── list.vue
│   │   │   ├── detail.vue
│   │   │   └── booking.vue
│   │   ├── visa/                 # 签证页面（新增）
│   │   │   ├── progress.vue      # 签证进度查询
│   │   │   └── materials.vue     # 材料提交
│   │   ├── coupon/               # 优惠券（新增）
│   │   ├── distributor/          # 分销商中心（新增）
│   │   └── order/
│   │       └── detail.vue        # 订单详情（扩展签证/尾款）
│   ├── pages-douyin/             # 抖音小程序特有页面（条件编译）
│   │   └── login.vue             # 抖音登录
│   └── utils/
│       ├── platform.js           # 平台适配层（条件编译）
│       └── payment.js            # 支付适配（微信/抖音）

# 基础设施
infra/
├── consul/                       # Consul 配置
├── traefik/                      # Traefik 路由配置
├── nats/                         # NATS 配置
├── meilisearch/                  # Meilisearch 索引配置
└── winsw/                        # WinSW 服务注册配置
```

**Structure Decision**: 采用微服务架构，后端按 DDD 域划分为 5 个独立服务（user/product/order/payment/distribution），前端按端分为 Web(SSR)、Admin(SPA)、MiniProgram(Uni-App) 三个项目。供应商域归属 product-service 或作为独立模块，分销域作为独立服务 distribution-service。签证域归属 order-service。

## Implementation Phases

### Phase 1: 基础设施与服务拆分（Week 1-2）

**目标**: 从单体拆分为微服务架构，搭建基础设施

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-001 | Consul 服务发现搭建 | 部署 Consul、封装服务注册/注销/健康检查 | 无 | Consul 集群配置、Go 客户端封装 |
| T-002 | NATS 消息队列搭建 | 部署 NATS、配置 JetStream 持久化流、定义主题 | 无 | NATS 配置、事件 DTO 定义、发布/订阅封装 |
| T-003 | Meilisearch 搭建与索引设计 | 部署 Meilisearch、设计产品索引 schema、同步任务 | 无 | Meilisearch 部署、索引 schema、Asynq 同步任务 |
| T-004 | Traefik 路由配置 | 配置动态路由、限流、SSL 终结 | T-001 | Traefik 配置、服务路由规则 |
| T-005 | 用户服务拆分 | 从单体拆出 user-service、WinSW 注册 | T-001 | user-service 独立部署 |
| T-006 | 产品服务拆分 | 从单体拆出 product-service、含出境游扩展 | T-001 | product-service 独立部署 |
| T-007 | 订单服务拆分 | 从单体拆出 order-service、NATS 事件发布 | T-001, T-002 | order-service 独立部署 |
| T-008 | 支付服务拆分 | 从单体拆出 payment-service、渠道网关抽象 | T-001, T-002 | payment-service 独立部署 |
| T-009 | distribution-service 骨架 | 创建分销服务空骨架、健康检查端点 | T-001, T-002 | distribution-service 骨架 |
| T-010 | 数据库迁移脚本 | 新增供应商/分销/签证/营销/支付扩展表 | 无 | 6 个 SQL 迁移脚本 |
| T-011 | 统一事件总线定义 | NATS 主题定义、事件 DTO、发布/订阅封装 | T-002 | shared/event 包 |
| T-012 | API 网关聚合文档 | OpenAPI v2 聚合文档、Swagger UI | T-004 | openapi/v2.yaml |

### Phase 2: 出境游业务线（Week 3-5）

**目标**: 实现出境游产品浏览、预订、签证服务闭环（15 个功能点）

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-013 | 出境游产品数据模型 | 出境游产品表扩展、签证信息表、国家/地区层级树 | T-006, T-010 | 数据模型 + 迁移 |
| T-014 | 出境游产品列表 API | 筛选（大洲/国家/签证类型/口岸）、Meilisearch 索引 | T-013, T-003 | GET /api/v2/products/outbound |
| T-015 | 出境游产品详情 API | 签证信息卡片、行前信息服务数据 | T-013 | GET /api/v2/products/outbound/:id |
| T-016 | 护照信息管理 | 护照 CRUD、OCR 识别封装、有效期校验 | T-005 | 护照 API + OCR 适配器 |
| T-017 | 出境游预订流程 API | 五步向导（含签证代办选择）、护照有效期拦截 | T-007, T-016 | POST /api/v2/orders/outbound |
| T-018 | 签证订单模型 | 签证订单/材料/进度表、五节点状态机 | T-010, T-007 | 数据模型 + 状态机 |
| T-019 | 签证材料管理 API | 材料上传（≤10MB）、清单生成（按职业）、预审 | T-018 | 签证材料 CRUD API |
| T-020 | 签证进度跟踪 API | 进度查询、状态变更 NATS 事件发布 | T-018, T-011 | 签证进度 API + 事件 |
| T-021 | 签证通知服务 | 短信+站内信通知（状态变更/出签/拒签/物流） | T-020, T-011 | NATS 消费者 + 通知模板 |
| T-022 | 签证历史与提醒 | 历史记录查询、有效期到期提醒（Asynq） | T-018 | API + 定时任务 |
| T-023 | 行前信息服务 API | 入境政策/材料/入境卡/海关/航班/天气/紧急联系 | T-015 | 行前服务 API |
| T-024 | Web 出境游产品列表页 | 出境游筛选栏、签证标签、产品卡片 | T-014 | pages/outbound/index.vue |
| T-025 | Web 出境游产品详情页 | 签证信息卡片、行程（含航班）、FAQ | T-015 | pages/outbound/[id].vue |
| T-026 | Web 出境游预订页 | 五步向导（护照/OCR/签证/附加/确认） | T-017 | pages/outbound/booking.vue |
| T-027 | Web 签证进度页 | 五节点进度条、材料上传、物流跟踪 | T-020 | 签证进度组件 |
| T-028 | Web 行前服务页 | 入境政策/入境卡模板/航班/清单 | T-023 | 行前服务页面 |
| T-029 | 小程序出境游页面 | 列表/详情/预订/签证进度（条件编译） | T-014~T-023 | miniprogram/outbound/* |
| T-030 | 后台签证管理页 | 签证订单列表、材料审核、进度更新 | T-018~T-022 | admin/views/visa/* |

### Phase 3: 供应商开放平台（Week 4-6）

**目标**: 实现供应商入驻、工作台、结算全流程

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-031 | 供应商数据模型 | 供应商/资质/结算单/佣金规则表 | T-010 | 数据模型 + 迁移 |
| T-032 | 供应商入驻申请 API | 入驻申请提交、资料校验、申请编号生成 | T-031 | POST /api/v2/suppliers/apply |
| T-033 | 供应商审核流程 API | 二级审核、审核超时告警（Asynq） | T-032 | 审核 API + 定时任务 |
| T-034 | 电子合同签署 | 合同模板管理、PDF 生成、CA 签章对接 | T-033 | 合同服务 |
| T-035 | 供应商工作台产品管理 | 产品发布/编辑/上下架/团期/库存/审核追踪 | T-031, T-006 | 供应商产品 API |
| T-036 | 供应商工作台订单处理 | 订单列表（数据隔离）、确认/拒绝、退改 | T-031, T-007 | 供应商订单 API |
| T-037 | 供应商结算五步流程 | 自动生成→核对→确认→打款→归档 | T-031, T-011 | 结算 API + NATS 事件 |
| T-038 | 供应商提现管理 | 提现申请、审批、打款记录 | T-037 | 提现 API |
| T-039 | 供应商佣金规则配置 | 品类级/供应商级佣金、阶梯佣金、审批 | T-031 | 佣金规则 API |
| T-040 | 供应商数据统计 | 销量/销售额/评价、Asynq 预聚合 | T-035, T-036 | 统计 API + 定时任务 |
| T-041 | 供应商数据隔离中间件 | tenant_id + supplier_id 隔离、RLS | T-031 | 中间件 |
| T-042 | Web 供应商入驻申请页 | 多步骤表单、自动保存、进度查询 | T-032 | 入驻申请页面 |
| T-043 | 后台供应商审核页 | 审核列表、资料预览、操作按钮 | T-033 | 审核页面 |
| T-044 | 供应商工作台页面 | 产品/订单/结算/提现/统计全功能 | T-035~T-040 | 工作台全部页面 |
| T-045 | 后台供应商管理页 | 供应商列表/详情/状态/佣金配置 | T-031 | 管理页面 |

### Phase 4: 二级分销体系（Week 5-7）

**目标**: 实现分销商入驻、推广、佣金计算、提现全流程

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-046 | 分销商数据模型 | 分销商/关系/佣金/提现/推广链接表 | T-010 | 数据模型 + 迁移 |
| T-047 | 分销商入驻申请 API | 个人/企业入驻、身份证/营业执照/银行卡校验 | T-046 | POST /api/v2/distributors/apply |
| T-048 | 分销商审核与协议签署 | 审核流程、分销编码生成、协议签署记录 | T-047 | 审核 API |
| T-049 | 分销关系与邀请机制 | 邀请链接/邀请码、二级关系建立、团队管理 | T-046 | 邀请 API |
| T-050 | 推广链接与二维码生成 | 短链接、二维码（含Logo）、URL+Cookie 双轨跟踪 | T-046 | 推广链接 API |
| T-051 | 分销订单跟踪 | 订单记录分销来源、点击/转化统计 | T-046, T-007 | 跟踪 API |
| T-052 | 佣金规则配置 API | 三级优先级（产品>品类>全局）、一级/二级比例 | T-046 | 佣金规则 API |
| T-053 | 佣金计算引擎 | NATS 事件驱动异步计算、基数/比例/归属/上限规则 | T-051, T-052, T-011 | 佣金计算服务 |
| T-054 | 佣金冻结与解冻 | T+N 冻结期、自动解冻定时任务 | T-053 | 冻结/解冻逻辑 + Asynq |
| T-055 | 佣金退款追回 | 全额/部分退款退佣、冻结期内/外处理 | T-053, T-007 | 退佣逻辑 |
| T-056 | 分销商提现管理 | 提现申请（≥100元）、审核、高级分销商加速（T+3） | T-054 | 提现 API |
| T-057 | 防薅羊毛引擎 | 自购禁止/身份隔离/设备关联/IP频率/违规处罚 | T-046, T-051 | 反作弊服务 |
| T-058 | 分销商等级体系 | 普通/高级、自动升降级定时任务（90天复核） | T-046 | 等级服务 + Asynq |
| T-059 | 分销商中心首页 | 数据概览看板、快捷入口、公告 | T-053, T-054 | 首页 API + 页面 |
| T-060 | 分销商"我的推广" | 推广产品列表、链接/二维码管理、数据统计 | T-050, T-051 | 推广 API + 页面 |
| T-061 | 分销商"我的团队" | 团队成员、业绩汇总、邀请、排行榜 | T-049 | 团队 API + 页面 |
| T-062 | 分销商"佣金明细" | 佣金记录列表、筛选导出、标签区分 | T-053 | 佣金明细 API + 页面 |
| T-063 | 分销商"佣金提现" | 可提现余额、提现申请、历史记录 | T-056 | 提现 API + 页面 |
| T-064 | 分销商"业绩看板" | 趋势图表、数据汇总、产品排行、渠道分析 | T-051 | 业绩 API + 页面 |
| T-065 | 后台分销商管理页 | 列表/审核/等级/冻结/注销、佣金审核 | T-046~T-058 | 管理页面 |
| T-066 | 后台分销规则配置页 | 三级佣金配置、结算规则、变更日志 | T-052 | 配置页面 |
| T-067 | 后台分销数据报表 | 订单统计、佣金支出、活跃度、导出 | T-051, T-053 | 报表页面 |

### Phase 5: 支付扩展（Week 6-8）

**目标**: 银联支付接入、定金+尾款模式、部分退款

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-068 | 银联支付网关适配器 | smartwalle/unionpay 集成、网关/WAP 支付 | T-008 | 银联网关适配器 |
| T-069 | 银联回调处理 | 双重通知机制（backUrl 确认/frontUrl 展示） | T-068 | 回调处理器 |
| T-070 | 银联退款接口 | 消费撤销（当日）/退货（隔日） | T-068 | 退款适配器 |
| T-071 | 定金+尾款订单模型 | 定金/尾款支付记录、状态扩展 | T-010, T-007 | 数据模型 |
| T-072 | 定金支付流程 | 定金支付订单创建、状态流转 | T-071, T-008 | 定金支付逻辑 |
| T-073 | 尾款支付流程 | 尾款支付订单创建、提醒（Asynq 3天前） | T-071, T-007 | 尾款支付逻辑 + 定时任务 |
| T-074 | 尾款逾期处理 | 宽限期 24 小时、自动取消、库存释放 | T-073 | 逾期处理逻辑 |
| T-075 | 部分退款支持 | 部分退款 API、原路退回、累计校验 | T-008 | 部分退款逻辑 |
| T-076 | 支付状态主动查询 | 30 秒未回调触发查询、60 秒重试 | T-008 | 主动查询逻辑 |
| T-077 | 对账系统扩展 | 银联对账文件解析、三方轧账 | T-068 | 银联对账逻辑 |
| T-078 | Web 支付方式扩展 | 银联支付选项、定金+尾款选择 UI | T-068 | 支付页面扩展 |
| T-079 | 小程序支付适配 | 抖音支付 API 适配、条件编译 | T-068 | 支付适配层 |
| T-080 | 后台财务管理增强 | 支付流水/退款/结算/发票/报表 | T-068~T-077 | 财务管理全部页面 |

### Phase 6: 营销系统（Week 7-9）

**目标**: 优惠券和促销活动的完整生命周期

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-081 | 优惠券数据模型 | 优惠券/领取记录/使用记录表 | T-010 | 数据模型 |
| T-082 | 优惠券创建与发放 API | 四种类型、六种发放方式、库存/限领控制 | T-081 | 优惠券 CRUD + 发放 API |
| T-083 | 优惠券使用与核销 API | 下单选择、支付后核销、退款退回 | T-081, T-007 | 使用/核销逻辑 |
| T-084 | 促销活动数据模型 | 限时特惠/满减/早鸟活动表、活动库存表 | T-010 | 数据模型 |
| T-085 | 促销活动引擎 | 限时特惠、阶梯满减、早鸟折扣自动匹配 | T-084 | 活动引擎 |
| T-086 | 优惠券效果分析 | 发放量/领取量/核销量/核销率/GMV | T-081 | 分析 API |
| T-087 | Web 领券中心页 | 可领取优惠券列表、一键领取 | T-082 | 领券页面 |
| T-088 | Web 下单优惠券选择 | 可用券列表、实时优惠金额计算 | T-083 | 优惠券选择组件 |
| T-089 | 小程序优惠券页面 | 领券中心/我的优惠券/下单选择 | T-082, T-083 | 小程序页面 |
| T-090 | 后台营销管理页 | 优惠券管理/效果分析/促销活动配置 | T-082~T-086 | 营销管理页面 |

### Phase 7: 抖音小程序（Week 8-9）

**目标**: 抖音小程序核心页面，与微信小程序共享代码基

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-091 | 抖音小程序项目配置 | Uni-App 条件编译配置、抖音 API 适配层 | 无 | 项目配置 + 适配层 |
| T-092 | 抖音登录适配 | 抖音 OpenID 获取、账号绑定/创建 | T-091, T-005 | 登录适配 |
| T-093 | 抖音支付适配 | 抖音支付 SDK 集成、支付参数生成 | T-091, T-008 | 支付适配 |
| T-094 | 核心页面适配 | 产品列表/详情/预订/订单/个人中心条件编译 | T-091 | 核心页面 |
| T-095 | 抖音小程序提审 | 抖音平台审核提交（7-14天周期） | T-092~T-094 | 审核提交 |

### Phase 8: 集成测试与上线（Week 9-10）

**目标**: 端到端测试、性能测试、灰度上线

| ID | 任务 | 描述 | 依赖 | 产出 |
|----|------|------|------|------|
| T-096 | 端到端测试 | 出境游/供应商/分销/支付全流程测试 | Phase 2~7 | 测试报告 |
| T-097 | 性能测试 | QPS/TPS/响应时间/并发压测 | T-096 | 性能报告 + 优化 |
| T-098 | 安全测试 | 等保三级合规、渗透测试、加密验证 | T-096 | 安全报告 |
| T-099 | 灰度上线 | 分批放量（供应商→出境游→分销→营销） | T-097, T-098 | 灰度发布 |
| T-100 | 监控告警配置 | 业务指标监控、告警规则、Grafana 仪表盘 | T-099 | 监控配置 |

## Dependency Graph

```text
Phase 1 (基础设施) ─────────────────────────────────────────────────────┐
  T-001 Consul ──┬── T-004 Traefik ── T-012 API文档                    │
  T-002 NATS ────┼── T-007 订单服务 ──┬── T-017 出境游预订              │
  T-003 Meili ───┼── T-008 支付服务 ──┼── T-068 银联支付                │
  T-010 迁移 ────┼── T-009 分销骨架   │                                 │
  T-011 事件总线 ─┘                   │                                 │
                                      │                                 │
Phase 2 (出境游) ─────────────────────┤                                 │
  T-013~T-015 产品模型/API ── T-017 预订 ── T-018~T-022 签证闭环        │
  T-016 护照 ──────────────── T-017                                      │
  T-023 行前服务                                                        │
  T-024~T-030 前端页面                                                  │
                                                                      │
Phase 3 (供应商) ───────────────────────────────────────────────────────┤
  T-031 数据模型 ── T-032~T-034 入驻/审核/合同                          │
  T-035~T-040 工作台功能                                                │
  T-042~T-045 前端页面                                                  │
                                                                      │
Phase 4 (分销) ─────────────────────────────────────────────────────────┤
  T-046 数据模型 ── T-047~T-049 入驻/审核/邀请                          │
  T-050~T-051 推广/跟踪 ── T-053 佣金计算                              │
  T-054~T-056 冻结/退佣/提现                                            │
  T-057 防薅羊毛 ── T-058 等级体系                                      │
  T-059~T-067 前端页面                                                  │
                                                                      │
Phase 5 (支付扩展) ─────────────────────────────────────────────────────┤
  T-068~T-070 银联支付                                                  │
  T-071~T-074 定金+尾款                                                 │
  T-075~T-077 部分退款/主动查询/对账                                    │
  T-078~T-080 前端页面                                                  │
                                                                      │
Phase 6 (营销) ─────────────────────────────────────────────────────────┤
  T-081~T-086 优惠券/促销引擎                                           │
  T-087~T-090 前端页面                                                  │
                                                                      │
Phase 7 (抖音小程序) ───────────────────────────────────────────────────┤
  T-091~T-095 适配/提审                                                 │
                                                                      │
Phase 8 (集成测试) ─────────────────────────────────────────────────────┘
  T-096~T-100 测试/上线
```

## Complexity Tracking

> 无宪法违规，无需复杂度追踪。
