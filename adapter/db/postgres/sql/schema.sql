CREATE TABLE IF NOT EXISTS version (
  version    INTEGER     NOT NULL DEFAULT 1,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO version (version) VALUES (1);

CREATE TABLE organizations (
    id          UUID        PRIMARY KEY DEFAULT uuidv7(),
    name        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE accounts (
    id              UUID        PRIMARY KEY DEFAULT uuidv7(),
    name            TEXT        NOT NULL,
    email           TEXT        NOT NULL UNIQUE,
    phone_number    TEXT        NOT NULL,
    password_hash   TEXT        NOT NULL,
    type            TEXT        NOT NULL CHECK (type IN ('admin', 'org_admin', 'user')),
    blocked         BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE account_organizations (
    account_id      UUID        NOT NULL,
    organization_id UUID        NOT NULL,
    PRIMARY KEY (account_id, organization_id),

    CONSTRAINT fk_account_organizations_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_account_organizations_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT
);

CREATE INDEX idx_account_organizations_account ON account_organizations (account_id);
CREATE INDEX idx_account_organizations_organization ON account_organizations (organization_id);

CREATE TABLE account_sessions (
    id          UUID        PRIMARY KEY DEFAULT uuidv7(),
    account_id  UUID        NOT NULL,
    token_hash  TEXT        NOT NULL UNIQUE,
    user_agent  TEXT,
    ip_address  INET,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_account_sessions_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX idx_account_sessions_account ON account_sessions (account_id);
CREATE INDEX idx_account_sessions_expires_at ON account_sessions (expires_at);

CREATE TABLE groups (
    id              UUID        PRIMARY KEY DEFAULT uuidv7(),
    name            TEXT        NOT NULL,
    organization_id UUID        NOT NULL,
    parent_id       UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_groups_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
    CONSTRAINT fk_groups_parent FOREIGN KEY (parent_id) REFERENCES groups(id) ON DELETE SET NULL
);

CREATE INDEX idx_groups_organization ON groups (organization_id);
CREATE INDEX idx_groups_parent ON groups (parent_id);

CREATE TABLE vehicles (
    id              UUID        PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID        NOT NULL,
    group_id        UUID,
    ivms_type       TEXT        NOT NULL CHECK (ivms_type IN ('cmsv6')),
    external_id     TEXT        NOT NULL,
    plate_number    TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_vehicles_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
    CONSTRAINT fk_vehicles_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL,
    CONSTRAINT uq_vehicles_ivms_type_external_id UNIQUE (ivms_type, external_id)
);

CREATE INDEX idx_vehicles_organization ON vehicles (organization_id);
CREATE INDEX idx_vehicles_group ON vehicles (group_id);