-- +goose Up
-- +goose StatementBegin
-- Kunlik faollik: bir foydalanuvchi uchun bir kunda bitta yozuv (upsert).
CREATE TABLE activities (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,

    user_id       UUID NOT NULL REFERENCES users(id),
    activity_date DATE NOT NULL,                         -- qaysi kun

    steps         INTEGER NOT NULL DEFAULT 0,
    calories      DOUBLE PRECISION NOT NULL DEFAULT 0,   -- kkal
    distance_m    DOUBLE PRECISION NOT NULL DEFAULT 0,   -- masofa (metr)
    active_min    INTEGER NOT NULL DEFAULT 0,            -- faol daqiqalar
    source        VARCHAR(50)                            -- manual | healthkit | googlefit
);

-- Bir kun — bir yozuv (upsert kaliti, ON CONFLICT uchun oddiy unique)
CREATE UNIQUE INDEX idx_activities_user_date_unique
    ON activities(user_id, activity_date);

-- Tez o'qish (oxirgi kunlar bo'yicha)
CREATE INDEX idx_activities_user_date ON activities(user_id, activity_date DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS activities;
-- +goose StatementEnd
