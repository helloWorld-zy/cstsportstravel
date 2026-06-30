-- Migration 007: Payment Extension Columns
-- Extends main_order and payment_transaction tables for:
-- - Deposit + balance payment mode
-- - Distribution tracking (distributor_id_l1/l2, promotion_code)
-- - Marketing (coupon_claim_id, coupon_discount, activity_id, activity_discount)
-- - UnionPay integration

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════
-- main_order table extensions
-- ═══════════════════════════════════════════════════════════════════════════

-- Payment mode (full payment vs deposit + balance)
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS payment_mode VARCHAR(20) DEFAULT 'full' CHECK (payment_mode IN ('full', 'deposit'));
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS deposit_amount DECIMAL(12,2);
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS balance_amount DECIMAL(12,2);
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS balance_deadline TIMESTAMPTZ;
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS deposit_paid_at TIMESTAMPTZ;
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS balance_paid_at TIMESTAMPTZ;

-- Distribution tracking
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS distributor_id_l1 BIGINT;
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS distributor_id_l2 BIGINT;
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS promotion_code VARCHAR(20);

-- Marketing (coupon + promotion)
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS coupon_claim_id BIGINT;
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS coupon_discount DECIMAL(10,2);
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS activity_id BIGINT;
ALTER TABLE main_order ADD COLUMN IF NOT EXISTS activity_discount DECIMAL(10,2);

-- Indexes for distribution queries
CREATE INDEX IF NOT EXISTS idx_order_distributor_l1 ON main_order (distributor_id_l1) WHERE distributor_id_l1 IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_order_distributor_l2 ON main_order (distributor_id_l2) WHERE distributor_id_l2 IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_order_promotion_code ON main_order (promotion_code) WHERE promotion_code IS NOT NULL;

-- ═══════════════════════════════════════════════════════════════════════════
-- payment_transaction table extensions
-- ═══════════════════════════════════════════════════════════════════════════

-- Payment type (deposit, balance, full, refund)
ALTER TABLE payment_transaction ADD COLUMN IF NOT EXISTS payment_type VARCHAR(20) DEFAULT 'full' CHECK (payment_type IN ('deposit', 'balance', 'full', 'refund'));

-- UnionPay integration fields
ALTER TABLE payment_transaction ADD COLUMN IF NOT EXISTS unionpay_trade_no VARCHAR(64);
ALTER TABLE payment_transaction ADD COLUMN IF NOT EXISTS unionpay_query_id VARCHAR(64);

-- Indexes for payment type queries
CREATE INDEX IF NOT EXISTS idx_payment_type ON payment_transaction (payment_type) WHERE payment_type != 'full';
CREATE INDEX IF NOT EXISTS idx_payment_unionpay ON payment_transaction (unionpay_trade_no) WHERE unionpay_trade_no IS NOT NULL;

COMMIT;
