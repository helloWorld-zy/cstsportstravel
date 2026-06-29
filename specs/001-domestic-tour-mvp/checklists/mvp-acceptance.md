# MVP 交付验收检查清单

**Purpose**: 验证境内跟团游 MVP 规格、契约与文档的需求质量——完整性、清晰度、一致性、可测量性与场景覆盖。

**Created**: 2026-06-28

**Scope**: 功能完整性 · 安全合规 · 性能基线 · 数据完整性 · 前端完整性

---

## 1. 功能完整性 — 用户体系

- [ ] CHK001 - 手机号+短信验证码注册流程的需求是否完整覆盖：验证码位数、有效期、重发间隔、手机号格式校验规则？ [Completeness, Spec §FR-001]
- [ ] CHK002 - 微信授权登录需求是否明确区分了三种模式（JSAPI/wx.login/扫码）的差异行为和绑定流程？ [Clarity, Spec §FR-002]
- [ ] CHK003 - 实名认证需求是否定义了认证失败的处理规则：重试次数限制（3次/天）、认证状态流转（unverified/pending/verified/rejected）、驳回原因反馈？ [Completeness, Spec §FR-003, Edge Case]
- [ ] CHK004 - 常用出游人管理需求是否明确了编辑已关联历史订单的出游人时的快照保留机制？ [Clarity, Spec §FR-004, PRD §4.1.2]
- [ ] CHK005 - 密码策略需求是否量化了全部参数：最小8位、复杂度组合规则、90天有效期、到期前14天提醒、过期后强制修改？ [Measurability, Spec §FR-005, Constitution §III]
- [ ] CHK006 - 登录失败锁定需求是否明确：5次失败/15分钟锁定窗口、锁定后解锁方式（自动/手动）、管理员重置流程？ [Clarity, Spec §FR-006]
- [ ] CHK007 - 用户体系需求是否定义了微信登录用户与手机号注册用户的账号合并/绑定规则及冲突处理？ [Coverage, Edge Case, Spec §FR-002]

## 2. 功能完整性 — 产品模块

- [ ] CHK008 - 产品发布流程需求是否完整覆盖五步骤（基础信息→行程编辑→价格配置→退改规则→库存设置）的必填项和校验规则？ [Completeness, Spec §FR-007]
- [ ] CHK009 - 产品审核流程需求是否明确了"关键字段变更"的具体字段清单及触发重新审核的条件？ [Clarity, Spec §FR-008]
- [ ] CHK010 - 团期日历管理需求是否定义了按日设置的价格类型（成人/儿童/婴儿/单房差）和库存数量的上下限约束？ [Completeness, Spec §FR-009]
- [ ] CHK011 - 产品列表筛选需求是否覆盖了PRD表4-1中全部16个功能点（F-I-L01至F-I-L16）？ [Coverage, PRD §4.2.1]
- [ ] CHK012 - 产品详情页需求是否覆盖了PRD表4-2中全部19个功能点（F-I-D01至F-I-D19），特别是退改政策"不可折叠隐藏"的交互约束？ [Coverage, PRD §4.2.2]
- [ ] CHK013 - 搜索联想推荐需求是否明确了联想结果的优先级排序规则（热门目的地→产品名称→景点）和响应时间要求？ [Clarity, Spec §FR-012, PRD §4.1.3]
- [ ] CHK014 - 批量调价需求是否明确区分了五种模式（固定价格/百分比/固定金额/公式/跟随）的计算规则和适用场景？ [Clarity, Spec §FR-013]
- [ ] CHK015 - 产品状态机需求是否完整定义了5个状态（draft/pending_review/approved/suspended/change_pending_review）之间的全部合法流转路径？ [Consistency, Data-Model §Product Status Machine]

## 3. 功能完整性 — 预订与订单

