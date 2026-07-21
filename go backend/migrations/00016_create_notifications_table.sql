-- +goose Up
-- +goose StatementBegin
-- Bildirishnomalar — ILOVA ICHIDA (in-app).
--
-- Push (FCM) hozircha yo'q: u Firebase loyihasi va iOS uchun APNs kaliti
-- talab qiladi — bular tashkiliy qadamlar. Shu jadval push qo'shilganda
-- ham o'zgarishsiz qoladi: push faqat YETKAZISH usuli, xabarning o'zi
-- shu yerda saqlanadi (tarix, "o'qilgan" holati, admin nazorati).
CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Kimga. Har bir foydalanuvchiga ALOHIDA qator: "o'qilgan" holati
    -- shaxsiy, ommaviy e'lon ham har kimda o'z holatiga ega bo'lishi kerak.
    user_id    UUID NOT NULL REFERENCES users(id),

    -- Tur — mobil ilova ikonka/rangni shunga qarab tanlaydi va
    -- foydalanuvchi filtrlashi mumkin.
    type       VARCHAR(50) NOT NULL,

    title      VARCHAR(255) NOT NULL,
    body       TEXT,

    -- Havola: bosilganda qaysi ekran ochilsin (reward | achievement |
    -- challenge | competition | news). Bo'sh bo'lsa oddiy xabar.
    ref_type   VARCHAR(50),
    ref_id     UUID,

    -- NULL — o'qilmagan.
    read_at    TIMESTAMPTZ,

    -- Kengayish uchun (§4.1): masalan buyurtma kodi, coin miqdori.
    metadata   JSONB
);

-- Ro'yxat: yangi -> eski.
CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);

-- O'qilmaganlar soni (qo'ng'iroq nishoni) — juda tez-tez so'raladi,
-- shuning uchun partial indeks: faqat o'qilmaganlar indeksda turadi.
CREATE INDEX idx_notifications_unread ON notifications(user_id)
    WHERE read_at IS NULL;

-- IDEMPOTENTLIK: bitta manba uchun bitta xabar.
--
-- Yutuq baholash har faollik yozuvida ishlaydi va chellenj mukofoti
-- qayta so'ralishi mumkin — himoyasiz foydalanuvchi bir xil xabarni
-- o'nlab marta olardi. Partial: umumiy e'lonlarda ref_id yo'q va ular
-- cheklanmaydi.
CREATE UNIQUE INDEX idx_notifications_source_once
    ON notifications(user_id, type, ref_id)
    WHERE ref_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
