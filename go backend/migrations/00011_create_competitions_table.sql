-- +goose Up
-- +goose StatementBegin
-- Musobaqalar (CLAUDE.md §16.3). Chellenj bilan bir xil andoza:
-- umumiy maydonlar — ustun, turga xos parametrlar — `config JSONB`.
--
-- Chellenjdan asosiy farqi: musobaqaga RO'YXATDAN O'TILADI (ixtiyoriy emas,
-- joy soni cheklangan bo'lishi mumkin) va natija qo'lda/hakam tomonidan
-- kiritiladi — avtomatik `activities` dan hisoblanmaydi.
CREATE TABLE competitions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    type         VARCHAR(50)  NOT NULL,                        -- individual | team | faculty_vs | custom
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    scope        VARCHAR(50)  NOT NULL DEFAULT 'university',   -- university | faculty | group
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft',        -- draft | registration | ongoing | finished

    starts_at    TIMESTAMPTZ,
    ends_at      TIMESTAMPTZ,
    -- Ro'yxatdan o'tish muddati: shu vaqtdan keyin yangi ishtirokchi qabul
    -- qilinmaydi (NULL — musobaqa boshlanishigacha ochiq).
    reg_ends_at  TIMESTAMPTZ,

    location     VARCHAR(255),
    -- max_participants: NULL yoki 0 — cheklovsiz.
    max_participants INTEGER CHECK (max_participants IS NULL OR max_participants >= 0),
    reward_coins INTEGER NOT NULL DEFAULT 0,

    -- Turga xos: {"sport":"futbol","team_size":5} / {"distance_km":5} ...
    config       JSONB        NOT NULL DEFAULT '{}',

    cover_url    TEXT,
    metadata     JSONB
);

CREATE INDEX idx_competitions_active ON competitions(status, starts_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_competitions_type ON competitions(type) WHERE deleted_at IS NULL;

-- Ro'yxatdan o'tishlar.
CREATE TABLE competition_registrations (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id        UUID NOT NULL REFERENCES users(id),
    competition_id UUID NOT NULL REFERENCES competitions(id),

    status         VARCHAR(20) NOT NULL DEFAULT 'registered', -- registered | cancelled
    registered_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Natija: turga qarab har xil ({"place":1,"time":"00:23:11"} / {"score":42}).
    -- Hakam/admin kiritadi.
    result         JSONB,
    place          SMALLINT,

    reward_granted BOOLEAN NOT NULL DEFAULT FALSE
);

-- Bitta foydalanuvchi bitta musobaqaga bir marta. Bekor qilib qayta yozilsa
-- o'sha qator status='registered' ga qaytadi (yangi qator qo'shilmaydi).
CREATE UNIQUE INDEX idx_comp_reg_unique ON competition_registrations(user_id, competition_id);
CREATE INDEX idx_comp_reg_competition ON competition_registrations(competition_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS competition_registrations;
DROP TABLE IF EXISTS competitions;
-- +goose StatementEnd
