<!-- SPECKIT START -->
## Project Constitution
Constitution: `.specify/memory/constitution.md` (v1.0.0)

## Current Plan
Plan: `specs/001-domestic-tour-mvp/plan.md`
Feature: `specs/001-domestic-tour-mvp/spec.md`
Data Model: `specs/001-domestic-tour-mvp/data-model.md`
API Contracts: `specs/001-domestic-tour-mvp/contracts/`
Tasks: `specs/001-domestic-tour-mvp/tasks.md` (164 tasks)
Quickstart: `specs/001-domestic-tour-mvp/quickstart.md`

## Key Constraints (from Constitution)
- Backend: Go 1.26+, Gin, GORM v2 + pgx
- Database: PostgreSQL 17.x (dev)
- Security: TLS 1.3, AES-256-GCM field encryption, JWT RS256, RBAC, audit logs ≥6 months
- Deployment: Windows Server 2022, WinSW services
- Delivery: MVP (domestic tour) → Phase 2 (outbound + supplier) → Phase 3 (cruise + analytics)
<!-- SPECKIT END -->
