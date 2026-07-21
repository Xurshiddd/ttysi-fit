-- +goose Up
-- +goose StatementBegin
-- Chellenjlar (CLAUDE.md §16). Kontent kodga yozilmaydi — admin panel yaratadi.
--
-- Sxema ataylab ikki qismli:
--   * umumiy maydonlar (hamma turda bor) — ustun;
--   * turga xos parametrlar — `config JSONB`.
-- Shu sababli yangi tur qo'shish (masalan "calories") migration talab qilmaydi:
-- admin yangi turni tanlaydi va uning maydonlarini config'ga yozadi.
CREATE TABLE challenges (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    type         VARCHAR(50)  NOT NULL,                        -- steps | distance | active_min | faculty_vs | group_vs | custom
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    scope        VARCHAR(50)  NOT NULL DEFAULT 'university',   -- university | faculty | group
    starts_at    TIMESTAMPTZ,
    ends_at      TIMESTAMPTZ,
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft',        -- draft | active | finished
    reward_coins INTEGER      NOT NULL DEFAULT 0,

    -- Turga xos parametrlar: {"target_steps":10000} / {"target_km":42,"metric":"distance"} ...
    config       JSONB        NOT NULL DEFAULT '{}',

    -- Kengayish uchun (§4.1)
    cover_url    TEXT,
    metadata     JSONB
);

-- Aktiv chellenjlarni tez topish (mobil ro'yxat shu bo'yicha so'raydi).
CREATE INDEX idx_challenges_active ON challenges(status, ends_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_challenges_type ON challenges(type) WHERE deleted_at IS NULL;

-- Foydalanuvchi–chellenj bog'lanishi va progress.
CREATE TABLE user_challenges (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id      UUID NOT NULL REFERENCES users(id),
    challenge_id UUID NOT NULL REFERENCES challenges(id),

    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    -- progress — turga qarab hisoblanadi (qadam yig'indisi, masofa, ...).
    -- Snapshot: og'ir agregatsiyani har so'rovda takrorlamaslik uchun.
    progress     DOUBLE PRECISION NOT NULL DEFAULT 0,

    -- Mukofot bir marta beriladi (FIT Coin ledger'iga yozilgach true).
    reward_granted BOOLEAN NOT NULL DEFAULT FALSE
);

-- Bitta foydalanuvchi bitta chellenjga bir marta qo'shiladi.
CREATE UNIQUE INDEX idx_user_challenges_unique ON user_challenges(user_id, challenge_id);
CREATE INDEX idx_user_challenges_user ON user_challenges(user_id, completed_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_challenges;
DROP TABLE IF EXISTS challenges;
-- +goose StatementEnd
