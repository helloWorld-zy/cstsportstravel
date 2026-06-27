# Spec-Kit 全流程执行 Prompt 集

> 本文件包含旅游预订系统（OTA 开放平台）使用 Spec-Kit 工作流的完整 Prompt 集。
> 共 18 个步骤 + 2 个贯穿全阶段的持续步骤，按顺序依次执行。
>
> **使用方式**：将每个步骤的 Prompt 内容完整复制后发送给 Claude Code 即可触发对应 Skill。
>
> **重要**：每个步骤均标注了**必须阅读的文档**，确保 AI 在执行时始终以 PRD 为业务真相源，避免开发偏移。

---

## 目录

| 阶段 | 步骤 | Skill | 说明 | 必读文档 |
|:---|:---|:---|:---|:---|
| Phase 0 | 步骤 1 | `/speckit-constitution` | 建立项目宪法 | PRD 全文概览 |
| Phase 0 | 步骤 2 | `/speckit-clarify` | 澄清需求冲突与缺失 | PRD 全文 |
| Phase 1 | 步骤 3 | `/speckit-specify` | MVP 功能规格定义 | PRD §4.1, §4.2, §10 |
| Phase 1 | 步骤 4 | `/speckit-plan` | MVP 实施计划 | PRD §3, §5, §10, §12 |
| Phase 2 | 步骤 5 | `/speckit-tasks` | MVP 任务拆解 | spec.md, plan.md, data-model.md, contracts/, research.md, PRD §12 |
| Phase 2 | 步骤 6 | `/speckit-analyze` | MVP 跨文档一致性分析 | spec.md, plan.md, tasks.md, data-model.md, contracts/, PRD §4.1, §4.2 |
| Phase 2 | 步骤 7 | `/speckit-implement` | MVP 代码实现（按 Phase 1-12 分批，每次一个 Phase） | plan.md, tasks.md, data-model.md, contracts/, PRD 对应章节 |
| Phase 2 | 步骤 8 | `/speckit-checklist` | MVP 交付验收检查 | spec.md, contracts/, quickstart.md, PRD §4.1, §4.2, §10 |
| Phase 3 | 步骤 9 | `/speckit-specify` | 一期功能规格定义 | PRD §4.3, §5, §6, §7, §8 |
| Phase 3 | 步骤 10 | `/speckit-plan` | 一期实施计划 | spec.md, PRD §3, §5, §11, §12 |
| Phase 3 | 步骤 11 | `/speckit-tasks` | 一期任务拆解 | spec.md, plan.md, data-model.md, contracts/, research.md, PRD §7, §8 |
| Phase 3 | 步骤 12 | `/speckit-analyze` | 一期跨文档一致性分析 | spec.md, plan.md, tasks.md, data-model.md, contracts/, PRD §4.3, §7, §8 |
| Phase 3 | 步骤 13 | `/speckit-implement` | 一期代码实现 | plan.md, tasks.md, data-model.md, contracts/, quickstart.md, PRD §5, §11 |
| Phase 4 | 步骤 14 | `/speckit-specify` | 二期功能规格定义 | PRD §4.4, §6, §9, §10 |
| Phase 4 | 步骤 15 | `/speckit-plan` | 二期实施计划 | spec.md, PRD §3, §9, §12 |
| Phase 4 | 步骤 16 | `/speckit-tasks` | 二期任务拆解 | spec.md, plan.md, data-model.md, contracts/, research.md, PRD §4.4, §9 |
| Phase 4 | 步骤 17 | `/speckit-analyze` | 二期跨文档一致性分析 | spec.md, plan.md, tasks.md, data-model.md, contracts/, PRD §4.4, §9, §10 |
| Phase 4 | 步骤 18 | `/speckit-implement` | 二期代码实现 | plan.md, tasks.md, data-model.md, contracts/, quickstart.md, PRD §9, §10, §12 |
| 贯穿 | 持续 A | `/speckit-agent-context-update` | 更新 Agent 上下文 | — |
| 贯穿 | 持续 B | `/speckit-converge` | 收敛检查 | spec.md, plan.md, tasks.md |

---

## PRD 文档索引

> **PRD 路径**：`docs/travel_booking_system_v3.agent.final.md`（约 312KB，分 13 章）

| 章节 | 内容 | 对应阶段 |
|:---|:---|:---|
| §2 总体描述 | 系统边界、用户角色、运行环境、设计约束 | 全阶段参考 |
| §3 系统架构 | 分层架构、微服务划分、技术栈、部署架构 | Phase 1-2 |
| §4.1 通用功能 | 用户注册登录、个人中心、搜索发现、首页 | MVP |
| §4.2 境内游跟团游 | 产品列表/详情、预订流程、订单管理、业务规则 | MVP |
| §4.3 出境游跟团游 | 签证服务、护照管理、行前信息 | 一期 |
| §4.4 邮轮游 | 搜索、航次选择、舱房预订、值船流程 | 二期 |
| §5 支付系统 | 支付渠道接入、支付模式、退款管理、财务对账 | MVP + 一期 |
| §6 后台管理系统 | 产品管理、订单管理、财务管理、权限管理、营销、报表 | 全阶段 |
| §7 供应商/开放平台 | 供应商入驻、工作台、佣金结算 | 一期 |
| §8 二级分销 | 分销关系、分销商管理、推广跟踪、佣金管理 | 一期 |
| §9 多租户管理 | 租户生命周期、数据隔离、品牌定制 | 二期 |
| §10 非功能需求 | 安全、性能、可用性、可维护性、兼容性 | 全阶段 |
| §11 外部接口 | 支付接口、地图方案、消息推送 | 一期 + 二期 |
| §12 数据库设计 | 核心实体关系、索引设计、分区策略、RLS | 全阶段 |
| §13 风险与实施 | 技术风险、业务风险、三阶段实施建议 | 全阶段参考 |

---

## Phase 0：项目初始化

### 步骤 1：建立项目宪法

**Skill**：`/speckit-constitution`

**必读文档**：PRD 全文概览（重点 §2 总体描述、§10 非功能需求、§13 风险与实施建议）

**Prompt**：

```
/speckit-constitution

请为旅游预订系统（OTA 开放平台）建立项目宪法。

执行前请先阅读以下文档：
1. docs/travel_booking_system_v3.agent.final.md（PRD，重点 §2、§10、§13）
2. .specify/templates/constitution-template.md（宪法模板）

以下是核心约束和技术决策：

## 项目信息
- 项目名称：旅游预订系统（Travel Booking OTA Platform）
- 项目定位：面向 C 端消费者的在线旅游预订 OTA 开放平台
- 首期 MVP 聚焦"境内跟团游完整交易闭环"

## 技术栈约束（MUST）
- 后端：Go 1.26+，Gin Web 框架，GORM v2 + pgx
- 数据库：PostgreSQL 17.x（开发阶段），后续验证 18+ 兼容性
- 缓存：Redis 7.2+（生产环境 Memurai，开发环境 Docker Redis）
- 搜索：Meilisearch 1.19+
- 消息队列：NATS 2.11+，异步任务 Asynq
- 前端：Web Nuxt.js 3（SSR），后台 Vue 3 + Element Plus（SPA），小程序 Uni-App
- 部署：Windows Server 2022/2025，Go 服务通过 WinSW 注册为 Windows 服务
- 网关：Traefik 3.x+
- 服务发现：Consul 1.22+
- 监控：Prometheus + Grafana + Jaeger
- CI/CD：GitHub Actions（Windows Runner）

## 安全基线（MUST）
- 等保三级基础要求
- TLS 1.3 全站加密
- 审计日志留存 ≥6 个月
- 字段级加密 AES-256-GCM
- RBAC 权限模型（功能权限 + 数据权限 + 字段权限）
- MFA 多因素认证（管理员敏感操作）
- JWT RS256 非对称签名，Access Token 15 分钟有效期

## 架构原则
- DDD 领域驱动设计，按限界上下文划分微服务
- 多租户架构：共享数据库 + tenant_id 字段隔离 + PostgreSQL RLS
- API 优先：OpenAPI 3.0 规范，前后端契约先行
- 渐进式交付：MVP（境内游）→ 一期（出境游+供应商）→ 二期（邮轮+数据分析+微服务完整化）

## 编码规范
- Go：遵循 Uber Go Style Guide，golangci-lint 静态检查
- 前端：ESLint + Prettier，TypeScript 严格模式
- 提交规范：Conventional Commits
- 测试：单元测试覆盖率 ≥70%，核心交易链路集成测试覆盖

请根据以上信息填充项目宪法模板，确保所有原则都是可验证的、明确的。
```

---

### 步骤 2：澄清需求冲突与缺失

**Skill**：`/speckit-clarify`

**必读文档**：PRD 全文（`docs/travel_booking_system_v3.agent.final.md`）

**Prompt**：

```
执行前请完整阅读 docs/travel_booking_system_v3.agent.final.md（PRD 全文）。

请对 PRD 文档中的以下已知问题进行澄清。

## 已知需求冲突（必须澄清）

### C01：退款审批金额阈值不一致
- 6.2.3 退改审核流程定义分级：500/5000/20000 元
- 2.2 节用户角色定义：1000/5000 元
→ 需确认：采用哪套阈值？

### C02：支付倒计时时长不一致
- 4.2.3 预订流程：30 分钟
- 5.2.1 支付模式：短线 15 分钟，长线 30 分钟
→ 需确认：境内游（短线）实际倒计时是 15 分钟还是 30 分钟？

### C03：分销体系优先级标注矛盾
第 8 章各节均标注"P0（一期核心功能）"，但总体架构将分销归入一期。MVP 是否需要分销的最小子集？
→ 需确认：MVP 阶段是否包含分销功能？

## 已知需求缺失（建议澄清）

### M01：附加服务产品模型
4.2.3 提及旅游保险、接送机、当地玩乐三类附加服务，但未定义产品模型、定价规则、与主订单的关系。
→ 附加服务作为子订单还是订单行项目？是否需要对接保险公司 API？

### M02：评价系统详细需求
4.2.2 F-I-D14 提及用户评价，但未定义评价提交流程、审核机制、与供应商评分的关系。
→ 评价是否需要审核？差评是否触发供应商预警？

### M03：库存并发控制方案
6.1.10 定义了库存预扣机制，但未明确高并发场景下的实现方案。
→ 使用 Redis 原子操作还是数据库行锁？超售防护策略？

### M04：小程序账号体系统一方案
4.1.1 提及微信授权登录和手机号注册，但未定义 openid 与 user_id 的绑定策略。
→ 小程序用户如何与 Web 用户打通？是否强制绑定手机号？

请逐项给出明确的决策结果。
```

