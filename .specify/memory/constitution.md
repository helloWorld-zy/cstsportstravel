<!--
  Sync Impact Report
  Version change: 0.0.0 (template) → 1.0.0 (initial adoption)
  Added sections:
    - Core Principles (5 principles)
    - Technology Stack Constraints
    - Deployment & Operations Constraints
    - Governance
  Removed sections: None
  Templates requiring updates:
    - .specify/templates/plan-template.md ✅ compatible (Constitution Check section exists)
    - .specify/templates/spec-template.md ✅ compatible (no constitution-specific references)
    - .specify/templates/tasks-template.md ✅ compatible (no constitution-specific references)
  Follow-up TODOs: None
-->

# Travel Booking OTA Platform Constitution

## Core Principles

### I. API-First Design

All system capabilities MUST be exposed through well-defined API contracts before any implementation begins.

- Every service endpoint MUST have a corresponding OpenAPI 3.0 specification
- API contracts MUST be authored and reviewed BEFORE writing handler or service code
- Frontend (Web, Mini Program, Admin) and backend teams MUST agree on API contracts before parallel development begins
- All API responses MUST follow a unified envelope format: `{code, message, data, trace_id}`
- Breaking API changes MUST increment the API version (e.g., `/api/v2/`) and maintain backward compatibility for at least one release cycle

**Rationale**: API-first ensures parallel team workflows, reduces integration friction, and produces machine-readable documentation that serves as the single source of truth for system behavior.

### II. Domain-Driven Service Boundaries

System architecture MUST follow Domain-Driven Design principles, organizing code around business domains rather than technical layers.

- Each bounded context (User, Product, Order, Payment, Cruise, Marketing, Platform Governance) MUST be implemented as an independent module with clear domain boundaries
- Cross-domain communication MUST use explicit interfaces (HTTP/gRPC for sync, NATS for async); direct database access across domain boundaries is PROHIBITED
- Each domain module MUST own its data model; shared tables across domains are PROHIBITED unless mediated through a well-defined shared kernel
- Service decomposition follows progressive delivery: monolith-first for MVP, extract services at Phase 1/2 when domain boundaries stabilize

**Rationale**: DDD boundaries prevent monolithic coupling, enable independent deployment and scaling, and align code structure with business capabilities for long-term maintainability.

### III. Security-by-Design (NON-NEGOTIABLE)

Security controls MUST be embedded into the system from day one, not retrofitted. All requirements in this principle are mandatory for every release.

- ALL data in transit MUST use TLS 1.3; HTTP (non-HTTPS) access is PROHIBITED in any environment
- Sensitive fields (national ID, passport number, phone number, bank card) MUST be encrypted at rest using AES-256-GCM with keys managed by a KMS
- User passwords MUST be hashed with Argon2id; plaintext or reversible password storage is PROHIBITED
- API authentication MUST use JWT with RS256 asymmetric signing; Access Token lifetime MUST NOT exceed 15 minutes
- All administrative operations MUST be protected by RBAC (function permission + data permission + field permission)
- Sensitive administrative operations (refund approval, permission changes, data export) MUST require MFA verification
- ALL user-facing and administrative operations MUST produce audit log entries; audit logs MUST be retained for at least 6 months with tamper-proof storage
- Security event logs (failed logins, privilege escalation attempts, anomaly detection) MUST be permanently retained

**Rationale**: The system handles personal identity documents and financial transactions, placing it under regulatory scrutiny (等保三级, Personal Information Protection Law). Security cannot be deferred.

### IV. Progressive Delivery

Features MUST be delivered incrementally through well-defined phases, each producing a deployable and independently valuable increment.

- **MVP** (Phase 1): Domestic group tour complete transaction loop — product browsing, booking, payment (Alipay + WeChat Pay), order management, basic admin
- **Phase 2**: Outbound tour + visa service loop, supplier open platform, two-level distribution system, UnionPay payment
- **Phase 3**: Cruise tour, data analytics dashboard, multi-tenant management, full microservice decomposition
- Each phase MUST define measurable success criteria before work begins
- Each phase MUST be independently deployable without breaking functionality delivered in prior phases
- Architectural decisions MUST defer complexity: start with a monolith, extract services only when domain boundaries and load patterns justify it

**Rationale**: Progressive delivery reduces risk by validating business assumptions early, limits scope creep, and ensures the system generates value from the first release.

### V. Code Quality Discipline

All code MUST adhere to established quality standards enforced through automated tooling.

