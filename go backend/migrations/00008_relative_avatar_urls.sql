-- +goose Up
-- +goose StatementBegin
-- Avatar URL'lari absolyut host bilan saqlangan edi ("http://localhost:8080/static/avatars/1.jpg").
-- Host o'zgarganda (port almashishi, local → production) hamma yozuv buzilardi.
-- Endi DB da faqat nisbiy yo'l turadi, to'liq URL javob qaytarishda
-- MEDIA_PUBLIC_BASE_URL bilan yasaladi (handler.absoluteMediaURL).
--
-- Faqat o'zimiz yuklab olgan fayllar (/static/avatars/...) qisqartiriladi.
-- Tashqi URL'lar (HEMIS: /static/pi/...) o'zgarishsiz qoladi — ular boshqa
-- serverga tegishli va nisbiy bo'lsa ishlamaydi.
UPDATE users
SET avatar_url = regexp_replace(avatar_url, '^https?://[^/]+', '')
WHERE avatar_url ~ '^https?://[^/]+/static/avatars/';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Orqaga: nisbiy yo'llarga eski asosni qaytaramiz.
UPDATE users
SET avatar_url = 'http://localhost:8090' || avatar_url
WHERE avatar_url LIKE '/static/avatars/%';
-- +goose StatementEnd
