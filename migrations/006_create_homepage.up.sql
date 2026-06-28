-- 006: Homepage content management tables
-- Ref: Phase 10 - Frontend Enhancement

CREATE TABLE IF NOT EXISTS homepage_banner (
    id              BIGSERIAL       PRIMARY KEY,
    title           VARCHAR(200)    NOT NULL,
    image_url       VARCHAR(500)    NOT NULL,
    link_url        VARCHAR(500),
    position        VARCHAR(50)     NOT NULL DEFAULT 'home_top',
    sort_order      INTEGER         NOT NULL DEFAULT 0,
    status          VARCHAR(20)     NOT NULL DEFAULT 'active',
    start_at        TIMESTAMP,
    end_at          TIMESTAMP,
    created_by      BIGINT,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_banner_position_status ON homepage_banner(position, status, sort_order);
CREATE INDEX IF NOT EXISTS idx_banner_active ON homepage_banner(status, start_at, end_at);

COMMENT ON TABLE homepage_banner IS '首页轮播图管理表';
COMMENT ON COLUMN homepage_banner.position IS '展示位置: home_top(首页顶部)';
COMMENT ON COLUMN homepage_banner.status IS '状态: active/inactive';
COMMENT ON COLUMN homepage_banner.start_at IS '生效开始时间 (null=立即生效)';
COMMENT ON COLUMN homepage_banner.end_at IS '生效结束时间 (null=永久有效)';
