-- +goose Up
-- +goose StatementBegin
-- users.faculty_id eski faculties ga bog'langan edi — uni uzamiz.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_faculty_id_fkey;

-- Eski faculties o'rniga HEMIS strukturasini saqlovchi yagona jadval.
DROP TABLE IF EXISTS faculties;

CREATE TABLE structures (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,

    -- HEMIS sync kaliti
    hemis_id        BIGINT NOT NULL UNIQUE,

    -- Asosiy ma'lumotlar
    name            VARCHAR(500) NOT NULL,
    code            VARCHAR(50),

    -- structureType (denormalizatsiya — joinsiz filtr uchun)
    structure_type_code VARCHAR(20),
    structure_type_name VARCHAR(255),

    -- localityType
    locality_type_code  VARCHAR(20),
    locality_type_name  VARCHAR(255),

    -- Daraxt: o'ziga havola
    parent_id       UUID REFERENCES structures(id),
    parent_hemis_id BIGINT,                          -- sync paytida bog'lash uchun

    active          BOOLEAN NOT NULL DEFAULT TRUE,

    -- Kelajakdagi qo'shimcha HEMIS maydonlari (forward-compat)
    raw             JSONB,
    synced_at       TIMESTAMPTZ
);

CREATE INDEX idx_structures_parent_id ON structures(parent_id);
CREATE INDEX idx_structures_type_code ON structures(structure_type_code);
CREATE INDEX idx_structures_code ON structures(code);
CREATE INDEX idx_structures_parent_hemis_id ON structures(parent_hemis_id);
CREATE INDEX idx_structures_active ON structures(structure_type_code) WHERE deleted_at IS NULL AND active = TRUE;
CREATE INDEX idx_structures_deleted_at ON structures(deleted_at) WHERE deleted_at IS NULL;

-- users.faculty_id endi structures ga ishora qiladi.
ALTER TABLE users
    ADD CONSTRAINT users_faculty_id_fkey
    FOREIGN KEY (faculty_id) REFERENCES structures(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_faculty_id_fkey;
DROP TABLE IF EXISTS structures;

CREATE TABLE faculties (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    name        VARCHAR(255) NOT NULL,
    short_name  VARCHAR(100),
    code        VARCHAR(50) UNIQUE,
    metadata    JSONB
);
CREATE INDEX idx_faculties_deleted_at ON faculties(deleted_at) WHERE deleted_at IS NULL;

ALTER TABLE users
    ADD CONSTRAINT users_faculty_id_fkey
    FOREIGN KEY (faculty_id) REFERENCES faculties(id);
-- +goose StatementEnd