---

## Phase 1：MVP 规格定义

### 步骤 3：MVP 功能规格定义

**Skill**：`/speckit-specify`

**必读文档**：
- PRD §4.1 通用功能（用户注册登录、个人中心、搜索发现、首页）
- PRD §4.2 境内游跟团游（产品列表/详情、预订流程、订单管理、业务规则）
- PRD §6 后台管理系统（产品管理、订单管理、权限管理）
- PRD §10 非功能需求（安全、性能、可用性）
- PRD §12 数据库设计（核心实体关系）

**Prompt**：

```
/speckit-specify

境内跟团游 MVP 功能规格定义。

执行前请先阅读以下 PRD 章节（docs/travel_booking_system_v3.agent.final.md）：
- §4.1 通用功能（三端共用）
- §4.2 境内游跟团游
- §6 后台管理系统（重点关注 §6.1 产品管理、§6.2 订单管理、§6.4 权限管理）
- §10 非功能需求
- §12 数据库设计要点（重点关注 §12.1 核心实体关系）

同时阅读 .specify/memory/constitution.md 了解项目宪法约束。

## 功能概述

本 MVP 为旅游预订系统的第一阶段交付，聚焦境内跟团游的完整交易闭环。目标是验证商业模式和技术架构，支撑日均 300 单的处理量。

## 核心用户故事

### US1：用户注册与登录
- 作为消费者，我可以通过手机号+验证码注册和登录平台
- 作为消费者，我可以通过微信授权快捷登录
- 作为消费者，我可以完成实名认证（姓名+身份证号）
→ 详细流程参考 PRD §4.1.1 用户注册与登录

### US2：产品搜索与浏览
- 作为消费者，我可以按目的地、出发城市、天数、价格筛选境内游产品
- 作为消费者，我可以查看产品详情（行程、费用说明、退改政策、团期日历）
- 作为消费者，我可以查看每日价格和库存状态
→ 详细功能点参考 PRD §4.2.1 产品列表与筛选、§4.2.2 产品详情页

### US3：预订下单
- 作为消费者，我可以选择团期和人数（成人/儿童/婴儿）
- 作为消费者，我可以填写出游人信息（支持从常用出游人快速选择）
- 作为消费者，我可以确认订单并选择支付方式（支付宝/微信）
- 作为消费者，我需要在 30 分钟内完成支付，否则订单自动取消
→ 详细流程参考 PRD §4.2.3 预订流程，业务规则参考 PRD §4.2.5

### US4：订单管理
- 作为消费者，我可以查看我的订单列表（按状态筛选）
- 作为消费者，我可以查看订单详情和支付状态
- 作为消费者，我可以申请退款
→ 详细功能参考 PRD §4.2.4 订单管理

### US5：后台产品管理
- 作为供应商，我可以发布境内游产品（基础信息/行程/价格/团期/库存）
- 作为运营人员，我可以审核产品（通过/驳回）
- 作为运营人员，我可以管理价格日历和库存
→ 详细功能参考 PRD §6.1.1 境内游产品发布、§6.1.8 产品审核流程、§6.1.9 价格日历管理、§6.1.10 库存管理

### US6：后台订单管理
- 作为运营人员，我可以查询和处理订单
- 作为运营人员，我可以审核退款申请
- 作为运营人员，我可以配置退改规则（阶梯费率）
→ 详细功能参考 PRD §6.2 订单管理（§6.2.1 列表查询、§6.2.2 状态流转、§6.2.3 退改审核、§6.2.4 退改规则配置）

### US7：基础权限管理
- 作为系统管理员，我可以管理角色和权限（RBAC）
- 作为系统管理员，我可以创建运营人员和供应商账号
→ 详细功能参考 PRD §6.4.3 运营人员管理，RBAC 矩阵参考 PRD 表6-8

## 非功能需求
- 响应时间：产品列表 P99 ≤200ms，订单确认 P99 ≤500ms
- 并发：支撑 10,000 并发用户
- 安全：TLS 1.3、JWT RS256、密码策略（8位+复杂度）、登录失败锁定（5次/15分钟）、审计日志、字段级加密
- 可用性：99.9%
→ 详细指标参考 PRD §10.2 性能需求（表9-2）、§10.1 安全需求

## 约束
- MVP 阶段不引入 Meilisearch，使用 PostgreSQL 全文搜索
- MVP 阶段仅支持全额支付，不支持定金+尾款
- MVP 阶段仅支持微信小程序，不支持支付宝/抖音小程序
- MVP 阶段不包含供应商自助入驻，由平台代为录入
- MVP 阶段不包含分销功能
- MVP 阶段不包含优惠券和营销活动

## 前端页面清单（必须在 spec 中明确列出）

每个用户故事必须同时定义后端 API 和前端页面/组件。以下为 MVP 阶段必须交付的前端页面：

### Web 销售平台（Nuxt.js 3 SSR）
- 首页（金刚区导航、Banner 轮播、热门目的地、推荐产品）
- 产品列表页（筛选栏、产品卡片、排序、分页）
- 产品详情页（图片轮播、行程详情、费用说明、退改政策、团期日历、评价区域）
- 预订流程页（四步向导：选团期→填出游人→附加服务→确认支付）
- 支付页面（支付宝/微信支付选择、30 分钟倒计时）
- 个人中心页（订单列表、订单详情、退款申请、常用出游人管理、实名认证）
- 登录/注册页（手机号+验证码、微信授权）

### 微信小程序（Uni-App）
- 与 Web 端共享核心页面，通过条件编译适配小程序交互
- 小程序登录页（wx.login + 手机号快捷登录）
- 产品列表/详情/预订流程（小程序适配版本）
- 订单管理（小程序版本）
- 支付流程（wx.requestPayment 调起微信支付）

### 后台管理系统（Vue 3 + Element Plus SPA）
- 登录页
- 产品管理页（产品列表、发布/编辑表单、行程编辑器、价格日历、库存管理）
- 产品审核页（审核列表、审核操作）
- 订单管理页（订单列表、订单详情、退款审核）
- 退改规则配置页（阶梯费率编辑器）
- 用户管理页（运营人员/供应商账号列表、角色分配）
- 权限管理页（角色列表、菜单权限配置）

请基于以上信息生成功能规格文档。确保 spec 中的功能需求覆盖 PRD 对应章节的所有关键业务规则，特别是：
- PRD §4.2.5 中的单房差规则、儿童价规则、超时未支付规则、退改规则
- PRD §6.2.2 中的订单状态机定义（表6-5）
- PRD §6.2.4 中的退改阶梯费率配置（表6-6）

**前端要求**：spec 中的每个用户故事必须同时包含"前端页面/组件"和"后端 API"两个维度的验收场景。禁止只定义 API 不定义页面。前端页面清单（上述"前端页面清单"章节）中的每个页面都必须在对应的用户故事中有功能需求覆盖。
```

---

### 步骤 4：MVP 实施计划

**Skill**：`/speckit-plan`

**必读文档**：
- `specs/001-domestic-tour-mvp/spec.md`（步骤 3 生成的 MVP 规格）
- PRD §3 系统架构（分层架构、微服务划分、技术栈）
- PRD §5 支付系统（支付渠道接入、支付模式）
- PRD §10 非功能需求（安全、性能、可用性）
- PRD §12 数据库设计（核心实体关系、索引设计、分区策略）
- `.specify/memory/constitution.md`（项目宪法）

**Prompt**：

```
/speckit-plan

请为境内跟团游 MVP 制定实施计划。

执行前请先阅读以下文档：
1. specs/001-domestic-tour-mvp/spec.md（MVP 功能规格）
2. .specify/memory/constitution.md（项目宪法）
3. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §3 系统架构（分层架构、微服务划分、技术栈选型）
   - §5 支付系统（§5.1 支付渠道接入、§5.2 支付模式）
   - §10 非功能需求（§10.1 安全需求、§10.2 性能需求、§10.3 可用性）
   - §12 数据库设计（§12.1 核心实体关系、§12.2 关键索引设计、§12.3 分区策略）

## 架构决策

### 部署架构
- 采用 Gin 单体架构（非微服务），所有业务模块编译为单个可执行文件
- 部署于 Windows Server 2022/2025
- 通过 WinSW 注册为 Windows 服务
- Traefik 作为反向代理和 API 网关

### 数据层
- PostgreSQL 17.x 主从架构（一主一从）
- Redis 7.2+ 用于会话管理和热点数据缓存
- MVP 阶段暂不引入 Meilisearch

### 前端架构

#### Web 销售平台（Nuxt.js 3 SSR）
- 框架：Nuxt.js 3 + Vue 3 + TypeScript
- UI 库：Element Plus
- 状态管理：Pinia + @tanstack/vue-query
- 路由：Nuxt 文件系统路由 + 动态路由（产品详情、订单详情）
- SSR 策略：产品列表/详情页 SSR（SEO），个人中心/订单管理 SPA 模式
- 页面清单：首页、产品列表、产品详情、预订流程（四步向导）、支付页、个人中心、订单列表/详情、登录/注册
- 组件设计：ProductCard、DepartureCalendar、PriceCalendar、TravellerForm、OrderStatusTag、PaymentCountdown 等核心组件

#### 微信小程序（Uni-App）
- 框架：Uni-App (Vue 3) + TypeScript
- UI 库：uView UI 2.x
- 与 Web 端共享业务逻辑层（API 调用、数据转换、表单校验）
- 条件编译：`#ifdef MP-WEIXIN` 适配微信小程序特有 API
- 页面清单：登录页、首页、产品列表/详情、预订流程、支付流程、订单管理

