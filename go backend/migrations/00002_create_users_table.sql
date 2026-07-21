-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,

    -- Asosiy ma'lumotlar
    full_name   VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL UNIQUE,
    phone       VARCHAR(20),
    password    VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL DEFAULT 'student',

    -- Universitet ma'lumotlari
    faculty_id  UUID REFERENCES faculties(id),
    department  VARCHAR(255),
    course      SMALLINT,
    group_name  VARCHAR(100),

    -- Kengayish uchun
    avatar_url  TEXT,
    bio         TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    metadata    JSONB
);

CREATE INDEX idx_users_faculty_id ON users(faculty_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_active ON users(role, faculty_id) WHERE deleted_at IS NULL AND is_active = TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
