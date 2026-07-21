-- +goose Up
-- +goose StatementBegin
-- FIT Coin — LEDGER modeli (CLAUDE.md §4.3: "fit_coins — tranzaksiyalar (ledger model)").
--
-- Nega ledger: balans ustuni emas, harakatlar ro'yxati. Har bir o'zgarish —
-- alohida qator, hech qachon tahrirlanmaydi/o'chirilmaydi. Balans = SUM(amount).
-- Shu sababli "bu coin qayerdan keldi?" degan savolga har doim javob bor va
-- balansni bir joyda yangilashda yuzaga keladigan poyga (race) yo'q.
CREATE TABLE fit_coins (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id    UUID NOT NULL REFERENCES users(id),

    -- amount: musbat — tushum, manfiy — chiqim. Nol ma'nosiz (CHECK).
    amount     INTEGER NOT NULL CHECK (amount <> 0),

    -- reason: nima uchun berildi (challenge_reward | competition_reward |
    -- admin_grant | admin_revoke | purchase ...). Kodda konstanta.
    reason     VARCHAR(50) NOT NULL,

    -- Manba havolasi: qaysi chellenj/musobaqa uchun.
    ref_type   VARCHAR(50),
    ref_id     UUID,

    note       TEXT,
    metadata   JSONB
);

-- Balans va tarix bo'yicha o'qish.
CREATE INDEX idx_fit_coins_user ON fit_coins(user_id, created_at DESC);

-- IDEMPOTENTLIK: bitta manba uchun bitta mukofot.
-- Chellenj mukofoti ikki marta berilib qolmasligi kerak (takroriy so'rov,
-- qayta urinish, parallel chaqiruv). Partial: qo'lda berilgan coin'larda
-- ref_id yo'q va ular cheklanmaydi.
CREATE UNIQUE INDEX idx_fit_coins_source_once
    ON fit_coins(user_id, reason, ref_id)
    WHERE ref_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS fit_coins;
-- +goose StatementEnd