#### 后台管理系统（Vue 3 + Element Plus SPA）
- 框架：Vue 3 + TypeScript + Vite
- UI 库：Element Plus
- 状态管理：Pinia
- 路由：vue-router 4.x 动态权限路由（基于 RBAC 菜单权限）
- 页面清单：登录页、产品管理（列表/发布/编辑/审核）、订单管理（列表/详情/退款审核）、退改规则配置、用户/角色/权限管理

#### 前端工程化
- 三端共享：API 类型定义（从 OpenAPI 自动生成 TypeScript 类型）、工具函数库（日期/金额/证件校验）、设计 Token（CSS 变量主题色）
- 构建：Vite 6，TypeScript 严格模式
- 代码规范：ESLint + Prettier，组件命名 PascalCase

### 支付
- 支付宝（PC 网站支付 + 手机网站支付）
- 微信支付（Native + JSAPI）
- 使用 smartwalle/alipay v3 和 wechatpay-go SDK

### 安全
- JWT RS256 非对称签名
- AES-256-GCM 字段级加密（身份证号、手机号）
- TLS 1.3（Traefik SSL 终结）
- 审计日志（zap + lumberjack 轮转）

## 模块划分（建议）
1. user-service：用户注册登录、实名认证、个人中心
2. product-service：产品管理、行程编辑、价格日历、库存管理、审核流程
3. order-service：订单创建、状态机、退改申请
4. payment-service：支付网关、回调处理、退款
5. admin-service：后台管理 API（产品审核、订单处理、权限管理）
6. common：共享组件（数据库连接、缓存、日志、中间件）

## 数据库 Schema 核心表
→ 参考 PRD §12.1 核心实体关系（ER 图）和 §12.2 关键索引设计
- user_account、real_name_verification、frequent_traveller
- product、itinerary、departure_date、price_rule、refund_rule
- main_order、sub_order、order_status_log
- payment_transaction、refund_record
- admin_user、role、permission、menu

请基于以上决策生成详细的实施计划，包括：
- 技术选型确认
- 数据模型设计（ER 图和核心表结构）
- API 契约（RESTful 端点列表）
- 开发阶段划分
- 验证场景（quickstart）
```

---

## Phase 2：MVP 实施

### 步骤 5：MVP 任务拆解

**Skill**：`/speckit-tasks`

**必读文档**：
- `specs/001-domestic-tour-mvp/spec.md`（MVP 规格）
- `specs/001-domestic-tour-mvp/plan.md`（MVP 实施计划）
- `specs/001-domestic-tour-mvp/data-model.md`（数据模型，包含表结构、索引、状态机）
- `specs/001-domestic-tour-mvp/contracts/`（API 契约，包含所有端点定义）
- `specs/001-domestic-tour-mvp/research.md`（技术决策，包含并发控制、幂等性等方案）
- PRD §12 数据库设计（确认实体关系和索引设计与 tasks 一致）
- `.specify/memory/constitution.md`（项目宪法）

**Prompt**：

```
/speckit-tasks

请根据 MVP 的 spec.md、plan.md 和相关设计文档生成可执行的任务列表。

执行前请先阅读以下文档：
1. specs/001-domestic-tour-mvp/spec.md（MVP 功能规格）
2. specs/001-domestic-tour-mvp/plan.md（MVP 实施计划）
3. specs/001-domestic-tour-mvp/data-model.md（数据模型，25+ 张表的字段定义和索引）
4. specs/001-domestic-tour-mvp/contracts/（API 契约，40+ 端点的请求/响应定义）
5. specs/001-domestic-tour-mvp/research.md（技术决策，并发控制/幂等性/加密等方案）
6. .specify/memory/constitution.md（项目宪法）
7. docs/travel_booking_system_v3.agent.final.md §12 数据库设计要点（确认实体关系）

## 任务生成要求

1. 按用户故事组织任务，每个用户故事独立可测试
2. 任务必须包含明确的文件路径
3. 标注可并行执行的任务 [P]
4. 设置任务依赖关系
5. 确保每个功能需求（FR-###）至少有一个对应的实现任务
6. 确保 PRD §4.2.5 中的业务规则（单房差、儿童价、超时未付、退改阶梯费率）有对应的实现任务

## 阶段结构建议

⚠️ **关键原则：每个用户故事必须同时包含后端 API 实现和前端页面实现。禁止只实现后端不实现前端。**

- Phase 1：项目初始化（Go module、目录结构、配置管理、数据库连接、前端项目脚手架）
- Phase 2：基础组件（后端：中间件/认证/RBAC/审计日志 | 前端：三端项目初始化/路由框架/API 封装层/公共组件库）
- Phase 3：US1 用户注册与登录（后端 API + Web 登录页 + 小程序登录页 + 后台登录页）
- Phase 4：US2 产品搜索与浏览（后端 API + Web 产品列表/详情页 + 小程序产品列表/详情页 + 后台产品管理页）
- Phase 5：US3 预订下单（后端 API + Web 预订流程四步向导 + 小程序预订流程 + 支付页面）
- Phase 6：US4 订单管理（后端 API + Web 订单列表/详情/退款页 + 小程序订单管理）
- Phase 7：US5 后台产品管理（后端 API + 后台产品发布/编辑/审核/价格日历页面）
- Phase 8：US6 后台订单管理（后端 API + 后台订单列表/详情/退款审核/退改规则配置页面）
- Phase 9：US7 基础权限管理（后端 API + 后台用户/角色/权限管理页面）
- Phase 10：前端增强（首页金刚区/Banner/推荐位、搜索联想、个人中心完善）
- Phase 11：集成测试与安全加固
- Phase 12：部署与运维（WinSW、Traefik、CI/CD）

## 技术约束
- 每个任务必须指定目标文件路径
- Go 代码遵循 DDD 分层：handler → service → repository → model
- API 定义使用 OpenAPI 3.0 注释（swaggo/swag）
- 前端使用 TypeScript 严格模式
7. **前端任务要求**：
   - 每个用户故事的实现任务必须包含至少一个前端页面/组件任务
   - 前端任务必须指定目标文件路径（如 `web/pages/products/[id].vue`）
   - 前端任务必须包含 API 对接逻辑（调用后端接口、处理响应、错误处理）
   - 前端任务必须包含加载状态、空状态、错误状态的处理
   - 小程序页面必须使用条件编译处理平台差异
   - 后台管理页面必须包含权限控制（菜单可见性、按钮操作权限）

请生成完整的 tasks.md。
```

---

### 步骤 6：MVP 跨文档一致性分析

**Skill**：`/speckit-analyze`

**必读文档**：
- `specs/001-domestic-tour-mvp/spec.md`
- `specs/001-domestic-tour-mvp/plan.md`
- `specs/001-domestic-tour-mvp/tasks.md`
- `specs/001-domestic-tour-mvp/data-model.md`（对照检查表结构与 tasks 一致性）
- `specs/001-domestic-tour-mvp/contracts/`（对照检查 API 端点与 tasks 一致性）
- PRD §4.1, §4.2（对照 spec 中的功能需求是否遗漏 PRD 中的关键业务规则）
- PRD §10（对照非功能需求是否完整）
- `.specify/memory/constitution.md`（宪法合规检查）

**Prompt**：

```
/speckit-analyze

请对 MVP 阶段的 spec.md、plan.md、tasks.md 进行跨文档一致性分析。

执行前请先阅读以下文档：
1. specs/001-domestic-tour-mvp/spec.md
2. specs/001-domestic-tour-mvp/plan.md
3. specs/001-domestic-tour-mvp/tasks.md
4. specs/001-domestic-tour-mvp/data-model.md（数据模型）
5. specs/001-domestic-tour-mvp/contracts/（API 契约）
6. .specify/memory/constitution.md（项目宪法）
7. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §4.1 通用功能（对照检查用户体系是否完整）
   - §4.2 境内游跟团游（对照检查产品/预订/订单流程是否完整）
   - §10 非功能需求（对照检查安全/性能需求是否覆盖）

## 分析重点

1. **PRD 覆盖度**：spec.md 是否遗漏了 PRD §4.1/§4.2 中的关键业务规则（如单房差自动计算、儿童关联成人、身份证校验码验证等）
2. **需求覆盖**：spec.md 中的所有功能需求（FR-###）是否在 tasks.md 中有对应任务
3. **用户故事覆盖**：所有用户故事（US1-US7）的验收场景是否被任务覆盖
4. **宪法合规**：plan.md 和 tasks.md 是否违反项目宪法中的 MUST 原则
5. **术语一致性**：三份文档中的实体名称、状态定义是否一致
6. **任务可执行性**：tasks.md 中的任务是否包含足够的文件路径和依赖说明
7. **成功标准映射**：spec.md 中的成功标准（SC-###）是否有对应的验证任务
8. **前端覆盖度（CRITICAL）**：
   - 每个用户故事是否有对应的前端页面/组件任务
   - 前端页面清单中的每个页面是否在 tasks.md 中有实现任务
   - 前端任务是否包含 API 对接、状态管理、错误处理
   - 三端（Web/小程序/后台）的页面是否都有覆盖
   - 前端任务是否有明确的文件路径

