-- 003: Order Domain tables
-- Ref: data-model.md Order Domain

CREATE TABLE IF NOT EXISTS main_order (
    id                      BIGSERIAL       PRIMARY KEY,
    order_no                VARCHAR(30)     NOT NULL UNIQUE,
    user_id                 BIGINT          NOT NULL REFERENCES user_account(id),
    product_id              BIGINT          NOT NULL REFERENCES product(id),
    departure_id            BIGINT          NOT NULL REFERENCES departure_date(id),
    order_status            VARCHAR(30)     NOT NULL DEFAULT 'pending_pay',
    payment_status          VARCHAR(30)     NOT NULL DEFAULT 'unpaid',
    total_amount            INTEGER         NOT NULL,       -- cents
    discount_amount         INTEGER         NOT NULL DEFAULT 0,
    payable_amount          INTEGER         NOT NULL,       -- cents
    adult_count             INTEGER         NOT NULL,
    child_count             INTEGER         NOT NULL DEFAULT 0,
    infant_count            INTEGER         NOT NULL DEFAULT 0,
    single_supplement_amount INTEGER        NOT NULL DEFAULT 0,
    addon_amount            INTEGER         NOT NULL DEFAULT 0,
    contact_name            VARCHAR(100)    NOT NULL,
    contact_phone           VARCHAR(20)     NOT NULL,
    channel                 VARCHAR(20)     NOT NULL DEFAULT 'web',
    remark                  VARCHAR(500),
    paid_at                 TIMESTAMP,
    departed_at             TIMESTAMP,
    completed_at            TIMESTAMP,
    cancelled_at            TIMESTAMP,
    cancel_reason           VARCHAR(500),
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_user_status ON main_order(user_id, order_status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_order_product ON main_order(product_id);
CREATE INDEX IF NOT EXISTS idx_order_departure ON main_order(departure_id);
CREATE INDEX IF NOT EXISTS idx_order_created ON main_order(created_at DESC);

COMMENT ON TABLE main_order IS '主订单表';
COMMENT ON COLUMN main_order.total_amount IS '总金额 (分)';
COMMENT ON COLUMN main_order.payable_amount IS '应付金额 (分)';
COMMENT ON COLUMN main_order.order_status IS 'pending_pay/paid_full/pending_travel/in_travel/completed/cancelled/refunding/refunded/closed';
COMMENT ON COLUMN main_order.channel IS 'web/miniapp/admin';

CREATE TABLE IF NOT EXISTS sub_order (
    id              BIGSERIAL       PRIMARY KEY,
    main_order_id   BIGINT          NOT NULL REFERENCES main_order(id),
    sub_order_no    VARCHAR(30)     NOT NULL UNIQUE,
    resource_type   VARCHAR(30)     NOT NULL,       -- insurance/transfer
    resource_id     BIGINT,
    resource_name   VARCHAR(200)    NOT NULL,
    supplier_id     BIGINT,
    status          VARCHAR(20)     NOT NULL DEFAULT 'pending',
    amount          INTEGER         NOT NULL,       -- cents
    commission_rate DECIMAL(5,4)    DEFAULT 0,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sub_order_main ON sub_order(main_order_id);

COMMENT ON TABLE sub_order IS '子订单表';
COMMENT ON COLUMN sub_order.resource_type IS 'insurance/transfer';

CREATE TABLE IF NOT EXISTS order_status_log (
    id              BIGSERIAL       PRIMARY KEY,
    order_id        BIGINT          NOT NULL REFERENCES main_order(id),
    from_status     VARCHAR(30)     NOT NULL,
    to_status       VARCHAR(30)     NOT NULL,
    operator_type   VARCHAR(20)     NOT NULL,       -- system/user/admin
    operator_id     BIGINT,
    reason          VARCHAR(500),
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_osl_order ON order_status_log(order_id, created_at DESC);

COMMENT ON TABLE order_status_log IS '订单状态变更日志表';

CREATE TABLE IF NOT EXISTS order_traveller (
    id              BIGSERIAL       PRIMARY KEY,
    order_id        BIGINT          NOT NULL REFERENCES main_order(id),
    real_name       TEXT            NOT NULL,       -- AES-256-GCM encrypted
    id_card_no      TEXT            NOT NULL,       -- AES-256-GCM encrypted
    phone           VARCHAR(20),
    birth_date      DATE,
    gender          VARCHAR(10),
    is_child        BOOLEAN         NOT NULL DEFAULT false,
    is_infant       BOOLEAN         NOT NULL DEFAULT false,
    linked_adult_id BIGINT          REFERENCES order_traveller(id),
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ot_order ON order_traveller(order_id);

COMMENT ON TABLE order_traveller IS '订单出行人表';
COMMENT ON COLUMN order_traveller.real_name IS '姓名 (AES-256-GCM 加密)';
COMMENT ON COLUMN order_traveller.id_card_no IS '身份证号 (AES-256-GCM 加密)';

-- Add FK from product_review.order_id to main_order.id
ALTER TABLE product_review
    ADD CONSTRAINT fk_review_order FOREIGN KEY (order_id) REFERENCES main_order(id);
