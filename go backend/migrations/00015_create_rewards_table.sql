-- +goose Up
-- +goose StatementBegin
-- FIT Coin do'koni: sovg'alar va ularni almashtirish.
--
-- Nega kerak: hozirgacha coin faqat YIG'ILARDI, sarflanmasdi — ya'ni
-- rag'batlantirish tizimining ikkinchi yarmi yo'q edi.
--
-- Kontent kodga yozilmaydi (CLAUDE.md §16): sovg'alarni admin panel yaratadi,
-- turga xos qo'shimcha maydonlar `config JSONB` da (o'lcham, rang, yetkazish
-- shartlari...) — yangi sovg'a turi migration talab qilmaydi.
CREATE TABLE rewards (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,

    title       VARCHAR(255) NOT NULL,
    description TEXT,
    image_url   TEXT,
    category    VARCHAR(50) NOT NULL DEFAULT 'other',

    -- Narx — musbat bo'lishi shart (tekin sovg'a coin tizimidan tashqarida).
    cost_coins  INTEGER NOT NULL CHECK (cost_coins > 0),

    -- stock: NULL — cheksiz (masalan raqamli sertifikat), 0 — tugagan.
    stock       INTEGER CHECK (stock IS NULL OR stock >= 0),

    -- per_user_limit: bitta foydalanuvchi necha marta ola oladi.
    -- NULL — cheklovsiz.
    per_user_limit INTEGER CHECK (per_user_limit IS NULL OR per_user_limit > 0),

    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    starts_at   TIMESTAMPTZ,
    ends_at     TIMESTAMPTZ,

    -- Kengayish uchun (§16.1): turga xos maydonlar.
    config      JSONB NOT NULL DEFAULT '{}'
);

-- Mobil do'kon ro'yxati: faqat aktiv sovg'alar, narx bo'yicha.
CREATE INDEX idx_rewards_active ON rewards(cost_coins)
    WHERE deleted_at IS NULL AND is_active = TRUE;

-- Almashtirish yozuvlari.
CREATE TABLE reward_redemptions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id     UUID NOT NULL REFERENCES users(id),
    reward_id   UUID NOT NULL REFERENCES rewards(id),

    -- Narx buyurtma paytida QOTIRILADI: admin keyin narxni o'zgartirsa
    -- allaqachon berilgan buyurtma va uning coin yozuvi mos qolishi kerak.
    cost_coins  INTEGER NOT NULL CHECK (cost_coins > 0),

    -- pending   — buyurtma qabul qilindi, hali topshirilmagan
    -- issued    — topshirildi
    -- cancelled — bekor qilindi (coin qaytarilgan)
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- code — foydalanuvchi topshirishda ko'rsatadigan qisqa kod.
    code        VARCHAR(16) NOT NULL,

    issued_at   TIMESTAMPTZ,
    issued_by   UUID REFERENCES users(id),
    note        TEXT
);

CREATE INDEX idx_redemptions_user ON reward_redemptions(user_id, created_at DESC);
-- Admin ro'yxati: kutayotgan buyurtmalar birinchi.
CREATE INDEX idx_redemptions_status ON reward_redemptions(status, created_at DESC);
-- Kod noyob: topshirishda kod bo'yicha qidiriladi.
CREATE UNIQUE INDEX idx_redemptions_code ON reward_redemptions(code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reward_redemptions;
DROP TABLE IF EXISTS rewards;
-- +goose StatementEnd