请输出结构化分析报告，标注每个发现的严重级别（CRITICAL/HIGH/MEDIUM/LOW）。
```

---

### 步骤 7：MVP 代码实现

**Skill**：`/speckit-implement`

**执行策略**：按 Phase 分批实现，每个 Phase 单独一次会话。tasks.md 共 12 个 Phase（164 个任务），每次只实现一个 Phase 的全部任务，完成后标记 [X] 并运行对应的 Checkpoint 验证。

**通用必读文档**（每次会话都需阅读）：
- `specs/001-domestic-tour-mvp/plan.md`（实施计划）
- `specs/001-domestic-tour-mvp/tasks.md`（任务列表，定位到当前 Phase）
- `.specify/memory/constitution.md`（编码规范和安全基线）

**通用实现要求**（适用于所有 Phase）：

1. 每完成一个任务，在 tasks.md 中标记为 [X]
2. 遵循 Go 编码规范：错误处理必须显式、zap 结构化日志、导出函数 godoc 注释、核心逻辑单元测试
3. 遵循前端编码规范：TypeScript 严格模式、组件 PascalCase、API 调用统一封装
4. 前端页面必须包含三种状态：Loading、正常显示、空状态/错误状态
5. 表单页面必须包含客户端校验（身份证号、手机号、必填项）
6. 小程序页面必须使用 `#ifdef` 条件编译处理微信特有 API
7. 后台管理页面必须实现动态权限路由和按钮级权限控制
8. 禁止只提交后端代码不提交前端代码

**代码结构（参考）**：

```
/cmd/server/main.go           # 入口
/internal/
  /user/                       # 用户模块
    /handler/                  # HTTP 处理器
    /service/                  # 业务逻辑
    /repository/               # 数据访问
    /model/                    # 数据模型
  /product/                    # 产品模块（同上分层）
  /order/                      # 订单模块
  /payment/                    # 支付模块
  /admin/                      # 后台管理模块
  /common/                     # 共享组件
    /middleware/               # 中间件（认证、RBAC、审计、签名、脱敏）
    /auth/                     # TOTP MFA
    /database/                 # 数据库连接
    /cache/                    # Redis 缓存
    /encrypt/                  # AES-256-GCM 字段加密
    /logger/                   # 日志
    /config/                   # 配置（Viper + Consul KV）
    /response/                 # 统一响应封装 + 字段脱敏
/web/                          # Nuxt.js Web 前端
/miniapp/                      # Uni-App 小程序
/admin-web/                    # Vue 3 后台管理
```

---

**以下为各 Phase 的独立 Prompt，每次复制一个发送：**

### Phase 1：项目初始化

```
/speckit-implement

请实现 tasks.md 中 Phase 1（Setup）的全部任务（T001-T015）。

执行前先阅读：
1. specs/001-domestic-tour-mvp/plan.md（项目结构和技术选型）
2. specs/001-domestic-tour-mvp/data-model.md（表结构，编写迁移文件时必须对照）
3. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 1 部分）

Phase 1 包含：Go module 初始化、配置系统（Viper）、数据库连接（GORM+pgx）、Redis 客户端、日志（zap）、统一响应封装、5 个数据库迁移文件、main.go 入口、三端前端项目脚手架。

完成后：
1. 将已完成任务标记为 [X]
2. 验证：后端能启动并连接 PostgreSQL + Redis，三端前端项目能启动
```

### Phase 2：基础组件

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 2（Foundational）的全部任务（T016-T036）。

执行前先阅读：
1. specs/001-domestic-tour-mvp/plan.md（安全架构、配置管理）
2. specs/001-domestic-tour-mvp/data-model.md（GORM 模型定义时对照表结构）
3. specs/001-domestic-tour-mvp/research.md（字段加密、请求签名方案）
4. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 2 部分）
5. .specify/memory/constitution.md（Principle III 安全基线）

Phase 2 包含：JWT RS256、认证中间件、RBAC 中间件、审计日志中间件、限流中间件、AES-256-GCM 字段加密、TOTP MFA 服务、MFA 中间件、API 响应脱敏、HMAC-SHA256 请求签名、Consul KV 动态配置、GORM 模型（5 个域）、路由框架、三端前端路由/API 层/公共组件。

完成后：
1. 将已完成任务标记为 [X]
2. 验证：认证中间件能拦截未登录请求，RBAC 中间件能校验权限，加密/脱敏工具可用
3. Checkpoint：所有基础组件就绪，用户故事可以开始实现

```

### Phase 3：US1 用户注册与登录

```
/speckit-implement
请依照TDD模式实现 tasks.md 中 Phase 3（US1 用户注册与登录）的全部任务（T037-T056）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US1 验收场景，FR-001~FR-006）
3. specs/001-domestic-tour-mvp/contracts/user-api.yaml（用户 API 端点定义）
4. specs/001-domestic-tour-mvp/data-model.md（user_account、real_name_verification、frequent_traveller 表）
5. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 3 部分）
6. docs/travel_booking_system_v3.agent.final.md §4.1.1（用户注册与登录流程）

Phase 3 包含：
- 后端：短信验证码服务、用户注册/登录 API、微信 OAuth、实名认证、常用出游人 CRUD、管理员登录
- Web：登录/注册页、个人中心、实名认证表单、常用出游人管理
- 小程序：wx.login 登录页
- 后台：管理员登录页、auth store

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 验证场景 VS1（用户注册与登录）
3. Checkpoint：用户可在 Web 和小程序完成注册→登录→实名认证
```

### Phase 4：US2 产品搜索与浏览

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 4（US2 产品搜索与浏览）的全部任务（T057-T073）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US2 验收场景，FR-007~FR-013）
3. specs/001-domestic-tour-mvp/contracts/product-api.yaml（产品 API 端点定义）
4. specs/001-domestic-tour-mvp/data-model.md（product、itinerary、departure_date、price_rule、category、product_review 表）
5. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 4 部分）
6. docs/travel_booking_system_v3.agent.final.md §4.2.1（产品列表与筛选）、§4.2.2（产品详情页）

Phase 4 包含：
- 后端：产品列表 API（筛选/排序/分页）、产品详情 API、团期日历 API、行程 API、评价 API、搜索联想 API、首页数据 API
- Web：首页（金刚区/Banner/推荐位）、产品列表页（筛选栏/排序/卡片）、产品详情页（行程/费用/退改/团期日历/评价）、DepartureCalendar 组件
- 小程序：首页、产品列表、产品详情
- 后台：产品列表页、首页配置页

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 验证场景 VS2（产品浏览）
3. Checkpoint：首页展示内容，产品列表筛选/排序可用，产品详情页完整展示，团期日历显示价格和库存
```

### Phase 5：US3 预订下单

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 5（US3 预订下单）的全部任务（T074-T096）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US3 验收场景，FR-014~FR-025）
3. specs/001-domestic-tour-mvp/contracts/order-api.yaml（订单 API）
4. specs/001-domestic-tour-mvp/contracts/payment-api.yaml（支付 API）
5. specs/001-domestic-tour-mvp/data-model.md（main_order、order_traveller、payment_transaction 表）
6. specs/001-domestic-tour-mvp/research.md（库存并发控制、支付幂等性、订单自动取消方案）
7. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 5 部分）
8. docs/travel_booking_system_v3.agent.final.md §4.2.3（预订流程）、§4.2.5（业务规则：单房差/儿童价/超时未付）、§5.1（支付渠道接入）、§5.2（支付模式）

Phase 5 包含：
- 后端：库存服务（Redis 原子操作+DB 行锁）、订单创建（含单房差/儿童价计算）、支付宝/微信支付集成、支付回调（幂等）、30 分钟超时自动取消（Asynq）、支付成功流程
- Web：预订四步向导（选团期→填出游人→附加服务→确认）、支付页面（倒计时）、PaymentCountdown 组件
- 小程序：预订流程、wx.requestPayment 支付

⚠️ 关键业务规则（必须正确实现）：
- 单房差：成人数为奇数时自动附加 1 份（PRD §4.2.5）
- 儿童价：2-12 岁不占床，需关联成人（PRD §4.2.5）
- 超时未付：30 分钟后自动取消+释放库存（PRD §4.2.5）
- 支付幂等：DB 唯一约束 + Redis 去重（research.md）

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 验证场景 VS3（完整预订流程）和 VS4（支付超时）
3. Checkpoint：用户可从产品详情完成预订→支付全流程，单房差和儿童价计算正确，超时自动取消生效
```

### Phase 6：US4 订单管理

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 6（US4 订单管理）的全部任务（T097-T105c）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US4 验收场景，FR-017~FR-021）
3. specs/001-domestic-tour-mvp/contracts/order-api.yaml（退款相关端点）
4. specs/001-domestic-tour-mvp/data-model.md（refund_rule、refund_record 表，订单状态机 9 状态映射）
5. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 6 部分）
6. docs/travel_booking_system_v3.agent.final.md §4.2.4（订单管理）、§4.2.5（退改规则）、§6.2.3（退改审核）、§6.2.4（退改阶梯费率表 6-6）

Phase 6 包含：
- 后端：退改规则引擎（阶梯费率匹配）、退款服务（原路退回）、退款 API、订单状态自动流转、评价提交 API
- Web：订单列表页（6 状态 Tab）、订单详情页、退款申请组件、评价表单组件
- 小程序：订单列表、订单详情

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 验证场景 VS5（退款流程）
3. Checkpoint：用户可查看订单、申请退款、退款金额按阶梯费率正确计算、可提交评价
```

### Phase 7：US5 后台产品管理

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 7（US5 后台产品管理）的全部任务（T106-T116）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US5 验收场景，FR-007~FR-013）
3. specs/001-domestic-tour-mvp/contracts/admin-api.yaml（产品管理相关端点）
4. specs/001-domestic-tour-mvp/data-model.md（product 状态机 5 状态、departure_date、price_rule 表）
5. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 7 部分）
6. docs/travel_booking_system_v3.agent.final.md §6.1.1（境内游产品发布）、§6.1.8（产品审核流程）、§6.1.9（价格日历管理）、§6.1.10（库存管理）

