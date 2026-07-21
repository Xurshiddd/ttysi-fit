-- +goose Up
-- +goose StatementBegin
-- HEMIS roster'da bir xil email takrorlanishi mumkin (junk/umumiy email).
-- Shuning uchun email uniqueligi faqat o'zi ro'yxatdan o'tgan (hemis_id IS NULL) userlarga.
DROP INDEX IF EXISTS idx_users_email_unique;
CREATE UNIQUE INDEX idx_users_email_unique ON users(email)
    WHERE email <> '' AND hemis_id IS NULL AND deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_email_unique;
CREATE UNIQUE INDEX idx_users_email_unique ON users(email)
    WHERE email <> '' AND deleted_at IS NULL;
-- +goose StatementEnd
