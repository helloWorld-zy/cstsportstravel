# Research: 一期扩展 — 出境游 + 供应商开放平台 + 分销体系

**Date**: 2026-06-30
**Feature**: specs/002-outbound-supplier-distribution/spec.md

## Research Tasks

### R-001: 微服务拆分策略

**Decision**: 从 Gin 单体拆分为 5 个独立微服务（user/product/order/payment/distribution），通过 NATS 异步通信，Traefik 统一路由。

**Rationale**:
- PRD §3.1.2 定义了 7 个微服务，一期聚焦 5 个核心服务（邮轮服务和营销服务在一期可作为 product-service 和独立模块处理）
- Constitution Principle II 要求按 DDD 域划分服务
- 分销域（distribution-service）是全新域，天然独立
- 供应商域归属 product-service（与产品管理紧密关联），签证域归属 order-service（与订单关联）

**Alternatives considered**:
- 保留单体 + 模块化：被拒绝，无法满足 Constitution Principle II 的跨域通信要求
- 全部 7 个微服务一步到位：被拒绝，Progressive Delivery 原则要求渐进式拆分

### R-002: NATS 消息队列使用模式

**Decision**: NATS Core 处理即时 Pub/Sub，JetStream 提供持久化消息流。一期三类异步事件：佣金计算、签证状态变更通知、对账任务。

**Rationale**:
- PRD §3.2.3 明确 NATS 作为消息队列首选（Windows 原生支持、极简部署）
- Constitution 要求跨域通信使用 NATS 异步
- JetStream 的至少一次投递保证适合佣金计算等不能丢失的事件
- 佣金计算由 order-service 发布事件，distribution-service 消费处理

**Alternatives considered**:
- RabbitMQ：被拒绝，需安装 Erlang，Windows 路径不支持中文或空格
- Apache Pulsar：被拒绝，不支持原生 Windows

### R-003: Meilisearch 索引同步策略

**Decision**: 通过 Asynq 异步任务将产品变更同步至 Meilisearch，延迟控制在 5 秒内。一期仅同步产品搜索，订单和用户搜索暂不迁移。

**Rationale**:
- PRD §3.2.2 明确 Meilisearch 作为搜索引擎
- 出境游产品筛选需支持大洲→国家层级树、签证类型等多维度过滤，Meilisearch 的分面过滤（faceting）天然适合
- 渐进式迁移降低风险

**Alternatives considered**:
- Elasticsearch：功能更强但部署复杂，Windows 支持不如 Meilisearch 轻量
- 数据库全文搜索（tsvector）：性能不足，无法支持多维度过滤和 typo 容错

### R-004: 银联支付 SDK 选型

**Decision**: 采用 smartwalle/unionpay Go SDK，覆盖消费、查询、撤销、退货核心功能。

**Rationale**:
- PRD §11.1.3 和表 5-1 明确银联接入基于 5.1.0 版本规范
- smartwalle/unionpay 已实现核心接口（100+ Stars），与 smartwalle/alipay 同一作者
- 系统架构采用渠道网关适配器模式，银联作为新渠道接入不影响已有支付逻辑

**Alternatives considered**:
- 自行封装银联 SDK：工作量大且签名/验签逻辑复杂，社区 SDK 已覆盖
- 暂不接入银联：被拒绝，spec 要求一期支持银联支付

### R-005: 电子合同签署方案

**Decision**: 一期支持合同模板管理和 PDF 生成，电子签章对接第三方 CA 服务（e签宝或法大大）。

**Rationale**:
- PRD §7.1.3 要求 CA 认证电子签章，满足《电子签名法》第 13 条
- e签宝和法大大是国内主流电子签章服务商，均提供 Go SDK
- 具体服务商在实施计划阶段确定，架构上通过适配器模式隔离

**Alternatives considered**:
- 自建 CA 签章系统：成本高、合规复杂，不适合一期
- 仅 PDF 生成不做电子签章：不满足 PRD 合规要求

### R-006: OCR 识别方案

**Decision**: 采用第三方 OCR 服务（百度 OCR 或腾讯 OCR），一期支持中国护照和常见国家护照识别。

**Rationale**:
- PRD §4.3.3 和 F-V-005 要求护照 OCR 识别
- 百度 OCR 和腾讯 OCR 均提供护照识别 API，准确率 ≥ 95%
- 小语种护照降级为人工录入（spec 假设）

**Alternatives considered**:
- 自建 OCR 模型：成本高、维护复杂
- 全部人工录入：用户体验差，不符合 spec 要求

### R-007: 防薅羊毛技术方案

**Decision**: 五层防护：自购禁止（用户ID比对）、身份隔离（分销商-消费者冲突检测）、设备关联（设备指纹30天窗口）、IP频率限制（同IP同链接1小时>10次）、违规模式识别（数据模型+人工审核）。

**Rationale**:
- PRD §8.7.2 明确定义五项防薅羊毛规则
- 设备指纹通过浏览器 Canvas/WebGL 指纹 + 小程序设备信息生成
- IP 频率限制使用 Redis 滑动窗口计数器
- 违规模式识别使用定时任务分析订单数据

**Alternatives considered**:
- 仅靠人工审核：效率低、无法实时拦截
- 引入专业风控系统（如阿里云风控）：成本高，一期先自研基础规则

### R-008: 分销佣金计算引擎设计

**Decision**: 佣金计算通过 NATS 事件驱动异步执行，订单支付成功后 order-service 发布事件，distribution-service 消费并计算。三级优先级规则（产品>品类>全局）通过 Redis 缓存加速匹配。

**Rationale**:
- PRD §8.7.1 定义佣金计算规则：基数规则、比例规则、归属规则、上限规则
- 异步计算避免阻塞支付回调处理
- Redis 缓存佣金规则，5 分钟内生效（spec 要求）

**Alternatives considered**:
- 同步计算：可能阻塞支付回调，影响支付成功率
- 数据库直接查询规则：性能不如 Redis 缓存

### R-009: 定金+尾款支付模式实现方案

**Decision**: 采用两笔独立支付方案，通过 payment_order 表的 biz_order_no 字段关联至同一业务订单。Asynq 定时任务处理尾款提醒和逾期取消。

**Rationale**:
- PRD §5.2.2 明确定义两笔独立支付方案
- 定金和尾款分别记录支付流水，财务对账清晰
- 尾款提醒通过 Asynq 定时任务在开放前 3 天触发

**Alternatives considered**:
- 单笔支付分阶段确认：复杂度高，不支持部分退款场景
- 预授权模式：银联/支付宝预授权接口限制较多

### R-010: 抖音小程序适配方案

**Decision**: 与微信小程序共享 Uni-App 代码基，通过条件编译（#ifdef MP-TOUTIAO）适配抖音 API。核心差异点：登录（抖音.login vs wx.login）、支付（tt.pay vs wx.requestPayment）、分享。

**Rationale**:
- PRD §3.3.3 明确 Uni-App 跨端框架支持抖音小程序
- Constitution 要求前端使用 Uni-App (Vue 3)
- 条件编译是 Uni-App 标准实践，代码复用率可达 80%+

**Alternatives considered**:
- 独立开发抖音小程序：代码重复度高、维护成本大
- 使用 Taro 框架：运行时方案性能不如 Uni-App 编译时方案