Phase 7 包含：
- 后端：产品 CRUD、行程服务、价格日历（5 种批量调价模式）、团期/库存管理、产品审核流程（5 状态：draft→pending_review→approved/suspended/change_pending_review）
- 后台前端：产品多步表单、行程编辑器、价格日历组件、产品审核页

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 验证场景 VS6（后台产品管理）
3. Checkpoint：供应商可创建产品→提交审核→运营审核通过→C 端可见
```

### Phase 8：US6 后台订单管理

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 8（US6 后台订单管理）的全部任务（T117-T124）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US6 验收场景，FR-020~FR-021）
3. specs/001-domestic-tour-mvp/contracts/admin-api.yaml（订单/退款管理端点）
4. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 8 部分）
5. docs/travel_booking_system_v3.agent.final.md §6.2.1（订单列表与查询）、§6.2.3（退改审核流程）、§6.2.4（退改规则配置）

Phase 8 包含：
- 后端：后台订单查询（多维度筛选）、退款审核（分级审批：≤1000 运营/1000-5000 财务主管/>5000 总监）、退改规则模板 CRUD
- 后台前端：订单列表页、订单详情页、退款审核页、退改规则编辑器

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 验证场景 VS7（退款审核）
3. Checkpoint：运营可搜索订单、审核退款（分级审批生效）、配置退改规则模板
```

### Phase 9：US7 权限管理

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 9（US7 基础权限管理）的全部任务（T125-T133b）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US7 验收场景，FR-026~FR-028, FR-030）
3. specs/001-domestic-tour-mvp/contracts/admin-api.yaml（用户/角色/权限/菜单端点）
4. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 9 部分）
5. .specify/memory/constitution.md（Principle III MFA 要求）

Phase 9 包含：
- 后端：管理员用户 CRUD、角色 CRUD、权限管理、RBAC 服务（菜单树）、供应商数据隔离中间件
- 后台前端：用户管理页、角色管理页、权限树编辑器、MFA 注册/验证组件

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 验证场景 VS8（RBAC 权限控制）
3. Checkpoint：管理员可创建用户、分配角色、用户仅见授权菜单、供应商数据隔离生效、MFA 验证弹窗正常
```

### Phase 10：前端增强

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 10（前端增强）的全部任务（T134-T139）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/spec.md（US2 首页相关验收场景）
3. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 10 部分）

Phase 10 包含：Banner 管理 API、热门目的地 API、首页动态内容对接、个人中心完善、图片上传服务（OSS STS）、响应式设计。

完成后：
1. 将已完成任务标记为 [X]
2. Checkpoint：首页展示动态 Banner 和热门目的地，搜索联想可用，个人中心完整，移动端响应式正常
```

### Phase 11：集成测试与安全加固

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 11（集成测试与安全加固）的全部任务（T140-T148）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/quickstart.md（全部验证场景 VS1-VS8）
3. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 11 部分）
4. .specify/memory/constitution.md（Principle III 安全基线逐项检查）

Phase 11 包含：4 个集成测试（用户流程/预订流程/退款流程/支付幂等）、TLS 1.3 验证、字段加密验证、审计日志覆盖验证、密码策略验证、quickstart 全场景验证。

完成后：
1. 将已完成任务标记为 [X]
2. 运行 quickstart.md 全部 VS1-VS8 场景，确认全部通过
3. Checkpoint：所有集成测试通过，安全检查清单全部勾选
```

### Phase 12：部署与运维

```
/speckit-implement

请依照TDD模式实现 tasks.md 中 Phase 12（部署与运维）的全部任务（T149-T157）。

执行前先阅读：
1. .specify/memory/constitution.md（项目宪法，在实现过程中严格遵守其中的原则）
2. specs/001-domestic-tour-mvp/plan.md（部署架构、CI/CD 方案）
3. specs/001-domestic-tour-mvp/tasks.md（仅 Phase 12 部分）
4. docs/travel_booking_system_v3.agent.final.md §10.4.2（监控告警，表 9-5 的 16 项指标和阈值）

Phase 12 包含：WinSW 服务配置、Traefik 配置（TLS 1.3）、GitHub Actions CI/CD、部署脚本、Prometheus 指标端点、数据库备份脚本、Grafana 仪表盘、Prometheus 告警规则、Windows Exporter。

完成后：
1. 将已完成任务标记为 [X]
2. Checkpoint：服务可作为 Windows 服务运行，Traefik 路由正确，CI/CD 流水线可用，监控告警就绪
```

---

**全部 Phase 完成后**，运行 `/speckit-converge` 检查是否有遗漏任务，然后进入步骤 8 验收检查。

---

### 步骤 8：MVP 交付验收检查

**Skill**：`/speckit-checklist`

**必读文档**：
- `specs/001-domestic-tour-mvp/spec.md`（MVP 规格，验收标准来源）
- `specs/001-domestic-tour-mvp/contracts/`（API 契约，对照检查端点实现完整性）
- `specs/001-domestic-tour-mvp/quickstart.md`（验证场景，对照检查端到端流程）
- PRD §4.1, §4.2（对照检查功能完整性）
- PRD §10（对照检查安全合规和性能基线）
- `.specify/memory/constitution.md`（宪法合规检查）

**Prompt**：

```
/speckit-checklist

请为 MVP 交付生成验收检查清单。

执行前请先阅读以下文档：
1. specs/001-domestic-tour-mvp/spec.md（MVP 规格）
2. specs/001-domestic-tour-mvp/contracts/（API 契约，验证每个端点是否实现）
3. specs/001-domestic-tour-mvp/quickstart.md（验证场景，验证端到端流程是否通过）
2. .specify/memory/constitution.md（项目宪法）
3. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §4.1 通用功能（用户体系完整性）
   - §4.2 境内游跟团游（产品/预订/订单流程完整性）
   - §4.2.5 业务规则（单房差、儿童价、超时未付、退改规则）
   - §10 非功能需求（安全合规、性能基线）

## 检查维度

### 功能完整性
- 用户注册登录（手机号+验证码、微信授权）是否完整实现
- 实名认证流程是否完整
- 产品搜索/筛选/详情是否完整
- 预订流程四步向导是否完整
- 支付流程（支付宝+微信）是否端到端打通
- 订单管理（列表/详情/状态流转）是否完整
- 退款流程是否完整
- 后台产品管理（发布/审核/价格/库存）是否完整
- 后台订单管理（查询/处理/退改审核）是否完整
- RBAC 权限体系是否完整

### 安全合规（参考 PRD §10.1 和宪法 Principle III）
- 密码策略是否符合等保三级（8位+复杂度+90天有效期）
- 登录失败锁定是否生效（5次/15分钟）
- TLS 1.3 是否全站启用
- JWT RS256 签名是否正确实现
- 字段级加密（身份证号、手机号）是否生效
- 审计日志是否覆盖所有关键操作
- 敏感信息脱敏是否在 API 响应和日志中双重实施

### 性能基线（参考 PRD §10.2 表9-2）
- 产品列表页 P99 是否 ≤200ms
- 订单确认 P99 是否 ≤500ms
- 数据库查询 P99 是否 ≤100ms
- 缓存命中率是否达到预期

### 数据完整性（参考 PRD §12）
- 库存预扣/释放机制是否正确
- 订单状态机流转是否完整（参考 PRD 表6-5）
- 退款金额计算是否正确（参考 PRD §6.2.4 表6-6）
- 多租户数据隔离是否生效

### 前端完整性（与后端同等重要）
- Web 销售平台：首页、产品列表、产品详情、预订流程、支付页、个人中心、订单管理是否全部实现
- 微信小程序：登录、产品浏览、预订、支付、订单管理是否全部实现
- 后台管理系统：产品管理、订单管理、退款审核、退改规则配置、用户/角色/权限管理是否全部实现
- 前端页面是否包含加载状态、空状态、错误状态三种状态处理
- API 对接是否完整（每个后端 API 是否都有对应的前端调用）
- 表单校验是否完整（客户端校验 + 服务端错误提示展示）
- 支付流程是否端到端打通（前端调起支付 → 后端创建支付单 → 回调处理 → 前端状态更新）
- 响应式布局是否适配（Web 端 PC + 移动端，小程序端适配不同屏幕尺寸）
- 三端代码是否都有实际可运行的页面（不是空壳或占位符）

请生成结构化的检查清单文件。
```

---

## Phase 3：一期规格与实施

### 步骤 9：一期功能规格定义

**Skill**：`/speckit-specify`

**必读文档**：
- PRD §4.3 出境游跟团游（签证服务、护照管理、行前信息）
- PRD §5 支付系统（银联支付、定金+尾款模式）
- PRD §6 后台管理系统（供应商管理、财务管理、营销管理）
- PRD §7 供应商/开放平台（入驻、工作台、佣金结算）
- PRD §8 二级分销功能（分销关系、佣金管理、推广跟踪）
- PRD §11 外部接口（支付接口对比、消息推送）

**Prompt**：

```
/speckit-specify

一期功能规格定义：出境游 + 供应商开放平台 + 分销体系。

执行前请先阅读以下 PRD 章节（docs/travel_booking_system_v3.agent.final.md）：
- §4.3 出境游跟团游（签证服务交易闭环、护照管理、行前信息服务）
- §5 支付系统（§5.1 银联支付接入、§5.2.2 定金+尾款模式）
- §6 后台管理系统（§6.3 财务管理、§6.5 营销管理）
- §7 供应商/开放平台（§7.1 入驻、§7.2 工作台、§7.3 佣金结算）
- §8 二级分销功能（§8.1-§8.7 全部）
- §11 外部接口（§11.1 支付接口、§11.3 消息推送）