- [ ] CHK016 - 四步预订流程需求是否覆盖了PRD表4-3中全部17个功能点（F-I-B01至F-I-B17）？ [Coverage, PRD §4.2.3]
- [ ] CHK017 - 单房差自动附加规则是否明确了边界条件：成人数为0（仅儿童/婴儿）时不计算、成人数为奇数时附加1份？ [Clarity, Spec §FR-015, Edge Case]
- [ ] CHK018 - 30分钟支付倒计时需求是否明确：倒计时起点（订单创建时刻）、展示格式（分钟:秒）、超时后自动取消+库存释放的原子性保证？ [Clarity, Spec §FR-016]
- [ ] CHK019 - 订单状态机需求是否完整定义了9个内部状态的全部合法流转路径及其触发条件（参照PRD表6-5和Data-Model状态机）？ [Completeness, Spec §FR-017, Data-Model §Order State Machine]
- [ ] CHK020 - 订单状态映射需求是否明确定义了9个内部状态到6个用户可见Tab的映射关系，特别是"待出行"合并paid_full和pending_travel的规则？ [Clarity, Spec §FR-017]
- [ ] CHK021 - 退款流程需求是否覆盖了退改阶梯费率计算、退款金额公式、原路退回机制和退款到账时间承诺？ [Completeness, Spec §FR-018, FR-019, FR-020]
- [ ] CHK022 - 退款分级审批需求是否明确定义了三级阈值（≤1000/1000-5000/>5000元）对应的审批角色和审批流程？ [Clarity, Spec §FR-021]
- [ ] CHK023 - 儿童/婴儿关联成人规则是否明确了每位儿童必须关联至少一位成人、每成人最多携带1名婴儿的约束校验时机？ [Coverage, Spec §User Story 3, Edge Case]
- [ ] CHK024 - 并发下单场景的需求是否定义了库存预扣机制、超时释放规则和不超售的业务保证方式？ [Coverage, Edge Case, Spec §Assumptions]
- [ ] CHK025 - 支付超时后用户仍完成支付的场景，需求是否定义了拒绝确认并触发退款的回退流程？ [Coverage, Edge Case]

## 4. 功能完整性 — 支付模块

- [ ] CHK026 - 支付方式需求是否明确了MVP支持的支付渠道（支付宝PC/H5、微信Native/JSAPI/小程序）及各渠道的调起方式？ [Completeness, Spec §FR-022, FR-023]
- [ ] CHK027 - 支付回调幂等性需求是否明确定义了重复回调的判定逻辑和去重机制？ [Clarity, Spec §FR-024]
- [ ] CHK028 - 支付渠道回调丢失场景的需求是否定义了主动查询支付状态的触发条件、频率和超时处理？ [Coverage, Spec §FR-025, Edge Case]
- [ ] CHK029 - 支付接口契约是否定义了统一的请求签名验证机制（HMAC-SHA256）和时间戳有效期（±5分钟）？ [Completeness, Constitution §III, Contract §payment-api]

## 5. 功能完整性 — 后台管理

- [ ] CHK030 - 后台产品管理API契约是否覆盖了产品CRUD、审核（提交/通过/驳回/下架）、团期管理和批量调价的全部端点？ [Completeness, Contract §admin-api]
- [ ] CHK031 - 后台订单管理需求是否定义了订单查询的全部筛选维度（订单号/手机号/状态/日期范围/产品类型/供应商）？ [Completeness, Spec §User Story 6]
- [ ] CHK032 - 退改规则配置需求是否明确定义了阶梯费率的数据模型（距出发天数区间→退款比例）和模板复用机制？ [Clarity, Spec §FR-020, Data-Model §refund_rule]
- [ ] CHK033 - RBAC权限模型需求是否明确定义了三个维度（功能权限/数据权限/字段权限）的控制粒度和实现方式？ [Completeness, Spec §FR-026, Constitution §III]
- [ ] CHK034 - 供应商数据隔离需求是否明确定义了隔离机制（supplier_id字段过滤）和跨供应商数据访问的禁止规则？ [Clarity, Spec §FR-027]
- [ ] CHK035 - MFA多因素认证需求是否明确了适用场景（退款审批>1000元、权限变更、数据导出）和验证方式（TOTP+短信）？ [Completeness, Spec §FR-030, Constitution §III]
- [ ] CHK036 - 管理员账号创建需求是否明确了初始密码生成规则和首次登录强制修改密码的流程？ [Clarity, Spec §User Story 7]

