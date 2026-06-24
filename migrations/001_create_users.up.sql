-- 001: User Domain tables
-- Ref: data-model.md User Domain

CREATE TABLE IF NOT EXISTS user_account (
    id              BIGSERIAL       PRIMARY KEY,
    phone           VARCHAR(20)     NOT NULL UNIQUE,
    password_hash   VARCHAR(255),
    nickname        VARCHAR(50)     NOT NULL,
    avatar_url      VARCHAR(500),
    real_name       TEXT,           -- AES-256-GCM encrypted
    id_card_no      TEXT,           -- AES-256-GCM encrypted
    real_name_status VARCHAR(20)    NOT NULL DEFAULT 'unverified',
    member_level    INTEGER         NOT NULL DEFAULT 1,
    status          VARCHAR(20)     NOT NULL DEFAULT 'active',
    wechat_openid   VARCHAR(100)    UNIQUE,
    wechat_unionid  VARCHAR(100),
    sms_code        VARCHAR(6),
    sms_code_expires_at TIMESTAMP,
    sms_send_count_today INTEGER    NOT NULL DEFAULT 0,
    login_fail_count INTEGER        NOT NULL DEFAULT 0,
    locked_until    TIMESTAMP,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_phone ON user_account(phone);
CREATE INDEX IF NOT EXISTS idx_user_wechat ON user_account(wechat_openid);

COMMENT ON TABLE user_account IS '用户账户表';
COMMENT ON COLUMN user_account.real_name IS '真实姓名 (AES-256-GCM 加密)';
COMMENT ON COLUMN user_account.id_card_no IS '身份证号 (AES-256-GCM 加密)';
COMMENT ON COLUMN user_account.real_name_status IS 'unverified/pending/verified/rejected';

CREATE TABLE IF NOT EXISTS real_name_verification (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES user_account(id),
    real_name       TEXT            NOT NULL,  -- AES-256-GCM encrypted
    id_card_no      TEXT            NOT NULL,  -- AES-256-GCM encrypted
    status          VARCHAR(20)     NOT NULL DEFAULT 'pending',
    reject_reason   VARCHAR(500),
    verified_at     TIMESTAMP,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rnv_user ON real_name_verification(user_id);

COMMENT ON TABLE real_name_verification IS '实名认证记录表';

CREATE TABLE IF NOT EXISTS frequent_traveller (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES user_account(id),
    real_name       TEXT            NOT NULL,  -- AES-256-GCM encrypted
    id_card_no      TEXT            NOT NULL,  -- AES-256-GCM encrypted
    phone           VARCHAR(20),
    birth_date      DATE,
    gender          VARCHAR(10),
    is_default      BOOLEAN         NOT NULL DEFAULT false,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ft_user ON frequent_traveller(user_id);

COMMENT ON TABLE frequent_traveller IS '常用出行人表';
COMMENT ON COLUMN frequent_traveller.real_name IS '姓名 (AES-256-GCM 加密)';
COMMENT ON COLUMN frequent_traveller.id_card_no IS '身份证号 (AES-256-GCM 加密)';