同时阅读 .specify/memory/constitution.md 和已完成的 specs/001-domestic-tour-mvp/spec.md（MVP 规格，确保一期与 MVP 的衔接）。

## 功能概述

一期在 MVP 基础上扩展三大核心能力：
1. 出境游业务线（签证服务交易闭环）
2. 供应商开放平台（入驻/审核/结算）
3. 二级分销体系

## 核心用户故事

### US1：出境游产品与预订
- 作为消费者，我可以浏览和筛选出境游产品（按国家/地区、签证类型）
- 作为消费者，我可以查看签证信息（类型、办理周期、材料清单）
- 作为消费者，我在预订时需要填写护照信息（系统校验有效期覆盖回程后6个月）
- 作为消费者，我可以选择签证代办服务
- 作为消费者，我可以跟踪签证办理进度
→ 详细流程参考 PRD §4.3.1-§4.3.5，签证闭环参考 PRD §4.3.4 表4-5

### US2：供应商入驻与管理
- 作为供应商，我可以在线提交入驻申请（企业信息、资质文件）
- 作为供应商，我可以在工作台发布和管理产品
- 作为供应商，我可以在工作台处理订单和退改申请
- 作为供应商，我可以查看结算单和申请提现
- 作为运营人员，我可以审核供应商入驻申请
- 作为运营人员，我可以配置佣金规则
→ 详细功能参考 PRD §7.1 入驻流程、§7.2 工作台（表7-1）、§7.3 佣金结算

### US3：二级分销
- 作为分销商，我可以申请入驻（个人/企业两种类型）
- 作为分销商，我可以为产品生成推广链接和二维码
- 作为一级分销商，我可以邀请二级分销商加入团队
- 作为分销商，我可以查看佣金明细和申请提现
- 作为运营人员，我可以管理分销商（审核/等级/状态）
→ 详细功能参考 PRD §8.2 分销商管理、§8.3 推广与跟踪、§8.4 佣金管理

### US4：支付扩展
- 作为消费者，我可以使用银联支付
- 作为消费者，我可以选择定金+尾款支付模式
- 作为消费者，我可以申请部分退款
→ 详细接口参考 PRD §5.1.3 银联支付、§5.2.2 定金+尾款模式、§5.3 退款管理

### US5：营销系统
- 作为消费者，我可以领取和使用优惠券
- 作为运营人员，我可以创建促销活动（限时特惠/满减/早鸟优惠）
→ 详细功能参考 PRD §6.5.2 优惠券管理（表6-9）、§6.5.3 促销活动管理

## 架构变更
- 订单服务和支付服务从单体拆分为独立部署单元
- 引入 NATS 消息队列处理异步事件
- 引入 Meilisearch 替代数据库全文搜索
- 新增抖音小程序

## 前端页面清单（一期新增）

### Web 销售平台新增
- 出境游产品列表/详情页（签证信息卡片、护照信息填写、签证进度跟踪）
- 优惠券领取/使用页面

### 微信小程序新增
- 出境游预订流程（护照信息填写、签证代办选择）
- 签证进度查询页

### 抖音小程序（新增端）
- 与微信小程序共享代码基，通过条件编译适配抖音 API
- 核心页面：登录、产品列表/详情、预订、订单管理

### 后台管理系统新增
- 供应商入驻审核页
- 供应商工作台（产品管理、订单处理、结算查看、提现申请）
- 财务管理页（支付流水、退款管理、供应商结算单、发票管理）
- 分销商管理页（分销商列表、审核、等级调整、佣金结算审核）
- 营销管理页（优惠券创建/发放、促销活动配置）
- 分销商中心（Web + 小程序：推广链接生成、佣金明细、团队管理、提现）

请基于以上信息生成功能规格文档。确保：
- 签证服务闭环（PRD §4.3.4 表4-5）的 15 个功能点全部覆盖
- 分销佣金计算规则（PRD §8.7.1）和防薅羊毛规则（PRD §8.7.2）明确写入 spec
- 供应商结算流程（PRD §7.3.2）的五步流程完整描述
- 每个用户故事必须同时定义前端页面和后端 API，禁止只定义 API 不定义页面
- 一期新增的所有前端平台（抖音小程序、供应商工作台、分销商中心）必须有完整的页面清单
```

---

### 步骤 10：一期实施计划

**Skill**：`/speckit-plan`

**必读文档**：
- `specs/002-phase1/spec.md`（一期规格）
- PRD §3 系统架构（微服务划分）
- PRD §5 支付系统（支付渠道技术参数对比表5-1）
- PRD §11 外部接口（支付接口详情、消息推送方案）
- PRD §12 数据库设计（新增实体关系）

**Prompt**：

```
/speckit-plan

请为一期（出境游+供应商开放平台+分销体系）制定实施计划。

执行前请先阅读以下文档：
1. specs/002-phase1/spec.md（一期功能规格）
2. .specify/memory/constitution.md（项目宪法）
3. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §3 系统架构（§3.1.2 微服务划分）
   - §5 支付系统（§5.1 支付渠道接入，表5-1 三大支付渠道技术参数对比）
   - §11 外部接口（§11.1 支付接口详情、§11.3 消息推送接口）
   - §12 数据库设计（新增供应商、分销、签证相关实体）

## 架构变更

### 服务拆分
- 从 Gin 单体拆分为独立服务：
  - user-service（用户服务）
  - product-service（产品服务）
  - order-service（订单服务）
  - payment-service（支付服务）
  - distribution-service（分销服务，新增）
- 各服务通过 NATS 消息队列异步通信
- Traefik 作为 API 网关统一入口

### 新增组件
- Meilisearch：产品搜索索引
- NATS：订单状态变更通知、支付回调异步处理、短信推送
- Consul：服务注册与发现

### 数据库变更
- 新增分销相关表：distributor、distributor_relation、commission_record、withdrawal_record
- 新增供应商相关表：supplier、supplier_qualification、settlement_statement、commission_rule
- 新增签证相关表：visa_order、visa_material、visa_progress

### 新增支付渠道
- 银联支付（网关支付 + WAP 支付）
- 定金+尾款支付模式

### 新增前端
- 抖音小程序（Uni-App 条件编译）
- 供应商工作台（Vue 3 + Element Plus）
- 分销商中心（Web + 小程序）

请生成详细的实施计划。
```

---

### 步骤 11：一期任务拆解

**Skill**：`/speckit-tasks`

**必读文档**：
- `specs/002-phase1/spec.md`（一期规格）
- `specs/002-phase1/plan.md`（一期实施计划）
- `specs/002-phase1/data-model.md`（一期数据模型，新增供应商/分销/签证表）
- `specs/002-phase1/contracts/`（一期 API 契约）
- `specs/002-phase1/research.md`（一期技术决策）
- PRD §7, §8（确认供应商和分销功能点的完整性）

**Prompt**：

```
/speckit-tasks

请根据一期的 spec.md、plan.md 和相关设计文档生成可执行的任务列表。

执行前请先阅读以下文档：
1. specs/002-phase1/spec.md（一期功能规格）
2. specs/002-phase1/plan.md（一期实施计划）
3. specs/002-phase1/data-model.md（一期数据模型）
4. specs/002-phase1/contracts/（一期 API 契约）
5. specs/002-phase1/research.md（一期技术决策）
6. .specify/memory/constitution.md（项目宪法）
7. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §7 供应商/开放平台（确认入驻/工作台/结算功能点完整性）
   - §8 二级分销功能（确认分销商管理/推广/佣金功能点完整性）

## 阶段结构建议

- Phase 1：服务拆分基础设施（NATS、Consul、服务间通信框架）
- Phase 2：出境游产品与预订（US1）
- Phase 3：供应商入驻与管理（US2）
- Phase 4：二级分销体系（US3）
- Phase 5：支付扩展 - 银联+定金尾款（US4）
- Phase 6：营销系统（US5）
- Phase 7：Meilisearch 搜索集成
- Phase 8：抖音小程序
- Phase 9：供应商工作台前端
- Phase 10：分销商中心前端
- Phase 11：集成测试与安全加固（MFA、WAF、异地备份）
- Phase 12：部署与监控（Prometheus + Grafana + Jaeger）

请生成完整的 tasks.md。
```

---

### 步骤 12：一期跨文档一致性分析

**Skill**：`/speckit-analyze`

**必读文档**：
- `specs/002-phase1/spec.md`, `plan.md`, `tasks.md`
- `specs/002-phase1/data-model.md`（一期数据模型）
- `specs/002-phase1/contracts/`（一期 API 契约）
- PRD §4.3, §7, §8（对照功能覆盖度）
- PRD §11（对照外部接口集成完整性）

**Prompt**：

```
/speckit-analyze

请对一期的 spec.md、plan.md、tasks.md 进行跨文档一致性分析。

执行前请先阅读以下文档：
1. specs/002-phase1/spec.md, plan.md, tasks.md
2. specs/002-phase1/data-model.md（一期数据模型）
3. specs/002-phase1/contracts/（一期 API 契约）
4. .specify/memory/constitution.md（项目宪法）
5. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §4.3 出境游跟团游（签证闭环 15 个功能点是否覆盖）
   - §7 供应商/开放平台（入驻/工作台/结算是否完整）
   - §8 二级分销（分销商管理/推广/佣金是否完整）
   - §11 外部接口（支付接口、消息推送集成是否到位）

## 分析重点

1. **PRD 覆盖度**：spec.md 是否遗漏了 PRD §4.3/§7/§8 中的关键功能点
2. **需求覆盖**：所有功能需求是否在 tasks.md 中有对应任务
3. **架构变更验证**：服务拆分方案是否在 tasks 中有对应的基础设施任务
4. **新增组件集成**：NATS、Meilisearch、Consul 的集成任务是否完整
5. **数据库迁移**：新增表和字段是否有对应的迁移任务
6. **前端覆盖**：供应商工作台、分销商中心、抖音小程序的任务是否完整
7. **安全增强**：MFA、WAF、异地备份等安全加固任务是否到位

