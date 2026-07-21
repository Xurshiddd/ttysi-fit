-- +goose Up
-- +goose StatementBegin
-- Mashg'ulotlar — video darsliklar (CLAUDE.md §16.3: admin kontent kiritadi).
--
-- KATEGORIYA haqida: u ataylab erkin matn (enum emas, alohida jadval ham emas).
-- Enum bo'lsa "Yoga" qo'shish uchun kod o'zgartirib redeploy qilish kerak
-- bo'lardi — §16 aynan shuni taqiqlaydi. Admin mavjud kategoriyalardan
-- tanlaydi (GET /training-categories DISTINCT qaytaradi) yoki yangisini yozadi.
--
-- LEVEL esa enum: bu chegaralangan shkala (boshlang'ich/o'rta/yuqori), kontent
-- emas. Uni kengaytirish ehtimoli yo'q darajada past.
CREATE TABLE trainings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    title        VARCHAR(255) NOT NULL,
    description  TEXT,

    category     VARCHAR(100),
    level        VARCHAR(20) NOT NULL DEFAULT 'beginner',  -- beginner | intermediate | advanced

    video_url    TEXT NOT NULL,
    thumbnail_url TEXT,
    -- duration_min: video davomiyligi. NULL — ko'rsatilmagan.
    duration_min SMALLINT CHECK (duration_min IS NULL OR duration_min > 0),

    status       VARCHAR(20) NOT NULL DEFAULT 'draft',      -- draft | published
    published_at TIMESTAMPTZ,
    views        INTEGER NOT NULL DEFAULT 0,
    -- sort_order: admin ro'yxat tartibini boshqaradi (kichik — yuqorida).
    sort_order   INTEGER NOT NULL DEFAULT 0,

    metadata     JSONB
);

CREATE INDEX idx_trainings_published ON trainings(status, sort_order, published_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_trainings_category ON trainings(category) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trainings;
-- +goose StatementEnd
