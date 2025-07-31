DO $$
DECLARE
    admin_type_id INTEGER;
    admin_username TEXT := current_setting('app.admin_username', TRUE);
    admin_password TEXT := current_setting('app.admin_password', TRUE);
    password_hash TEXT;
BEGIN
    -- Если переменные не заданы, используем значения по умолчанию
    IF admin_username IS NULL THEN
        admin_username := 'admin';
    END IF;

    IF admin_password IS NULL THEN
        admin_password := 'admin';
    END IF;

    -- Хешируем пароль (SHA-256)
    password_hash := encode(digest(admin_password, 'sha256'), 'hex');

    -- Добавляем тип контента admin, если он отсутствует
    SELECT id INTO admin_type_id FROM content_types WHERE name = 'admin';

    IF admin_type_id IS NULL THEN
        INSERT INTO content_types (name, deleted) VALUES ('admin', false) RETURNING id INTO admin_type_id;
    END IF;

    -- Добавляем учетную запись администратора, если она отсутствует
    IF NOT EXISTS (SELECT 1 FROM accounts WHERE username = admin_username) THEN
        INSERT INTO accounts (username, content_type_id, password, admin, deleted)
        VALUES (admin_username, admin_type_id, password_hash, true, false);
    END IF;
END $$;