请输出结构化分析报告。
```

---

### 步骤 13：一期代码实现

**Skill**：`/speckit-implement`

**必读文档**：
- `specs/002-phase1/plan.md`, `tasks.md`
- `specs/002-phase1/data-model.md`（一期数据模型，新增供应商/分销/签证表结构）
- `specs/002-phase1/contracts/`（一期 API 契约）
- `specs/002-phase1/quickstart.md`（一期验证场景）
- PRD §5 支付系统（银联支付 SDK 用法、定金+尾款逻辑）
- PRD §11 外部接口（支付接口参数、消息推送模板）

**Prompt**：

```
/speckit-implement

请按照一期 tasks.md 中的任务列表逐步实现。

执行前请先阅读以下文档：
1. specs/002-phase1/plan.md（实施计划）
2. specs/002-phase1/tasks.md（任务列表）
3. specs/002-phase1/data-model.md（一期数据模型）
4. specs/002-phase1/contracts/（一期 API 契约）
5. specs/002-phase1/quickstart.md（一期验证场景）
6. .specify/memory/constitution.md（编码规范）
7. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §5.1.3 银联支付接入（SDK 用法和回调处理）
   - §5.2.2 定金+尾款模式（两笔独立支付方案）
   - §11.1 外部接口（支付接口参数详情）
   - §11.3 消息推送接口（短信/小程序订阅消息模板）

## 实现要求

1. 服务拆分时保持 API 兼容性，Web 前端和小程序无需改动
2. NATS 消息格式使用 JSON，定义统一的事件信封结构
3. 分销佣金计算必须原子性，防止并发场景下金额错误
4. 供应商数据隔离通过 tenant_id + supplier_id 双重过滤
5. 银联支付接入使用 smartwalle/unionpay SDK
6. 所有新服务必须注册到 Consul 并暴露 /health 和 /ready 端点
7. **前端实现要求**：
   - 抖音小程序必须与微信小程序共享业务逻辑层，仅在平台 API 层做条件编译
   - 供应商工作台必须实现完整的 RBAC 权限控制（菜单+按钮级）
   - 分销商中心必须实现推广链接生成、二维码下载、佣金明细展示
   - 签证进度跟踪页面必须展示五节点状态机（待提交→审核中→已送签→已出签/已拒签）

请开始实现。
```

---

## Phase 4：二期规格与实施

### 步骤 14：二期功能规格定义

**Skill**：`/speckit-specify`

**必读文档**：
- PRD §4.4 邮轮游（搜索、航次选择、舱房预订、值船流程）
- PRD §6 后台管理系统（邮轮产品管理、数据报表）
- PRD §9 多租户管理（租户生命周期、数据隔离、品牌定制）
- PRD §10 非功能需求（可用性、可维护性）

**Prompt**：

```
/speckit-specify

二期功能规格定义：邮轮游 + 数据分析 + 多租户 + 微服务完整化。

执行前请先阅读以下 PRD 章节（docs/travel_booking_system_v3.agent.final.md）：
- §4.4 邮轮游（§4.4.1 搜索、§4.4.2 航次选择与舱房预订、§4.4.3 业务规则）
- §6 后台管理系统（§6.1.3-§6.1.7 邮轮产品管理、§6.6 数据报表）
- §9 多租户管理（§9.1 租户生命周期、§9.2 数据隔离、§9.3 品牌定制）
- §10 非功能需求（§10.3 可用性、§10.4 可维护性）

同时阅读 .specify/memory/constitution.md 和已完成的 MVP/一期 spec（确保衔接）。

## 功能概述

二期聚焦三大能力：
1. 邮轮游完整业务线（航次/舱房/值船/船票）
2. 数据分析与报表系统
3. 多租户管理与品牌定制

## 核心用户故事

### US1：邮轮产品管理
- 作为运营人员，我可以管理邮轮公司和船只信息
- 作为运营人员，我可以创建和管理航次（航线/停靠港口/日期）
- 作为运营人员，我可以管理舱房类型和库存
- 作为运营人员，我可以配置值船流程说明文档
- 作为运营人员，我可以上传和分发船票
→ 详细功能参考 PRD §6.1.3-§6.1.7（邮轮产品发布、基础信息、设施、舱房、航次管理）

### US2：邮轮产品浏览与预订
- 作为消费者，我可以按航线区域、出发港口、日期搜索邮轮产品
- 作为消费者，我可以查看船只详情和设施介绍
- 作为消费者，我可以对比不同舱房类型（面积/景观/价格）
- 作为消费者，我可以预订舱房并填写出游人信息（含护照）
- 作为消费者，我可以查看值船指南和下载船票
→ 详细功能参考 PRD §4.4.1 搜索（表4-7）、§4.4.2 航次选择与舱房预订（表4-8）

### US3：数据分析
- 作为运营人员，我可以查看销售数据看板（订单量/GMV/退款率）
- 作为运营人员，我可以查看转化率漏斗（曝光→点击→下单→支付）
- 作为运营人员，我可以进行 RFM 用户分析
- 作为财务人员，我可以查看收入日报/月报和毛利分析
→ 详细功能参考 PRD §6.6 数据报表（表6-10）

### US4：多租户管理
- 作为系统管理员，我可以创建和管理租户
- 作为租户管理员，我可以配置品牌展示（Logo/主题色/首页布局）
- 作为租户管理员，我可以独立配置支付渠道参数
→ 详细功能参考 PRD §9.1 租户生命周期、§9.2 数据隔离（表8-1、表8-2）、§9.3 品牌定制

### US5：营销增强
- 作为运营人员，我可以配置首页 Banner 和推荐位
- 作为运营人员，我可以创建专题活动页（可视化编辑器）
→ 详细功能参考 PRD §6.5.1 首页/专题配置

## 前端页面清单（二期新增）

### Web 销售平台新增
- 邮轮搜索页（航线区域筛选、出发港口、日期范围、邮轮公司）
- 邮轮航次详情页（航线地图、船只参数、设施介绍、舱房类型对比）
- 邮轮预订流程（选航次→选舱房→填出游人→附加服务→确认支付）
- 值船指南页、船票下载页

### 微信小程序新增
- 邮轮搜索/详情/预订（小程序适配版本）

### 后台管理系统新增
- 邮轮公司/船只管理页
- 航次管理页（航线编辑、停靠港口、库存矩阵）
- 舱房类型管理页（四层类型配置、房号批量导入）
- 值船流程文档管理页、船票上传/分发页
- 数据分析看板（ECharts：销售看板、转化漏斗、RFM 分析、财务报表）
- 多租户管理页（租户创建/配置、品牌定制、支付配置、资源配额）
- 首页配置页（Banner 管理、推荐位配置、专题活动页可视化编辑器）

## 技术目标
- 微服务完整化：7 个服务独立部署
- 全链路监控：Prometheus + Grafana + Jaeger
- 日志聚合：Loki + Promtail
- 等保三级正式测评通过

请基于以上信息生成功能规格文档。确保：
- 邮轮舱房按人计价逻辑（PRD §4.4.2 第3/4人优惠）明确描述
- 值船流程说明文档/船票上传分发/提醒通知三个节点完整覆盖
- 多租户 RLS 策略和资源配额（PRD §9.2）明确写入 spec
```

---

### 步骤 15：二期实施计划

**Skill**：`/speckit-plan`

**必读文档**：
- `specs/003-phase2/spec.md`（二期规格）
- PRD §3 系统架构（微服务完整化方案）
- PRD §9 多租户（数据隔离技术方案）
- PRD §12 数据库设计（邮轮域实体、RLS 策略）

**Prompt**：

```
/speckit-plan

请为二期（邮轮游+数据分析+多租户+微服务完整化）制定实施计划。

执行前请先阅读以下文档：
1. specs/003-phase2/spec.md（二期功能规格）
2. .specify/memory/constitution.md（项目宪法）
3. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §3 系统架构（微服务划分和部署架构）
   - §9.2 数据隔离方案（RLS 策略、资源配额）
   - §12 数据库设计（§12.1.4 邮轮域实体、§12.3 分区与分表策略）

## 架构目标

### 微服务完整化
- 7 个服务独立部署：user-service、product-service、order-service、payment-service、cruise-service、marketing-service、file-service
- 各服务拥有独立数据库 Schema
- 通过 Consul 实现服务注册与发现
- Traefik 按路径前缀路由

### 邮轮数据模型
→ 参考 PRD §12.1.4 邮轮域实体（邮轮公司→船只→舱房类型→航次→停靠港口→舱房库存）
- cruise_line → cruise_ship → cabin_category（四层：内舱/海景/阳台/套房）
- sailing → port_of_call（停靠港口序列）
- sailing × cabin_category → cabin_inventory（二维库存矩阵）
- 同舱第 3/4 人折扣定价逻辑

### 数据分析
- 核心指标准实时（5-15 分钟延迟）
- 明细报表 T+1 更新
- Apache ECharts 可视化

### 多租户
→ 参考 PRD §9.2.1 数据隔离方案（表8-1）
- 共享数据库 + PostgreSQL RLS
- Redis Key 前缀隔离：tenant:{id}:{entity}:{pk}
- Meilisearch 索引隔离：products_tenant_{id}
- 大客户可迁移至独立数据库实例

请生成详细的实施计划。
```

---

### 步骤 16：二期任务拆解

**Skill**：`/speckit-tasks`

**必读文档**：
- `specs/003-phase2/spec.md`, `plan.md`
- `specs/003-phase2/data-model.md`（二期数据模型，邮轮域五层实体）
- `specs/003-phase2/contracts/`（二期 API 契约）
- `specs/003-phase2/research.md`（二期技术决策）
- PRD §4.4（邮轮功能点完整性）
- PRD §9（多租户功能点完整性）

**Prompt**：

```
/speckit-tasks

