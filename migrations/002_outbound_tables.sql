-- Migration 002: Outbound Product Extensions
-- Creates country table, visa_material_template table, and extends product table
-- for outbound travel products with visa information.

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════
-- Country/Region Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS country (
    id                      BIGSERIAL PRIMARY KEY,
    tenant_id               BIGINT NOT NULL,
    name_cn                 VARCHAR(100) NOT NULL,
    name_en                 VARCHAR(100) NOT NULL,
    continent               VARCHAR(20) NOT NULL CHECK (continent IN ('asia', 'europe', 'north_america', 'south_america', 'oceania', 'africa')),
    visa_type               VARCHAR(20) NOT NULL CHECK (visa_type IN ('visa_free', 'visa_on_arrival', 'e_visa', 'visa_required')),
    visa_processing_days    INT,
    passport_validity_months INT DEFAULT 6,
    entry_policy            JSONB,
    cash_regulation         JSONB,
    prohibited_items        JSONB,
    entry_card_guide        JSONB,
    customs_guide           JSONB,
    emergency_contacts      JSONB,
    status                  VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_country_continent ON country (tenant_id, continent);
CREATE INDEX IF NOT EXISTS idx_country_visa_type ON country (tenant_id, visa_type);

-- ═══════════════════════════════════════════════════════════════════════════
-- Product Table Extensions (outbound travel fields)
-- ═══════════════════════════════════════════════════════════════════════════

ALTER TABLE product ADD COLUMN IF NOT EXISTS product_type VARCHAR(30) DEFAULT 'domestic_group';
ALTER TABLE product ADD COLUMN IF NOT EXISTS destination_country_id BIGINT REFERENCES country(id);
ALTER TABLE product ADD COLUMN IF NOT EXISTS visa_info JSONB;
ALTER TABLE product ADD COLUMN IF NOT EXISTS international_flight_info JSONB;
ALTER TABLE product ADD COLUMN IF NOT EXISTS insurance_requirements JSONB;
ALTER TABLE product ADD COLUMN IF NOT EXISTS pre_trip_services JSONB;

CREATE INDEX IF NOT EXISTS idx_product_type ON product (tenant_id, product_type);
CREATE INDEX IF NOT EXISTS idx_product_country ON product (destination_country_id) WHERE destination_country_id IS NOT NULL;

-- ═══════════════════════════════════════════════════════════════════════════
-- Visa Material Template Table
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS visa_material_template (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    country_id          BIGINT NOT NULL REFERENCES country(id),
    occupation_type     VARCHAR(20) NOT NULL CHECK (occupation_type IN ('employed', 'freelance', 'retired', 'student', 'child')),
    material_type       VARCHAR(50) NOT NULL,
    material_name       VARCHAR(100) NOT NULL,
    is_required         BOOLEAN NOT NULL DEFAULT true,
    description         TEXT,
    file_format         VARCHAR(50),
    max_size_mb         INT DEFAULT 10,
    sort_order          INT NOT NULL DEFAULT 0,
    status              VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_visa_template_country ON visa_material_template (tenant_id, country_id, occupation_type);

COMMIT;
