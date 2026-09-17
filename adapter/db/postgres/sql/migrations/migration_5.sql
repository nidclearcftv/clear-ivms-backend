-- Adds the equipment_models table: an organization's equipment catalog.
-- public defaults to false and is only ever changed via a dedicated
-- admin-only route, never through the ordinary update endpoint.
CREATE TABLE equipment_models (
    id              UUID        PRIMARY KEY DEFAULT uuidv7(),
    name            TEXT        NOT NULL,
    description     TEXT        NOT NULL,
    type            TEXT        NOT NULL CHECK (type IN ('primary', 'accessory')),
    public          BOOLEAN     NOT NULL DEFAULT FALSE,
    organization_id UUID        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_equipment_models_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT
);

CREATE INDEX idx_equipment_models_organization ON equipment_models (organization_id);
