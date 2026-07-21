-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS faculties (
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS faculties;
-- +goose StatementEnd
