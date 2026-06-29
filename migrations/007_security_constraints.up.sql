-- Migration 007: RLS policies and payment idempotency constraints
-- CHK060: PostgreSQL Row-Level Security for supplier data isolation
-- CHK061: Payment idempotency DB-level unique constraint

-- ============================================================
-- CHK061: Add unique constraint for payment idempotency
-- Prevents duplicate active payments per order per channel
-- ============================================================

-- First, close any duplicate active payments (cleanup)
UPDATE payment_transaction
SET status = 'closed', updated_at = NOW()
WHERE id NOT IN (
    SELECT MIN(id)
    FROM payment_transaction
    WHERE status IN ('created', 'paying')
    GROUP BY order_id, channel
)
AND status IN ('created', 'paying');

-- Add unique index for active payments per order per channel
-- This enforces at the DB level that only one active payment exists per order+channel
CREATE UNIQUE INDEX IF NOT EXISTS uk_payment_order_channel_active
    ON payment_transaction (order_id, channel)
    WHERE status IN ('created', 'paying');

-- ============================================================
-- CHK060: Row-Level Security policies for supplier data isolation
-- ============================================================

-- Enable RLS on product table
ALTER TABLE product ENABLE ROW LEVEL SECURITY;

-- Policy: Suppliers can only see their own products
CREATE POLICY product_supplier_isolation ON product
    FOR ALL
    USING (
        supplier_id IS NULL  -- platform products (no supplier)
        OR supplier_id = current_setting('app.current_supplier_id', true)::bigint
        OR current_setting('app.current_user_type', true) = 'admin'  -- platform admins see all
    );

-- Enable RLS on main_order table
ALTER TABLE main_order ENABLE ROW LEVEL SECURITY;

-- Policy: Suppliers can only see orders for their products
CREATE POLICY order_supplier_isolation ON main_order
    FOR ALL
    USING (
        product_id IN (
            SELECT id FROM product
            WHERE supplier_id IS NULL
               OR supplier_id = current_setting('app.current_supplier_id', true)::bigint
        )
        OR current_setting('app.current_user_type', true) = 'admin'
    );

-- ============================================================
-- Order status CHECK constraint (defense in depth)
-- ============================================================

-- Add CHECK constraint for valid order statuses
ALTER TABLE main_order
    ADD CONSTRAINT chk_order_status
    CHECK (order_status IN (
        'pending_pay', 'paid_full', 'pending_travel', 'in_travel',
        'completed', 'cancelled', 'refunding', 'refunded', 'closed'
    ));

-- Add CHECK constraint for valid product statuses
ALTER TABLE product
    ADD CONSTRAINT chk_product_status
    CHECK (status IN (
        'draft', 'pending_review', 'approved', 'suspended', 'change_pending_review'
    ));
