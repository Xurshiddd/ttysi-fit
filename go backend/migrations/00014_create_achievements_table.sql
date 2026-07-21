-- +goose Up
-- +goose StatementBegin
-- Yutuqlar va sertifikatlar (CLAUDE.md §16.3). Kontent kodga yozilmaydi —
-- shablonni admin panel yaratadi, berish mezoni `criteria JSONB` da saqlanadi.
--
-- Sxema chellenjlar bilan bir xil andozada (§16.1):
--   * umumiy maydonlar — ustun;
--   * turga xos mezon — `criteria JSONB`.
-- Yangi yutuq turi qo'shish migration talab qilmaydi.
--
-- award_mode ikki xil haqiqiy ehtiyojni qoplaydi:
--   'auto'   — mezon ma'lumotdan hisoblanadi (100 000 qadam, 30 faol kun);
--   'manual' — admin qo'lda beradi (musobaqa g'olibi, oflayn tadbir ishtiroki),
--              chunki bunday yutuqni hech qanday mezon bilan ifodalab bo'lmaydi.
CREATE TABLE achievements (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    type         VARCHAR(50)  NOT NULL,                      -- steps_total | distance_total | active_days | challenge_count | manual
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    award_mode   VARCHAR(20)  NOT NULL DEFAULT 'auto',       -- auto | manual
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft',      -- draft | active | archived
    reward_coins INTEGER      NOT NULL DEFAULT 0,

    -- Turga xos mezon: {"threshold":100000} / {"threshold":42} ...
    criteria     JSONB        NOT NULL DEFAULT '{}',

    -- Ko'rinish va sertifikat shabloni (§4.1 — kengayish uchun nullable)
    icon_url     TEXT,
    cover_url    TEXT,
    -- certificate_template — PDF/rasm generatsiyasi qo'shilganda ishlatiladi.
    -- Hozir bo'sh turadi: yutuq berish undan mustaqil ishlaydi.
    certificate_template TEXT,
    metadata     JSONB
);

-- Mobil "aktiv yutuqlar" ro'yxati va avtomatik baholash shu bo'yicha o'qiydi.
CREATE INDEX idx_achievements_active ON achievements(status, award_mode) WHERE deleted_at IS NULL;
CREATE INDEX idx_achievements_type ON achievements(type) WHERE deleted_at IS NULL;

-- Berilgan yutuqlar. Sertifikat alohida jadval EMAS: sertifikat — shu
-- yozuvning ko'rinishi. Aks holda "kim nimani qozongan" ikki joyda saqlanib,
-- vaqt o'tib bir-biridan ajralib ketardi.
CREATE TABLE user_achievements (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id        UUID NOT NULL REFERENCES users(id),
    achievement_id UUID NOT NULL REFERENCES achievements(id),

    awarded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- awarded_by — qo'lda berilganda admin; avtomatik berilganda NULL.
    awarded_by     UUID REFERENCES users(id),

    -- progress_value — berilgan paytdagi qiymat (masalan 100 342 qadam).
    -- Sertifikatda ko'rsatish uchun saqlanadi: keyin faollik o'zgarsa ham
    -- sertifikatdagi raqam o'zgarmasligi kerak (snapshot).
    progress_value DOUBLE PRECISION,
    note           TEXT,

    -- certificate_url — generatsiya qilingan fayl (keyingi bosqich).
    certificate_url TEXT,
    -- reward_granted — FIT Coin mukofoti ledger'ga yozilganmi (bir marta).
    reward_granted BOOLEAN NOT NULL DEFAULT FALSE,
    metadata       JSONB
);

-- Bitta yutuq bitta foydalanuvchiga bir marta beriladi. Avtomatik baholash
-- takror ishga tushsa ham dublikat yozuv paydo bo'lmaydi (idempotentlik).
CREATE UNIQUE INDEX idx_user_achievements_unique ON user_achievements(user_id, achievement_id);
CREATE INDEX idx_user_achievements_user ON user_achievements(user_id, awarded_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS achievements;
-- +goose StatementEnd
