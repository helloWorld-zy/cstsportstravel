-- Migration 004: Distribution Tables
-- Creates distributor, distributor_relation, promotion_link, commission_detail,
-- withdrawal_record, promotion_click tables for the two-level distribution system.

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════
-- Distributor Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS distributor (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    user_id             BIGINT NOT NULL UNIQUE,
    distributor_no      VARCHAR(32) NOT NULL UNIQUE,
    distributor_type    VARCHAR(10) NOT NULL CHECK (distributor_type IN ('personal', 'enterprise')),
    level               INT NOT NULL DEFAULT 1 CHECK (level IN (1, 2)),
    grade               VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (grade IN ('normal', 'senior')),
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'frozen', 'cancelled', 'deactivated')),
    real_name           VARCHAR(50),
    id_card_number      VARCHAR(255),  -- AES-256-GCM encrypted
    id_card_front_url   VARCHAR(500),
    id_card_back_url    VARCHAR(500),
    enterprise_name     VARCHAR(200),
    credit_code         VARCHAR(18),
    business_license_url VARCHAR(500),
    bank_name           VARCHAR(100),
    bank_account_name   VARCHAR(100),
    bank_account_number VARCHAR(255),  -- AES-256-GCM encrypted
    phone               VARCHAR(20) NOT NULL,
    email               VARCHAR(100),
    promotion_channel   TEXT,
    invite_code         VARCHAR(10) UNIQUE,
    agreement_signed_at TIMESTAMPTZ,
    agreement_signed_ip VARCHAR(45),
    grade_valid_until   TIMESTAMPTZ,
    frozen_reason       TEXT,
    frozen_until        TIMESTAMPTZ,
    total_commission    DECIMAL(15,2) NOT NULL DEFAULT 0,
    withdrawable_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    frozen_amount       DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distributor_user ON distributor (user_id);
CREATE INDEX IF NOT EXISTS idx_distributor_status ON distributor (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_distributor_invite_code ON distributor (invite_code);

-- ═══════════════════════════════════════════════════════════════════════════
-- Distributor Relation Table (parent-child hierarchy)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS distributor_relation (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    distributor_id  BIGINT NOT NULL UNIQUE REFERENCES distributor(id),
    parent_id       BIGINT REFERENCES distributor(id),
    level           INT NOT NULL CHECK (level IN (1, 2)),
    bind_time       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'dissolved')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distributor_rel_parent ON distributor_relation (parent_id);
CREATE INDEX IF NOT EXISTS idx_distributor_rel_distributor ON distributor_relation (distributor_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- Promotion Link Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS promotion_link (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    distributor_id  BIGINT NOT NULL REFERENCES distributor(id),
    product_id      BIGINT NOT NULL,
    short_link      VARCHAR(100) NOT NULL UNIQUE,
    qr_code_url     VARCHAR(500),
    click_pv        BIGINT NOT NULL DEFAULT 0,
    click_uv        BIGINT NOT NULL DEFAULT 0,
    order_count     BIGINT NOT NULL DEFAULT 0,
    order_amount    DECIMAL(15,2) NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_promo_link_distributor ON promotion_link (distributor_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_promo_link_product ON promotion_link (distributor_id, product_id);
CREATE INDEX IF NOT EXISTS idx_promo_link_short ON promotion_link (short_link);

-- ═══════════════════════════════════════════════════════════════════════════
-- Commission Detail Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS commission_detail (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    order_id            BIGINT NOT NULL,
    distributor_id      BIGINT NOT NULL REFERENCES distributor(id),
    commission_level    INT NOT NULL CHECK (commission_level IN (1, 2)),
    order_actual_amount DECIMAL(12,2) NOT NULL,
    commission_rate     DECIMAL(5,2) NOT NULL,
    commission_amount   DECIMAL(12,2) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'frozen', 'withdrawable', 'withdrawn', 'recovered')),
    frozen_until        TIMESTAMPTZ,
    settled_at          TIMESTAMPTZ,
    withdrawn_at        TIMESTAMPTZ,
    recovered_amount    DECIMAL(12,2),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_commission_distributor ON commission_detail (distributor_id, status);
CREATE INDEX IF NOT EXISTS idx_commission_order ON commission_detail (order_id);
CREATE INDEX IF NOT EXISTS idx_commission_frozen ON commission_detail (status, frozen_until);

-- ═══════════════════════════════════════════════════════════════════════════
-- Withdrawal Record Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS withdrawal_record (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    withdrawal_no       VARCHAR(32) NOT NULL UNIQUE,
    distributor_id      BIGINT NOT NULL REFERENCES distributor(id),
    amount              DECIMAL(12,2) NOT NULL,
    bank_name           VARCHAR(100) NOT NULL,
    bank_account_name   VARCHAR(100) NOT NULL,
    bank_account_number VARCHAR(255) NOT NULL,  -- AES-256-GCM encrypted
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'paid')),
    reviewed_by         BIGINT,
    reviewed_at         TIMESTAMPTZ,
    reject_reason       TEXT,
    paid_at             TIMESTAMPTZ,
    payment_voucher_url VARCHAR(500),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawal_distributor ON withdrawal_record (distributor_id, status);
CREATE INDEX IF NOT EXISTS idx_withdrawal_status ON withdrawal_record (tenant_id, status);

-- ═══════════════════════════════════════════════════════════════════════════
-- Promotion Click Record Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS promotion_click (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    promotion_link_id   BIGINT NOT NULL REFERENCES promotion_link(id),
    distributor_id      BIGINT NOT NULL REFERENCES distributor(id),
    visitor_id          VARCHAR(64),
    ip_address          VARCHAR(45) NOT NULL,
    user_agent          VARCHAR(500),
    device_fingerprint  VARCHAR(64),
    source              VARCHAR(20) NOT NULL CHECK (source IN ('link', 'qrcode')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_click_link ON promotion_click (promotion_link_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_click_ip ON promotion_click (ip_address, created_at DESC);

COMMIT;
