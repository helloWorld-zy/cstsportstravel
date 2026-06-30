-- Migration 005: Visa Tables
-- Creates visa_order, visa_material, visa_progress tables
-- for the visa service workflow (5-node state machine).

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════
-- Visa Order Table
-- State machine: pending_submit → reviewing → submitted → approved/rejected
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS visa_order (
    id                          BIGSERIAL PRIMARY KEY,
    tenant_id                   BIGINT NOT NULL,
    visa_order_no               VARCHAR(32) NOT NULL UNIQUE,
    main_order_id               BIGINT NOT NULL,
    user_id                     BIGINT NOT NULL,
    country_id                  BIGINT NOT NULL,
    visa_type                   VARCHAR(50) NOT NULL,
    status                      VARCHAR(20) NOT NULL DEFAULT 'pending_submit' CHECK (status IN ('pending_submit', 'reviewing', 'submitted', 'approved', 'rejected')),
    submitted_at                TIMESTAMPTZ,
    reviewed_at                 TIMESTAMPTZ,
    approved_at                 TIMESTAMPTZ,
    rejected_at                 TIMESTAMPTZ,
    reject_reason               TEXT,
    estimated_completion_date   DATE,
    visa_fee                    DECIMAL(10,2),
    tracking_company            VARCHAR(50),
    tracking_number             VARCHAR(50),
    visa_expiry_date            DATE,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visa_order_main ON visa_order (main_order_id);
CREATE INDEX IF NOT EXISTS idx_visa_order_user ON visa_order (user_id, status);
CREATE INDEX IF NOT EXISTS idx_visa_order_status ON visa_order (tenant_id, status);

-- ═══════════════════════════════════════════════════════════════════════════
-- Visa Material Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS visa_material (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    visa_order_id   BIGINT NOT NULL REFERENCES visa_order(id),
    material_type   VARCHAR(50) NOT NULL,
    material_name   VARCHAR(100) NOT NULL,
    file_url        VARCHAR(500),
    file_size       BIGINT,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'submitted', 'approved', 'rejected', 'supplement')),
    review_comment  TEXT,
    reviewed_by     BIGINT,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visa_material_order ON visa_material (visa_order_id, status);

-- ═══════════════════════════════════════════════════════════════════════════
-- Visa Progress Table (audit trail for status changes)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS visa_progress (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    visa_order_id   BIGINT NOT NULL REFERENCES visa_order(id),
    from_status     VARCHAR(20),
    to_status       VARCHAR(20) NOT NULL,
    operator_id     BIGINT,
    operator_type   VARCHAR(20) NOT NULL CHECK (operator_type IN ('system', 'admin', 'supplier')),
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visa_progress_order ON visa_progress (visa_order_id, created_at DESC);

COMMIT;
