-- +goose Up
-- +goose StatementBegin
-- Nullable + default: tez ishlaydi, jadval lock olmaydi (CLAUDE.md 4.2).
ALTER TABLE users ADD COLUMN language VARCHAR(5) DEFAULT 'uz';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN language;
-- +goose StatementEnd
