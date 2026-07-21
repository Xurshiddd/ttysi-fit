-- +goose Up
-- +goose StatementBegin
-- Guruhlar (HEMIS /data/group-list). Guruh fakultetga tegishli.
CREATE TABLE groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,

    hemis_id    BIGINT NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,

    -- Fakultet (group.department = Fakultet)
    faculty_id        UUID REFERENCES structures(id),
    faculty_hemis_id  BIGINT,

    specialty_code    VARCHAR(20),
    specialty_name    VARCHAR(500),
    education_lang    VARCHAR(50),

    active      BOOLEAN NOT NULL DEFAULT TRUE,
    synced_at   TIMESTAMPTZ
);

CREATE INDEX idx_groups_faculty_id ON groups(faculty_id);
CREATE INDEX idx_groups_faculty_hemis_id ON groups(faculty_hemis_id);
CREATE INDEX idx_groups_active ON groups(faculty_id) WHERE deleted_at IS NULL AND active = TRUE;

-- users ni HEMIS roster sync uchun kengaytirish
ALTER TABLE users ADD COLUMN hemis_id           BIGINT UNIQUE;       -- NULL lar ko'p bo'lishi mumkin (ro'yxatdan o'tganlar)
ALTER TABLE users ADD COLUMN hemis_login        VARCHAR(50);         -- student_id_number / employee_id_number
ALTER TABLE users ADD COLUMN group_id           UUID REFERENCES groups(id);
ALTER TABLE users ADD COLUMN department_id      UUID REFERENCES structures(id);  -- o'qituvchi kafedrasi
ALTER TABLE users ADD COLUMN faculty_hemis_id   BIGINT;   -- relink uchun
ALTER TABLE users ADD COLUMN department_hemis_id BIGINT;
ALTER TABLE users ADD COLUMN group_hemis_id     BIGINT;
ALTER TABLE users ADD COLUMN gender             VARCHAR(10);   -- male / female
ALTER TABLE users ADD COLUMN birth_date         DATE;
ALTER TABLE users ADD COLUMN position           VARCHAR(255);  -- xodim lavozimi
ALTER TABLE users ADD COLUMN specialty          VARCHAR(500);  -- talaba/xodim mutaxassisligi

-- Eski matn ustunlar (FK bilan almashtirildi)
ALTER TABLE users DROP COLUMN IF EXISTS department;
ALTER TABLE users DROP COLUMN IF EXISTS group_name;

-- Sinxron foydalanuvchilarda parol bo'lmaydi
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- email/hemis_login: bo'sh qiymatlar ko'p bo'lgani uchun UNIQUE o'rniga partial unique index.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX idx_users_email_unique ON users(email)
    WHERE email <> '' AND deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_hemis_login_unique ON users(hemis_login)
    WHERE hemis_login <> '' AND hemis_login IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX idx_users_group_id ON users(group_id);
CREATE INDEX idx_users_department_id ON users(department_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_hemis_login_unique;
DROP INDEX IF EXISTS idx_users_email_unique;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
ALTER TABLE users DROP COLUMN IF EXISTS specialty;
ALTER TABLE users DROP COLUMN IF EXISTS position;
ALTER TABLE users DROP COLUMN IF EXISTS birth_date;
ALTER TABLE users DROP COLUMN IF EXISTS gender;
ALTER TABLE users DROP COLUMN IF EXISTS group_hemis_id;
ALTER TABLE users DROP COLUMN IF EXISTS department_hemis_id;
ALTER TABLE users DROP COLUMN IF EXISTS faculty_hemis_id;
ALTER TABLE users DROP COLUMN IF EXISTS department_id;
ALTER TABLE users DROP COLUMN IF EXISTS group_id;
ALTER TABLE users DROP COLUMN IF EXISTS hemis_login;
ALTER TABLE users DROP COLUMN IF EXISTS hemis_id;

ALTER TABLE users ADD COLUMN department VARCHAR(255);
ALTER TABLE users ADD COLUMN group_name VARCHAR(100);

DROP TABLE IF EXISTS groups;
-- +goose StatementEnd
