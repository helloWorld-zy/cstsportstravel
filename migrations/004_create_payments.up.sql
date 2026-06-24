-- 004: Payment Domain tables
-- Ref: data-model.md Payment Domain

CREATE TABLE IF NOT EXISTS payment_transaction (
    id              BIGSERIAL       PRIMARY KEY,
    order_id        BIGINT          NOT NULL REFERENCES main_order(id),
    payment_no      VARCHAR(30)     NOT NULL UNIQUE,
    channel         VARCHAR(20)     NOT NULL,       -- alipay/wechat/unionpay
    method          VARCHAR(30)     NOT NULL,       -- native/jsapi/h5/wap
    amount          INTEGER         NOT NULL,       -- cents
    status          VARCHAR(20)     NOT NULL DEFAULT 'created',
    channel_trade_no VARCHAR(100),
    paid_at         TIMESTAMP,
    expire_at       TIMESTAMP       NOT NULL,
    notify_url      VARCHAR(500)    NOT NULL,
    extra_params    JSONB,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_order ON payment_transaction(order_id, channel);

COMMENT ON TABLE payment_transaction IS '支付交易表';
COMMENT ON COLUMN payment_transaction.amount IS '支付金额 (分)';
COMMENT ON COLUMN payment_transaction.status IS 'created/paying/paid/failed/closed/refunded';
COMMENT ON COLUMN payment_transaction.channel IS 'alipay/wechat/unionpay';
COMMENT ON COLUMN payment_transaction.method IS 'native/jsapi/h5/wap';

CREATE TABLE IF NOT EXISTS refund_record (
    id                  BIGSERIAL       PRIMARY KEY,
    order_id            BIGINT          NOT NULL REFERENCES main_order(id),
    payment_id          BIGINT          NOT NULL REFERENCES payment_transaction(id),
    refund_no           VARCHAR(30)     NOT NULL UNIQUE,
    refund_amount       INTEGER         NOT NULL,       -- cents
    refund_reason       VARCHAR(500)    NOT NULL,
    refund_type         VARCHAR(20)     NOT NULL,       -- full/partial
    status              VARCHAR(20)     NOT NULL DEFAULT 'pending',
    approval_level      VARCHAR(30)     NOT NULL,       -- operator/finance_director/director
    approved_by         BIGINT,
    approved_at         TIMESTAMP,
    channel_refund_no   VARCHAR(100),
    completed_at        TIMESTAMP,
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refund_order ON refund_record(order_id);
CREATE INDEX IF NOT EXISTS idx_refund_status ON refund_record(status);

COMMENT ON TABLE refund_record IS '退款记录表';
COMMENT ON COLUMN refund_record.refund_amount IS '退款金额 (分)';
COMMENT ON COLUMN refund_record.refund_type IS 'full/partial';
COMMENT ON COLUMN refund_record.status IS 'pending/approved/processing/success/failed';
COMMENT ON COLUMN refund_record.approval_level IS 'operator/finance_director/director';
