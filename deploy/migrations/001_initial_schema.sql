CREATE TABLE IF NOT EXISTS content_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    deleted boolean,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    img_url TEXT,
    deleted BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS category_content_types (
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    content_type_id INTEGER REFERENCES content_types(id) ON DELETE CASCADE,
    PRIMARY KEY (category_id, content_type_id)
);

CREATE TABLE IF NOT EXISTS videos (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    img_url TEXT,
    deleted BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS video_categories (
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (video_id, category_id)
);

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    content_type_id INTEGER REFERENCES content_types(id) ON DELETE SET NULL,
    admin BOOLEAN,
    password TEXT,
    deleted BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS feedback (
    id SERIAL PRIMARY KEY,
    username TEXT,
    email TEXT,
    telegram TEXT,
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для ускорения запросов по связям между таблицами и ID
CREATE INDEX IF NOT EXISTS idx_video_categories_video_id ON video_categories(video_id);
CREATE INDEX IF NOT EXISTS idx_video_categories_category_id ON video_categories(category_id);
CREATE INDEX IF NOT EXISTS idx_category_content_types_category_id ON category_content_types(category_id);
CREATE INDEX IF NOT EXISTS idx_category_content_types_content_type_id ON category_content_types(content_type_id);

-- Индекс для ускорения поиска по username в таблице accounts
CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);

-- Индексы для фильтров по deleted
CREATE INDEX IF NOT EXISTS idx_videos_deleted ON videos(deleted);
CREATE INDEX IF NOT EXISTS idx_categories_deleted ON categories(deleted);
CREATE INDEX IF NOT EXISTS idx_content_types_deleted ON content_types(deleted);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted ON accounts(deleted);

-- Индексы для сортировки по created_at
CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at);
CREATE INDEX IF NOT EXISTS idx_categories_created_at ON categories(created_at);

-- Композитный индекс для ускорения запросов с фильтрацией по deleted и сортировкой по created_at
CREATE INDEX IF NOT EXISTS idx_videos_deleted_created_at ON videos(deleted, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_categories_deleted_created_at ON categories(deleted, created_at DESC);

-- Индексы по полям name для часто используемых таблиц для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_videos_name ON videos(name);
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
CREATE INDEX IF NOT EXISTS idx_content_types_name ON content_types(name);

-- Индекс для ускорения запросов, использующих JOIN между accounts и content_types
CREATE INDEX IF NOT EXISTS idx_accounts_content_type_id ON accounts(content_type_id);
