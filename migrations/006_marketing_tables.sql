-- Migration 006: Marketing Tables
-- Creates coupon, coupon_claim, promotion_activity tables
-- for the marketing system (coupons + promotions).

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════
-- Coupon Table
-- Types: full_reduction, discount, cash, exchange
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS coupon (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    coupon_name         VARCHAR(100) NOT NULL,
    coupon_type         VARCHAR(20) NOT NULL CHECK (coupon_type IN ('full_reduction', 'discount', 'cash', 'exchange')),
    discount_amount     DECIMAL(10,2),
    discount_rate       DECIMAL(5,2),
    discount_cap        DECIMAL(10,2),
    min_consumption     DECIMAL(10,2),
    total_stock         INT NOT NULL,
    claimed_count       INT NOT NULL DEFAULT 0,
    used_count          INT NOT NULL DEFAULT 0,
    per_user_limit      INT NOT NULL DEFAULT 1,
    per_device_limit    INT,
    validity_type       VARCHAR(20) NOT NULL CHECK (validity_type IN ('fixed', 'relative')),
    valid_from          TIMESTAMPTZ,
    valid_to            TIMESTAMPTZ,
    valid_days          INT,
    applicable_scope    VARCHAR(20) NOT NULL DEFAULT 'all' CHECK (applicable_scope IN ('all', 'category', 'product')),
    applicable_ids      BIGINT[],
    applicable_channels VARCHAR(50)[],
    stackable           BOOLEAN NOT NULL DEFAULT false,
    stackable_types     VARCHAR(20)[],
    exchange_product_id BIGINT,
    status              VARCHAR(20) NOT NULL DEFAULT 'not_started' CHECK (status IN ('not_started', 'active', 'expired', 'exhausted')),
    created_by          BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coupon_status ON coupon (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_coupon_type ON coupon (tenant_id, coupon_type);

-- ═══════════════════════════════════════════════════════════════════════════
-- Coupon Claim Record Table
-- Status flow: available → occupied → used / expired / returned / voided
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS coupon_claim (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    coupon_id   BIGINT NOT NULL REFERENCES coupon(id),
    user_id     BIGINT NOT NULL,
    device_id   VARCHAR(64),
    status      VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'occupied', 'used', 'expired', 'returned', 'voided')),
    order_id    BIGINT,
    claimed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at     TIMESTAMPTZ,
    expired_at  TIMESTAMPTZ,
    returned_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coupon_claim_user ON coupon_claim (user_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_coupon_claim_coupon ON coupon_claim (coupon_id, user_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- Promotion Activity Table
-- Types: flash_sale, full_reduction, early_bird
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS promotion_activity (
    id                      BIGSERIAL PRIMARY KEY,
    tenant_id               BIGINT NOT NULL,
    activity_name           VARCHAR(200) NOT NULL,
    activity_type           VARCHAR(20) NOT NULL CHECK (activity_type IN ('flash_sale', 'full_reduction', 'early_bird')),
    start_time              TIMESTAMPTZ NOT NULL,
    end_time                TIMESTAMPTZ NOT NULL,
    applicable_products     BIGINT[],
    applicable_categories   BIGINT[],
    rules                   JSONB NOT NULL,
    activity_stock          INT,
    per_user_limit          INT,
    stackable_with_coupon   BOOLEAN NOT NULL DEFAULT false,
    status                  VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'ended', 'cancelled')),
    created_by              BIGINT NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_promotion_activity_status ON promotion_activity (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_promotion_activity_time ON promotion_activity (start_time, end_time);

COMMIT;
