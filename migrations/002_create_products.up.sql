-- 002: Product Domain tables
-- Ref: data-model.md Product Domain

CREATE TABLE IF NOT EXISTS category (
    id              BIGSERIAL       PRIMARY KEY,
    name            VARCHAR(100)    NOT NULL,
    parent_id       BIGINT          REFERENCES category(id),
    icon_url        VARCHAR(500),
    sort_order      INTEGER         NOT NULL DEFAULT 0,
    status          VARCHAR(20)     NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE category IS '产品分类表';

CREATE TABLE IF NOT EXISTS product (
    id                  BIGSERIAL       PRIMARY KEY,
    product_no          VARCHAR(30)     NOT NULL UNIQUE,
    product_name        VARCHAR(200)    NOT NULL,
    category_id         BIGINT          NOT NULL REFERENCES category(id),
    product_type        VARCHAR(30)     NOT NULL DEFAULT 'group_tour',
    origin_city         VARCHAR(50)     NOT NULL,
    destination_cities  JSONB           NOT NULL,
    destination_tags    JSONB,
    days                INTEGER         NOT NULL,
    nights              INTEGER         NOT NULL,
    transport_mode      VARCHAR(50),
    min_group_size      INTEGER         NOT NULL DEFAULT 2,
    max_group_size      INTEGER         NOT NULL DEFAULT 50,
    product_grade       VARCHAR(20),
    cover_image         VARCHAR(500),
    images              JSONB,
    summary             VARCHAR(500),
    description         TEXT,
    fee_included        TEXT,
    fee_excluded        TEXT,
    booking_notes       TEXT,
    status              VARCHAR(30)     NOT NULL DEFAULT 'draft',
    reject_reason       VARCHAR(500),
    supplier_id         BIGINT,
    commission_rate     DECIMAL(5,4)    DEFAULT 0,
    view_count          INTEGER         NOT NULL DEFAULT 0,
    order_count         INTEGER         NOT NULL DEFAULT 0,
    satisfaction_rate   DECIMAL(5,2),
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_status_created ON product(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_product_category ON product(category_id);
CREATE INDEX IF NOT EXISTS idx_product_supplier ON product(supplier_id);

COMMENT ON TABLE product IS '产品表';
COMMENT ON COLUMN product.destination_cities IS '目的地城市列表 (JSONB)';
COMMENT ON COLUMN product.status IS 'draft/pending_review/approved/suspended/change_pending_review';

CREATE TABLE IF NOT EXISTS itinerary (
    id              BIGSERIAL       PRIMARY KEY,
    product_id      BIGINT          NOT NULL REFERENCES product(id),
    day_no          INTEGER         NOT NULL,
    title           VARCHAR(200)    NOT NULL,
    description     TEXT,
    meals           JSONB,          -- {breakfast, lunch, dinner} booleans
    hotel           VARCHAR(200),
    transport       VARCHAR(100),
    spots           JSONB,          -- [{name, description, duration, image}]
    images          JSONB,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    UNIQUE(product_id, day_no)
);

CREATE INDEX IF NOT EXISTS idx_itinerary_product_day ON itinerary(product_id, day_no);

COMMENT ON TABLE itinerary IS '行程安排表';
COMMENT ON COLUMN itinerary.meals IS '用餐计划 {breakfast, lunch, dinner}';
COMMENT ON COLUMN itinerary.spots IS '景点列表 [{name, description, duration, image}]';

CREATE TABLE IF NOT EXISTS departure_date (
    id              BIGSERIAL       PRIMARY KEY,
    product_id      BIGINT          NOT NULL REFERENCES product(id),
    departure_date  DATE            NOT NULL,
    return_date     DATE            NOT NULL,
    adult_price     INTEGER         NOT NULL,       -- cents
    child_price     INTEGER         NOT NULL,       -- cents
    infant_price    INTEGER         NOT NULL DEFAULT 0,
    single_supplement INTEGER       NOT NULL DEFAULT 0,
    total_stock     INTEGER         NOT NULL,
    sold_count      INTEGER         NOT NULL DEFAULT 0,
    locked_count    INTEGER         NOT NULL DEFAULT 0,
    cutoff_days     INTEGER         NOT NULL DEFAULT 1,
    status          VARCHAR(20)     NOT NULL DEFAULT 'open',
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    UNIQUE(product_id, departure_date)
);

CREATE INDEX IF NOT EXISTS idx_departure_product_date ON departure_date(product_id, departure_date);

COMMENT ON TABLE departure_date IS '出发日期及库存表';
COMMENT ON COLUMN departure_date.adult_price IS '成人价格 (分)';
COMMENT ON COLUMN departure_date.child_price IS '儿童价格 (分)';
COMMENT ON COLUMN departure_date.single_supplement IS '单房差 (分)';
COMMENT ON COLUMN departure_date.status IS 'open/full/closed/cancelled';

CREATE TABLE IF NOT EXISTS price_rule (
    id              BIGSERIAL       PRIMARY KEY,
    product_id      BIGINT          NOT NULL REFERENCES product(id),
    date_from       DATE            NOT NULL,
    date_to         DATE            NOT NULL,
    adult_price     INTEGER,        -- cents, null = use departure default
    child_price     INTEGER,
    infant_price    INTEGER,
    single_supplement INTEGER,
    price_type      VARCHAR(20)     NOT NULL DEFAULT 'standard',
    priority        INTEGER         NOT NULL DEFAULT 0,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_price_rule_product ON price_rule(product_id, date_from, date_to);

COMMENT ON TABLE price_rule IS '价格规则表';
COMMENT ON COLUMN price_rule.price_type IS 'standard/early_bird/promotion';

CREATE TABLE IF NOT EXISTS refund_rule (
    id                  BIGSERIAL       PRIMARY KEY,
    product_id          BIGINT          REFERENCES product(id),  -- null for global template
    rule_name           VARCHAR(100)    NOT NULL,
    days_before_min     INTEGER         NOT NULL,
    days_before_max     INTEGER,        -- null = no upper bound
    refund_percentage   DECIMAL(5,2)    NOT NULL,
    description         VARCHAR(500),
    is_template         BOOLEAN         NOT NULL DEFAULT false,
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refund_rule_product ON refund_rule(product_id);

COMMENT ON TABLE refund_rule IS '退改规则表';
COMMENT ON COLUMN refund_rule.refund_percentage IS '退款百分比 (0.00-100.00)';

CREATE TABLE IF NOT EXISTS product_review (
    id              BIGSERIAL       PRIMARY KEY,
    product_id      BIGINT          NOT NULL REFERENCES product(id),
    user_id         BIGINT          NOT NULL REFERENCES user_account(id),
    order_id        BIGINT          NOT NULL,  -- FK to main_order added in 003
    rating          INTEGER         NOT NULL,
    content         TEXT,
    images          JSONB,
    is_anonymous    BOOLEAN         NOT NULL DEFAULT false,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    UNIQUE(order_id)
);

CREATE INDEX IF NOT EXISTS idx_review_product ON product_review(product_id, created_at DESC);

COMMENT ON TABLE product_review IS '产品评价表';
COMMENT ON COLUMN product_review.rating IS '评分 (1-5)';

CREATE TABLE IF NOT EXISTS destination (
    id              BIGSERIAL       PRIMARY KEY,
    name            VARCHAR(100)    NOT NULL,
    province        VARCHAR(50),
    city            VARCHAR(50),
    cover_image     VARCHAR(500),
    description     TEXT,
    sort_order      INTEGER         NOT NULL DEFAULT 0,
    status          VARCHAR(20)     NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE destination IS '目的地表';
