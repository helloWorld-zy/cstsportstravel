-- 005: Admin Domain tables
-- Ref: data-model.md Admin Domain

CREATE TABLE IF NOT EXISTS admin_user (
    id                  BIGSERIAL       PRIMARY KEY,
    username            VARCHAR(50)     NOT NULL UNIQUE,
    password_hash       VARCHAR(255)    NOT NULL,
    real_name           VARCHAR(100)    NOT NULL,
    phone               VARCHAR(20),
    email               VARCHAR(200),
    supplier_id         BIGINT,
    status              VARCHAR(20)     NOT NULL DEFAULT 'active',
    must_change_password BOOLEAN        NOT NULL DEFAULT true,
    totp_secret         VARCHAR(255),
    last_login_at       TIMESTAMP,
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_username ON admin_user(username);
CREATE INDEX IF NOT EXISTS idx_admin_supplier ON admin_user(supplier_id);

COMMENT ON TABLE admin_user IS '后台管理员表';
COMMENT ON COLUMN admin_user.status IS 'active/locked/disabled';

CREATE TABLE IF NOT EXISTS role (
    id              BIGSERIAL       PRIMARY KEY,
    role_name       VARCHAR(50)     NOT NULL UNIQUE,
    role_code       VARCHAR(50)     NOT NULL UNIQUE,
    description     VARCHAR(200),
    is_system       BOOLEAN         NOT NULL DEFAULT false,
    status          VARCHAR(20)     NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE role IS '角色表';

CREATE TABLE IF NOT EXISTS permission (
    id                  BIGSERIAL       PRIMARY KEY,
    permission_name     VARCHAR(100)    NOT NULL,
    permission_code     VARCHAR(100)    NOT NULL UNIQUE,
    permission_type     VARCHAR(20)     NOT NULL,       -- menu/button/api/data
    parent_id           BIGINT          REFERENCES permission(id),
    resource_path       VARCHAR(200),
    http_method         VARCHAR(10),
    description         VARCHAR(200),
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE permission IS '权限表';
COMMENT ON COLUMN permission.permission_type IS 'menu/button/api/data';

CREATE TABLE IF NOT EXISTS menu (
    id                  BIGSERIAL       PRIMARY KEY,
    menu_name           VARCHAR(100)    NOT NULL,
    menu_path           VARCHAR(200),
    component_name      VARCHAR(200),
    icon                VARCHAR(100),
    parent_id           BIGINT          REFERENCES menu(id),
    sort_order          INTEGER         NOT NULL DEFAULT 0,
    permission_code     VARCHAR(100),
    status              VARCHAR(20)     NOT NULL DEFAULT 'active',
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE menu IS '菜单表';

CREATE TABLE IF NOT EXISTS admin_user_role (
    admin_user_id   BIGINT          NOT NULL REFERENCES admin_user(id),
    role_id         BIGINT          NOT NULL REFERENCES role(id),
    PRIMARY KEY (admin_user_id, role_id)
);

COMMENT ON TABLE admin_user_role IS '管理员-角色关联表';

CREATE TABLE IF NOT EXISTS role_permission (
    role_id         BIGINT          NOT NULL REFERENCES role(id),
    permission_id   BIGINT          NOT NULL REFERENCES permission(id),
    PRIMARY KEY (role_id, permission_id)
);

COMMENT ON TABLE role_permission IS '角色-权限关联表';

CREATE TABLE IF NOT EXISTS role_menu (
    role_id         BIGINT          NOT NULL REFERENCES role(id),
    menu_id         BIGINT          NOT NULL REFERENCES menu(id),
    PRIMARY KEY (role_id, menu_id)
);

COMMENT ON TABLE role_menu IS '角色-菜单关联表';

CREATE TABLE IF NOT EXISTS audit_log (
    id              BIGSERIAL       PRIMARY KEY,
    operator_id     BIGINT,
    operator_type   VARCHAR(20)     NOT NULL,       -- user/admin/system
    action          VARCHAR(100)    NOT NULL,
    target_type     VARCHAR(50)     NOT NULL,
    target_id       BIGINT,
    detail          JSONB,
    ip_address      VARCHAR(45),
    user_agent      VARCHAR(500),
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_operator ON audit_log(operator_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_target ON audit_log(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at DESC);

COMMENT ON TABLE audit_log IS '审计日志表';
COMMENT ON COLUMN audit_log.operator_type IS 'user/admin/system';