## 6. 安全合规 — 认证与授权

- [ ] CHK037 - 密码策略需求是否与等保三级要求（GB/T 22239-2019）逐项对齐：≥8位、复杂度、90天更换周期、到期提醒？ [Consistency, Spec §FR-005, PRD §10.1.5 #2]
- [ ] CHK038 - 登录失败锁定需求是否与等保三级要求对齐：5次失败/15分钟自动锁定？ [Consistency, Spec §FR-006, PRD §10.1.5 #3]
- [ ] CHK039 - JWT RS256非对称签名需求是否明确定义了Access Token有效期（15分钟）、Refresh Token有效期（7天）和Token黑名单机制？ [Completeness, Constitution §III, PRD §10.1.1]
- [ ] CHK040 - RBAC权限模型需求是否覆盖了等保三级要求的最小权限、权限分离、敏感操作二次授权、默认账户管理和控制粒度五项测评要求？ [Coverage, PRD §10.1.5 #7-#11]
- [ ] CHK041 - 管理员敏感操作MFA需求是否与等保三级"重要操作需双因素"要求一致？ [Consistency, Spec §FR-030, PRD §10.1.5 #5]

## 7. 安全合规 — 数据安全

- [ ] CHK042 - 字段级加密需求是否明确了加密算法（AES-256-GCM）、加密字段清单（身份证号、手机号、护照号、银行卡号）和密钥管理方式（KMS）？ [Completeness, Constitution §III, PRD §10.1.2]
- [ ] CHK043 - 传输层安全需求是否明确定义了全站TLS 1.3强制启用、HSTS头部配置和证书自动续期？ [Completeness, Constitution §III, PRD §10.1.2]
- [ ] CHK044 - 敏感信息脱敏需求是否明确定义了各字段的脱敏规则（手机号前3后4、身份证前6后4等）和双重实施位置（API响应+日志）？ [Clarity, PRD §10.1.2]
- [ ] CHK045 - 审计日志需求是否覆盖了等保三级要求的全部审计范围：登录事件、操作事件、敏感操作、安全事件？ [Coverage, Spec §FR-028, PRD §10.1.4, PRD §10.1.5 #12-#16]
- [ ] CHK046 - 审计日志留存需求是否与等保三级要求对齐：≥6个月（实际定义为2年），防篡改保护机制？ [Consistency, Spec §FR-028, PRD §10.1.5 #15]
- [ ] CHK047 - 密码存储需求是否明确定义了Argon2id单向哈希算法，禁止明文或可逆存储？ [Completeness, Constitution §III]

## 8. 安全合规 — 攻击防护

- [ ] CHK048 - SQL注入防护需求是否明确定义了参数化查询、输入白名单校验和最小权限数据库账号的组合策略？ [Completeness, PRD §10.1.3]
- [ ] CHK049 - API请求签名需求是否明确定义了HMAC-SHA256签名算法、时间戳有效期（±5分钟）和nonce唯一性校验？ [Completeness, PRD §10.1.2]
- [ ] CHK050 - 限流需求是否明确定义了五个维度的限流规则：全局QPS、单IP、单用户、接口级、服务间调用？ [Completeness, PRD §10.3.2]

## 9. 性能基线

- [ ] CHK051 - 产品列表页响应时间需求是否量化为P99 ≤200ms（通用API）或P99 ≤1.0s（含Meilisearch命中）？ [Measurability, PRD §10.2.1 表9-2]
- [ ] CHK052 - 订单确认页响应时间需求是否量化为P99 ≤500ms？ [Measurability, Spec §SC-003, PRD §10.2.1 表9-2]
- [ ] CHK053 - 数据库查询响应时间需求是否量化为P99 ≤100ms，并明确了索引覆盖策略？ [Measurability, PRD §10.2.1 表9-2, PRD §10.2.3]
- [ ] CHK054 - 缓存策略需求是否明确定义了五级缓存架构（浏览器→CDN→本地→Redis→数据库）和8类数据的缓存位置、过期策略、更新模式？ [Completeness, PRD §10.2.4 表9-3]
- [ ] CHK055 - 并发能力需求是否量化了基线指标：峰值同时在线10,000人、全站QPS≥10,000、订单TPS≥500？ [Measurability, PRD §10.2.1 表9-2]
- [ ] CHK056 - 性能需求是否定义了扩展目标（5倍）和扩展方式（水平扩容+数据库分片），确保架构可扩展？ [Coverage, PRD §10.2.1]

## 10. 数据完整性 — 库存与状态机

- [ ] CHK057 - 库存预扣/释放需求是否明确定义了预扣时机（下单时）、释放时机（超时取消/退款完成）和库存计算公式（available = total - sold - locked）？ [Clarity, Data-Model §departure_date]
- [ ] CHK058 - 订单状态机需求是否定义了全部11条合法状态转换路径及其触发条件（参照Data-Model状态机）？ [Completeness, Data-Model §Order State Machine]
- [ ] CHK059 - 退款金额计算需求是否明确定义了退改阶梯费率的计算公式和各阶梯的退款比例？ [Clarity, Spec §FR-020, PRD §4.2.5]
- [ ] CHK060 - 多租户数据隔离需求是否明确定义了供应商数据隔离机制和行级安全策略（RLS）？ [Completeness, Spec §FR-027, PRD §1.2.2]
- [ ] CHK061 - 支付回调幂等性需求是否在数据模型层面定义了唯一约束（order_id, channel, attempt_no）防止重复支付记录？ [Consistency, Data-Model §payment_transaction]

## 11. 前端完整性 — Web 销售平台

- [ ] CHK062 - Web销售平台需求是否完整定义了首页的四个核心区域：搜索框、金刚区导航、Banner轮播、内容推荐区？ [Completeness, PRD §4.1.4]
- [ ] CHK063 - 产品列表页需求是否定义了双模式切换（列表/网格）的交互规则和每种模式的信息展示密度？ [Clarity, PRD §4.2.1]
- [ ] CHK064 - 产品详情页需求是否覆盖了PRD表4-2中全部19个功能点，包括退改政策"不可折叠隐藏"的交互约束？ [Coverage, PRD §4.2.2]
- [ ] CHK065 - 预订流程四步向导需求是否定义了步骤进度条、底部实时价格汇总栏和返回修改前序步骤的交互规则？ [Completeness, PRD §4.2.3]
- [ ] CHK066 - 支付页面需求是否定义了支付宝/微信支付选择、30分钟倒计时展示（分钟:秒格式）和支付成功/失败/超时的状态展示？ [Completeness, Spec §User Story 3]
- [ ] CHK067 - 个人中心需求是否定义了全部功能菜单分组：账号管理、出行管理、订单管理、服务组？ [Completeness, PRD §4.1.2]
- [ ] CHK068 - 订单管理页面需求是否定义了6种状态Tab、订单卡片信息展示和不同状态下的操作按钮差异？ [Completeness, Spec §User Story 4, PRD §4.2.4]

## 12. 前端完整性 — 微信小程序

- [ ] CHK069 - 微信小程序需求是否定义了登录流程（wx.login + 手机号绑定）和与Web端账号体系打通的规则？ [Completeness, Spec §FR-002, Spec §Assumptions]
- [ ] CHK070 - 小程序产品浏览需求是否定义了与Web端共享核心业务逻辑但适配小程序交互的差异点？ [Consistency, Spec §User Story 2]
- [ ] CHK071 - 小程序支付需求是否明确定义了wx.requestPayment调起微信支付的流程和参数？ [Completeness, Spec §User Story 3]
- [ ] CHK072 - 小程序订单管理需求是否定义了与Web端功能一致但适配移动端交互的规则？ [Consistency, Spec §User Story 4]

## 13. 前端完整性 — 后台管理系统

- [ ] CHK073 - 后台产品管理页面需求是否覆盖了多步骤表单（基础信息→行程编辑→价格配置→退改规则→库存设置）的全部字段和校验规则？ [Completeness, Spec §User Story 5]
- [ ] CHK074 - 行程编辑器需求是否定义了按天数自动生成行程卡片、拖拽排序和行程模板复用的交互规则？ [Clarity, Spec §User Story 5]
- [ ] CHK075 - 价格日历页面需求是否定义了月历网格视图、每日价格/库存展示和批量调价五种模式的交互？ [Completeness, Spec §User Story 5]
- [ ] CHK076 - 退改规则配置页面需求是否定义了阶梯费率编辑器和模板保存/复用机制？ [Completeness, Spec §User Story 6]
- [ ] CHK077 - 后台订单管理页面需求是否定义了多维度筛选、订单详情展示和退款审核操作的完整流程？ [Completeness, Spec §User Story 6]
- [ ] CHK078 - 用户/角色/权限管理页面需求是否定义了用户列表、角色配置和权限树（菜单+按钮）的交互？ [Completeness, Spec §User Story 7]

## 14. 前端完整性 — 通用交互质量

- [ ] CHK079 - 前端页面需求是否为所有异步数据加载定义了三种状态处理：加载中（loading）、空状态（empty）、错误状态（error）？ [Coverage, Gap]
- [ ] CHK080 - 前端表单校验需求是否定义了客户端实时校验规则和服务端错误提示的展示方式？ [Completeness, Gap]
- [ ] CHK081 - 支付流程前端需求是否定义了端到端的完整链路：调起支付→创建支付单→回调处理→状态更新→结果展示？ [Completeness, Gap]
- [ ] CHK082 - 响应式布局需求是否明确定义了Web端PC和移动端的断点规则和适配策略？ [Clarity, Gap]
- [ ] CHK083 - 三端（Web/小程序/后台）需求是否定义了实际可运行页面的验收标准，排除空壳或占位符？ [Measurability, Spec §SC-009]

## 15. API 契约完整性

- [ ] CHK084 - 用户API契约是否覆盖了全部8个端点：短信验证码、登录、微信授权、管理员登录、用户信息、实名认证、常用出游人CRUD？ [Completeness, Contract §user-api]
- [ ] CHK085 - 产品API契约是否覆盖了全部6个端点：产品列表、产品详情、团期日历、行程信息、评价列表、搜索联想？ [Completeness, Contract §product-api]
- [ ] CHK086 - 订单API契约是否覆盖了全部5个端点：创建订单、订单详情、取消订单、申请退款、确认收货？ [Completeness, Contract §order-api]
- [ ] CHK087 - 支付API契约是否覆盖了全部6个端点：创建支付、支付状态、支付宝回调、微信回调、主动查询、测试模拟？ [Completeness, Contract §payment-api]
- [ ] CHK088 - 管理API契约是否覆盖了全部17个端点：产品CRUD、审核、团期管理、批量调价、订单管理、退款审批、用户管理、角色管理、菜单管理、退改规则？ [Completeness, Contract §admin-api]
- [ ] CHK089 - 所有API响应格式是否统一遵循信封格式 `{code, message, data, trace_id}`？ [Consistency, Constitution §I]
- [ ] CHK090 - API契约是否为所有端点定义了错误响应格式和错误码体系？ [Completeness, Gap]

## 16. 需求质量 — 一致性检查

- [ ] CHK091 - Spec中FR-017定义的订单状态值（snake_case）是否与Data-Model中main_order.order_status字段的枚举值完全一致？ [Consistency, Spec §FR-017, Data-Model §main_order]
- [ ] CHK092 - Spec中FR-021定义的退款审批阈值（1000/5000元）是否与PRD §4.2.5和Spec Clarifications中的决策一致？ [Consistency, Spec §FR-021]
- [ ] CHK093 - Spec中支付倒计时（30分钟）是否与PRD §4.2.3和Spec Clarifications中的刻意决策一致，且明确了覆盖PRD的15分钟选项？ [Consistency, Spec §Clarifications]
- [ ] CHK094 - Data-Model中amount字段的单位约定（分/cents）是否与Quickstart中的API示例（yuan with 2 decimal places）的转换规则一致？ [Consistency, Data-Model §Amount Convention, Quickstart §Important Notes]
- [ ] CHK095 - Constitution中要求的JWT RS256签名和15分钟有效期是否在Spec和API契约中一致体现？ [Consistency, Constitution §III, Spec §FR-001]
- [ ] CHK096 - PRD §10.1.5等保三级对照表的32项测评要求是否在Spec的FR和Data-Model中逐项有对应实现？ [Traceability, PRD §10.1.5]

## 17. 需求质量 — 边界与异常场景

- [ ] CHK097 - 并发库存扣减场景的需求是否定义了技术方案（Redis原子操作 vs 数据库行锁）和不超售的验证方式？ [Clarity, Spec §Assumptions, Edge Case]
- [ ] CHK098 - 支付回调延迟/丢失场景的需求是否定义了主动查询的触发条件、查询频率和最终超时处理？ [Coverage, Spec §FR-025, Edge Case]
- [ ] CHK099 - 退款到账户失败场景的需求是否定义了异常状态记录和运营人员人工处理流程？ [Coverage, Edge Case]
- [ ] CHK100 - 团期售罄后用户仍在预订流程中的需求是否定义了确认订单时的拦截提示和引导规则？ [Coverage, Edge Case]
- [ ] CHK101 - 身份证号校验需求是否明确定义了18位校验码规则（ISO 7064:1983.MOD 11-2）的实现标准？ [Clarity, Edge Case]
- [ ] CHK102 - 产品审核驳回后重新提交的需求是否定义了审核历史记录和状态回退规则？ [Coverage, Edge Case]

## 18. 需求质量 — 依赖与假设

- [ ] CHK103 - Spec Assumptions中声明的"供应商由平台代为录入"是否在后台管理需求中明确体现，排除了供应商自助入驻功能？ [Consistency, Spec §Assumptions]
- [ ] CHK104 - Spec Assumptions中声明的"全额支付为唯一支付模式"是否在支付需求中明确排除了定金+尾款模式？ [Consistency, Spec §Assumptions]
- [ ] CHK105 - Spec Assumptions中声明的"附加服务为简单商品"是否在预订流程需求中明确了不对接保险公司API的简化处理？ [Consistency, Spec §Assumptions]
- [ ] CHK106 - Spec Assumptions中声明的"评价系统为简化版本"是否在产品详情页需求中明确了自动发布无需审核的规则？ [Consistency, Spec §Assumptions]
- [ ] CHK107 - 快速启动文档中定义的验证场景（VS1-VS8）是否覆盖了Spec中全部P0级用户故事的关键验收条件？ [Coverage, Quickstart, Spec]

## 19. 需求质量 — 可追溯性

- [ ] CHK108 - Spec中每个FR编号是否都有对应的API契约端点、Data-Model实体和Quickstart验证场景？ [Traceability]
- [ ] CHK109 - PRD §10.1.5等保三级对照表的32项是否每项都有对应的系统实现描述和验证方式？ [Traceability, PRD §10.1.5]
- [ ] CHK110 - 需求规格是否建立了统一的ID体系（FR编号、功能编号、CHK编号）支持端到端追溯？ [Traceability, Gap]