请根据二期的 spec.md、plan.md 和相关设计文档生成可执行的任务列表。

执行前请先阅读以下文档：
1. specs/003-phase2/spec.md（二期功能规格）
2. specs/003-phase2/plan.md（二期实施计划）
3. specs/003-phase2/data-model.md（二期数据模型）
4. specs/003-phase2/contracts/（二期 API 契约）
5. specs/003-phase2/research.md（二期技术决策）
6. .specify/memory/constitution.md（项目宪法）
4. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §4.4 邮轮游（确认搜索/预订/值船功能点完整性）
   - §9 多租户管理（确认租户生命周期/数据隔离功能点完整性）

## 阶段结构建议

- Phase 1：邮轮基础数据管理（邮轮公司/船只/舱房类型/设施）
- Phase 2：航次管理（航线/停靠港口/库存矩阵）
- Phase 3：邮轮 C 端搜索与预订（US1+US2）
- Phase 4：值船流程与船票管理
- Phase 5：邮轮退改规则体系
- Phase 6：数据分析与报表（US3）
- Phase 7：多租户管理（US4）
- Phase 8：营销增强 - Banner/推荐位/专题页（US5）
- Phase 9：支付宝小程序
- Phase 10：微服务完整化（服务拆分部署）
- Phase 11：监控体系（Prometheus + Grafana + Jaeger + Loki）
- Phase 12：等保三级正式测评准备
- Phase 13：性能压测与优化

请生成完整的 tasks.md。
```

---

### 步骤 17：二期跨文档一致性分析

**Skill**：`/speckit-analyze`

**必读文档**：
- `specs/003-phase2/spec.md`, `plan.md`, `tasks.md`
- `specs/003-phase2/data-model.md`（二期数据模型）
- `specs/003-phase2/contracts/`（二期 API 契约）
- PRD §4.4, §9, §10（对照功能覆盖度和非功能需求）

**Prompt**：

```
/speckit-analyze

请对二期的 spec.md、plan.md、tasks.md 进行跨文档一致性分析。

执行前请先阅读以下文档：
1. specs/003-phase2/spec.md, plan.md, tasks.md
2. specs/003-phase2/data-model.md（二期数据模型）
3. specs/003-phase2/contracts/（二期 API 契约）
4. .specify/memory/constitution.md（项目宪法）
3. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §4.4 邮轮游（邮轮五层数据模型、值船流程、退改规则是否完整覆盖）
   - §9 多租户管理（RLS 策略、资源配额、品牌定制是否完整覆盖）
   - §10 非功能需求（等保三级正式测评准备是否到位）

## 分析重点

1. **邮轮数据模型**：五层嵌套关系（公司→船只→航次→舱房→库存）是否在 tasks 中完整覆盖
2. **值船流程**：文档上传、船票分发、提醒通知等任务是否完整
3. **数据分析**：物化视图、预聚合、ECharts 可视化任务是否到位
4. **多租户**：RLS 策略、资源配额、品牌定制任务是否覆盖
5. **微服务拆分**：7 个服务的独立部署和 Consul 注册任务是否完整
6. **监控体系**：Prometheus + Grafana + Jaeger + Loki 四层监控任务是否到位
7. **等保合规**：正式测评准备任务是否覆盖所有控制点（参考 PRD 表9-1）

请输出结构化分析报告。
```

---

### 步骤 18：二期代码实现

**Skill**：`/speckit-implement`

**必读文档**：
- `specs/003-phase2/plan.md`, `tasks.md`
- `specs/003-phase2/data-model.md`（二期数据模型，邮轮域五层实体+多租户 RLS）
- `specs/003-phase2/contracts/`（二期 API 契约）
- `specs/003-phase2/quickstart.md`（二期验证场景）
- PRD §9（多租户 RLS 实现细节）
- PRD §10（等保合规实现细节）
- PRD §12（邮轮域表结构和索引）

**Prompt**：

```
/speckit-implement

请按照二期 tasks.md 中的任务列表逐步实现。

执行前请先阅读以下文档：
1. specs/003-phase2/plan.md（实施计划）
2. specs/003-phase2/tasks.md（任务列表）
3. specs/003-phase2/data-model.md（二期数据模型）
4. specs/003-phase2/contracts/（二期 API 契约）
5. specs/003-phase2/quickstart.md（二期验证场景）
6. .specify/memory/constitution.md（编码规范）
7. docs/travel_booking_system_v3.agent.final.md 的以下章节：
   - §9.2 数据隔离方案（RLS 策略实现、资源配额校验）
   - §10 非功能需求（等保三级合规实现细节，表9-1 对照表）
   - §12 数据库设计（邮轮域表结构、分区策略、多租户索引设计）

## 实现要求

1. 邮轮服务作为独立服务部署，拥有独立数据库 Schema
2. 舱房库存使用行级锁 + 乐观锁双重防护
3. 数据分析报表使用物化视图预聚合，避免实时查询压力
4. 多租户 RLS 策略必须在所有业务表上启用
5. 租户配置变更通过 Consul KV 热刷新，无需重启服务
6. 监控告警覆盖四层：基础设施、中间件、应用、业务
7. **前端实现要求**：
   - 邮轮搜索/预订页面必须适配 Web 和小程序双端
   - 数据分析看板必须使用 ECharts 实现可视化（折线图、柱状图、饼图、漏斗图）
   - 多租户品牌定制必须支持运行时动态切换主题色和 Logo
   - 专题活动页编辑器必须支持拖拽组件和实时预览

请开始实现。
```

---

## 贯穿全阶段的持续步骤

### 持续 A：更新 Agent 上下文

**Skill**：`/speckit-agent-context-update`

**触发时机**：每次完成 `/speckit-plan` 后执行一次

**Prompt**：

```
/speckit-agent-context-update

请更新 CLAUDE.md 中的 Spec Kit 管理区域，使其指向最新的 plan.md 文件路径。
```

---

### 持续 B：收敛检查

**Skill**：`/speckit-converge`

**触发时机**：每个阶段（MVP/一期/二期）的 `/speckit-implement` 完成后执行

**Prompt**：

```
/speckit-converge

请评估当前代码实现与 spec.md、plan.md、tasks.md 的对齐程度。

执行前请先阅读以下文档：
1. specs/<当前阶段>/spec.md, plan.md, tasks.md
2. .specify/memory/constitution.md（项目宪法）
3. docs/travel_booking_system_v3.agent.final.md 的对应章节（按阶段参照目录表）

## 检查范围

1. 所有功能需求（FR-###）是否已实现
2. 所有成功标准（SC-###）是否已满足
3. 所有用户故事的验收场景是否已覆盖
4. 实施计划中的技术决策是否已落地
5. 项目宪法中的 MUST 原则是否无违反
6. PRD 对应章节的关键业务规则是否无遗漏

## 期望输出

- 如果存在未实现的需求：将剩余工作作为新任务追加到 tasks.md
- 如果全部已实现：报告"已收敛"并建议进入下一阶段或提交 PR
```

---

## 附录：执行顺序总览

```
Phase 0（项目初始化）
  ├── 步骤 1: /speckit-constitution          ← 建立项目宪法（读 PRD 概览）
  └── 步骤 2: /speckit-clarify               ← 澄清需求冲突（读 PRD 全文）

Phase 1（MVP 规格定义）
  ├── 步骤 3: /speckit-specify               ← MVP 功能规格（读 PRD §4.1, §4.2, §6, §10, §12）
  └── 步骤 4: /speckit-plan                  ← MVP 实施计划（读 PRD §3, §5, §10, §12）
       └── 持续 A: /speckit-agent-context-update

Phase 2（MVP 实施）
  ├── 步骤 5: /speckit-tasks                 ← 任务拆解（读 PRD §12）
  ├── 步骤 6: /speckit-analyze               ← 跨文档分析（读 PRD §4.1, §4.2, §10）
  ├── 步骤 7: /speckit-implement             ← 代码实现（读 PRD §5, §10, §12）
  ├── 步骤 8: /speckit-checklist             ← 验收检查（读 PRD §4.1, §4.2, §10）
  └── 持续 B: /speckit-converge              ← 收敛检查

Phase 3（一期规格与实施）
  ├── 步骤 9:  /speckit-specify              ← 一期功能规格（读 PRD §4.3, §5-§8, §11）
  ├── 步骤 10: /speckit-plan                 ← 一期实施计划（读 PRD §3, §5, §11, §12）
  │    └── 持续 A: /speckit-agent-context-update
  ├── 步骤 11: /speckit-tasks                ← 一期任务拆解（读 PRD §7, §8）
  ├── 步骤 12: /speckit-analyze              ← 跨文档分析（读 PRD §4.3, §7, §8, §11）
  ├── 步骤 13: /speckit-implement            ← 一期代码实现（读 PRD §5, §11）
  └── 持续 B: /speckit-converge

Phase 4（二期规格与实施）
  ├── 步骤 14: /speckit-specify              ← 二期功能规格（读 PRD §4.4, §6, §9, §10）
  ├── 步骤 15: /speckit-plan                 ← 二期实施计划（读 PRD §3, §9, §12）
  │    └── 持续 A: /speckit-agent-context-update
  ├── 步骤 16: /speckit-tasks                ← 二期任务拆解（读 PRD §4.4, §9）
  ├── 步骤 17: /speckit-analyze              ← 跨文档分析（读 PRD §4.4, §9, §10）
  ├── 步骤 18: /speckit-implement            ← 二期代码实现（读 PRD §9, §10, §12）
  └── 持续 B: /speckit-converge
```