- Go code MUST pass `golangci-lint` with zero errors before merge; the linter configuration MUST be version-controlled
- Frontend code (TypeScript) MUST pass ESLint + Prettier checks with zero warnings
- All commits MUST follow Conventional Commits format: `type(scope): description`
- Unit test coverage for core business logic (payment calculation, order state machine, inventory management, commission calculation) MUST be ≥70%
- Critical transaction paths (order creation → payment → confirmation) MUST have integration test coverage
- Every exported Go function MUST have a godoc comment
- Every API endpoint MUST have a corresponding OpenAPI operation description

**Rationale**: Automated quality gates prevent regression, enforce consistency across team members, and ensure the codebase remains maintainable as it grows.

## Technology Stack Constraints

The following technology choices are FIXED for this project. Deviations require explicit constitution amendment.

| Layer | Technology | Version | Constraint |
|-------|-----------|---------|------------|
| Backend Language | Go | 1.26+ | All server-side code MUST be Go |
| Web Framework | Gin | latest stable | All HTTP handlers MUST use Gin |
| ORM | GORM v2 | v2.x | CRUD operations; raw SQL permitted for complex queries via pgx |
| Database Driver | pgx | v5.x | Primary PostgreSQL driver |
| Primary Database | PostgreSQL | 17.x (dev), 18+ (prod target) | All persistent data MUST use PostgreSQL |
| Cache / Session | Redis / Memurai | 7.2+ | Session, distributed lock, hot data cache |
| Search Engine | Meilisearch | 1.19+ | Product search index (Phase 2+) |
| Message Queue | NATS | 2.11+ | Async events, service decoupling |
| Task Queue | Asynq | latest | Delayed/scheduled tasks |
| API Gateway | Traefik | 3.x+ | SSL termination, routing, rate limiting |
| Service Discovery | Consul | 1.22+ | Service registration, health check, KV config |
| Frontend Web | Nuxt.js 3 | 3.x+ | SSR for SEO-critical pages |
| Frontend Admin | Vue 3 + Element Plus | 3.x+ | SPA for internal operations |
| Frontend Mini Program | Uni-App (Vue 3) | 3.x+ | Cross-platform WeChat/Douyin/Alipay mini programs |
| CI/CD | GitHub Actions | - | Windows Runner for build and test |
| Service Wrapper | WinSW | - | Register Go binaries as Windows services |
| Monitoring | Prometheus + Grafana + Jaeger | - | Four-layer observability |
| Logging | Zap + Lumberjack | - | Structured JSON logs, auto-rotation |

**Rationale**: These choices have been validated for Windows Server native deployment compatibility. Components that do not support native Windows (e.g., APISIX, Apache Pulsar server) are explicitly excluded.

## Deployment & Operations Constraints

- ALL backend services MUST run on Windows Server 2022/2025 as native Windows services via WinSW
- Go binaries MUST be compiled with `CGO_ENABLED=0` for static linking; runtime dependencies are PROHIBITED
- Database backup MUST follow the schedule: daily full backup (`pg_basebackup`) + incremental backup every 15 minutes (WAL archival)
- RTO (Recovery Time Objective) for core services MUST be < 5 minutes
- RPO (Recovery Point Objective) for database MUST be < 1 minute
- Production deployments MUST use rolling update strategy with zero-downtime; health check (`/ready`) MUST pass before traffic is routed
- All configuration (database credentials, API keys, encryption keys) MUST be stored in environment variables or Consul KV; hardcoded secrets in source code are PROHIBITED

## Governance

This constitution is the supreme governing document for the Travel Booking OTA Platform project. It supersedes all other development practices, coding guidelines, and architectural decisions when conflicts arise.

**Amendment Procedure**:
1. Propose the amendment with rationale in writing
2. Review impact on existing code, templates, and in-progress features
3. Obtain approval from project lead
4. Update constitution version following semantic versioning rules
5. Propagate changes to dependent templates and documentation

**Compliance Review**:
- Every pull request MUST be reviewed against applicable constitution principles
- Constitution violations in code MUST be flagged and resolved before merge
- The `/speckit-analyze` command MUST be run after task generation to verify cross-artifact consistency with constitution principles

**Versioning Policy**:
- MAJOR: Backward-incompatible principle removal or redefinition
- MINOR: New principle added or materially expanded guidance
- PATCH: Clarifications, wording fixes, non-semantic refinements

**Version**: 1.0.0 | **Ratified**: 2026-06-19 | **Last Amended**: 2026-06-19
