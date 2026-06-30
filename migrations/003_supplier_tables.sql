-- Migration 003: Supplier Tables
-- Creates supplier, supplier_qualification, settlement_statement, commission_rule tables
-- for the supplier open platform.

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════
-- Supplier Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS supplier (
    id                          BIGSERIAL PRIMARY KEY,
    tenant_id                   BIGINT NOT NULL,
    supplier_no                 VARCHAR(32) NOT NULL UNIQUE,
    company_name                VARCHAR(200) NOT NULL,
    unified_social_credit_code  VARCHAR(18) NOT NULL UNIQUE,
    registered_address          VARCHAR(500) NOT NULL,
    registered_capital          DECIMAL(15,2),
    establishment_date          DATE,
    business_license_url        VARCHAR(500) NOT NULL,
    legal_person_name           VARCHAR(50) NOT NULL,
    legal_person_id_card        VARCHAR(255) NOT NULL,  -- AES-256-GCM encrypted
    business_scope              VARCHAR(500) NOT NULL,
    travel_license_no           VARCHAR(50),
    travel_license_url          VARCHAR(500),
    contact_name                VARCHAR(50) NOT NULL,
    contact_phone               VARCHAR(20) NOT NULL,
    contact_email               VARCHAR(100),
    finance_contact_name        VARCHAR(50),
    finance_contact_phone       VARCHAR(20),
    bank_name                   VARCHAR(100),
    bank_account_name           VARCHAR(100),
    bank_account_number         VARCHAR(255),  -- AES-256-GCM encrypted
    commission_rate             DECIMAL(5,2),
    settlement_cycle            VARCHAR(10) NOT NULL DEFAULT 'monthly' CHECK (settlement_cycle IN ('daily', 'weekly', 'monthly')),
    settlement_day              INT,
    rating_score                DECIMAL(3,1),
    status                      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewing', 'active', 'suspended', 'terminated')),
    application_no              VARCHAR(32) NOT NULL,
    applied_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at                 TIMESTAMPTZ,
    contract_signed_at          TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_supplier_status ON supplier (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_supplier_credit_code ON supplier (unified_social_credit_code);

-- ═══════════════════════════════════════════════════════════════════════════
-- Supplier Qualification Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS supplier_qualification (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    supplier_id         BIGINT NOT NULL REFERENCES supplier(id),
    qualification_type  VARCHAR(30) NOT NULL CHECK (qualification_type IN ('business_license', 'travel_license', 'id_card_front', 'id_card_back', 'other')),
    file_url            VARCHAR(500) NOT NULL,
    file_name           VARCHAR(200) NOT NULL,
    expiry_date         DATE,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    review_comment      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_supplier_qual_supplier ON supplier_qualification (supplier_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- Settlement Statement Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS settlement_statement (
    id                      BIGSERIAL PRIMARY KEY,
    tenant_id               BIGINT NOT NULL,
    settlement_no           VARCHAR(32) NOT NULL UNIQUE,
    supplier_id             BIGINT NOT NULL REFERENCES supplier(id),
    period_start            DATE NOT NULL,
    period_end              DATE NOT NULL,
    order_count             INT NOT NULL DEFAULT 0,
    order_total_amount      DECIMAL(15,2) NOT NULL DEFAULT 0,
    refund_amount           DECIMAL(15,2) NOT NULL DEFAULT 0,
    platform_commission     DECIMAL(15,2) NOT NULL DEFAULT 0,
    refund_commission_deduct DECIMAL(15,2) NOT NULL DEFAULT 0,
    payable_amount          DECIMAL(15,2) NOT NULL DEFAULT 0,
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'disputed', 'paid')),
    supplier_confirmed_at   TIMESTAMPTZ,
    dispute_reason          TEXT,
    approved_by             BIGINT,
    approved_at             TIMESTAMPTZ,
    paid_at                 TIMESTAMPTZ,
    payment_voucher_url     VARCHAR(500),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_settlement_supplier ON settlement_statement (tenant_id, supplier_id, period_start DESC);
CREATE INDEX IF NOT EXISTS idx_settlement_status ON settlement_statement (tenant_id, status);

-- ═══════════════════════════════════════════════════════════════════════════
-- Commission Rule Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS commission_rule (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    rule_name       VARCHAR(100) NOT NULL,
    scope_type      VARCHAR(20) NOT NULL CHECK (scope_type IN ('global', 'category', 'supplier', 'product')),
    scope_id        BIGINT,
    commission_rate DECIMAL(5,2) NOT NULL,
    priority        INT NOT NULL DEFAULT 0,
    effective_from  TIMESTAMPTZ NOT NULL,
    effective_to    TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_commission_rule_scope ON commission_rule (tenant_id, scope_type, scope_id, priority DESC);

COMMIT;
