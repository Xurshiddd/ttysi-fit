-- +goose Up
-- +goose StatementBegin
-- Yangiliklar (CLAUDE.md §16.3: "Yangiliklar (news) — admin kontent kiritadi").
--
-- Chellenj/musobaqadan farqli: bu yerda TUR REGISTRI yo'q va bo'lishi ham
-- shart emas. Yangilikning turga qarab o'zgaradigan parametrlari yo'q —
-- u oddiy kontent (sarlavha + matn). JSONB config qo'shish keraksiz
-- murakkablik bo'lardi. Kengayish uchun `metadata` qoldirilgan (§4.1).
CREATE TABLE news (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    title        VARCHAR(255) NOT NULL,
    -- excerpt: ro'yxatda ko'rsatiladigan qisqa matn. Bo'sh bo'lsa mijoz
    -- body'ning boshini oladi.
    excerpt      TEXT,
    body         TEXT NOT NULL,

    cover_url    TEXT,
    status       VARCHAR(20) NOT NULL DEFAULT 'draft',   -- draft | published
    -- published_at: e'lon vaqti. Kelajakdagi sana — rejalashtirilgan e'lon
    -- (mobil ro'yxat NOW() dan keyingilarni ko'rsatmaydi).
    published_at TIMESTAMPTZ,

    author_id    UUID REFERENCES users(id),
    -- pinned: muhim yangilik ro'yxat boshida turadi.
    pinned       BOOLEAN NOT NULL DEFAULT FALSE,
    views        INTEGER NOT NULL DEFAULT 0,

    metadata     JSONB
);

-- Mobil ro'yxat shu bo'yicha so'raydi: e'lon qilinganlar, yangi -> eski,
-- pinned yuqorida.
CREATE INDEX idx_news_published ON news(status, pinned DESC, published_at DESC)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS news;
-- +goose StatementEnd
