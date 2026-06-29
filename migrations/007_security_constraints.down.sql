-- Rollback migration 007: Remove RLS policies and security constraints

-- Remove CHECK constraints
ALTER TABLE main_order DROP CONSTRAINT IF EXISTS chk_order_status;
ALTER TABLE product DROP CONSTRAINT IF EXISTS chk_product_status;

-- Remove RLS policies
DROP POLICY IF EXISTS order_supplier_isolation ON main_order;
DROP POLICY IF EXISTS product_supplier_isolation ON product;

-- Disable RLS
ALTER TABLE main_order DISABLE ROW LEVEL SECURITY;
ALTER TABLE product DISABLE ROW LEVEL SECURITY;

-- Remove payment idempotency index
DROP INDEX IF EXISTS uk_payment_order_channel_active;
