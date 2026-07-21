-- +goose Up
-- +goose StatementBegin
-- Qurilma sessiyalari — "Mening qurilmalarim" va bir qurilma cheklovi.
--
-- Avval sessiya faqat Redis dagi `refresh:{user_id}` kalitida edi: ikkinchi
-- qurilmada kirilganda birinchisi JIMGINA ishdan chiqardi — foydalanuvchi
-- nima bo'lganini bilmasdi va qaysi qurilmalarda ochiq ekanini ko'ra
-- olmasdi. Bu jadval o'sha holatni ko'rinadigan va boshqariladigan qiladi.
CREATE TABLE user_sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id    UUID NOT NULL REFERENCES users(id),

    -- device_id — mijoz yasaydigan BARQAROR identifikator (ilova qayta
    -- o'rnatilmaguncha o'zgarmaydi). Shu bo'yicha "bu o'sha qurilmami?"
    -- degan savolga javob beriladi.
    device_id   VARCHAR(128) NOT NULL,
    device_name VARCHAR(255),                 -- "Samsung SM-S918B"
    platform    VARCHAR(20),                  -- android | ios | web
    app_version VARCHAR(32),

    -- Oxirgi kirish tafsilotlari (foydalanuvchi shubhali kirishni tanisin).
    ip          VARCHAR(45),                  -- IPv6 ham sig'adi
    user_agent  TEXT,

    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- NULL — sessiya faol.
    revoked_at     TIMESTAMPTZ,
    -- new_device | user | admin | logout
    revoked_reason VARCHAR(50)
);

-- "Mening qurilmalarim" ro'yxati.
CREATE INDEX idx_sessions_user ON user_sessions(user_id, last_seen_at DESC);

-- Bitta qurilmada bitta FAOL sessiya: qayta kirishda yangi qator emas,
-- o'shaning o'zi yangilanadi.
CREATE UNIQUE INDEX idx_sessions_active_device
    ON user_sessions(user_id, device_id)
    WHERE revoked_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_sessions;
-- +goose StatementEnd
