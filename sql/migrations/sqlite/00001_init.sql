-- +goose Up
-- SQL schema migration for SiteCraft (SQLite dialect)

CREATE TABLE IF NOT EXISTS site_settings (
    id TEXT PRIMARY KEY,
    site_name TEXT NOT NULL,
    tagline TEXT NOT NULL DEFAULT '',
    theme TEXT NOT NULL DEFAULT 'main',
    primary_color TEXT NOT NULL DEFAULT '',
    accent_color TEXT NOT NULL DEFAULT '',
    background_color TEXT NOT NULL DEFAULT '',
    font_family TEXT NOT NULL DEFAULT '',
    enable_a11y INTEGER NOT NULL DEFAULT 1,
    enable_alpine INTEGER NOT NULL DEFAULT 0,
    enable_turnstile INTEGER NOT NULL DEFAULT 0,
    turnstile_site_key TEXT NOT NULL DEFAULT '',
    turnstile_secret_key TEXT NOT NULL DEFAULT '',
    telegram_bot_token TEXT NOT NULL DEFAULT '',
    telegram_chat_id TEXT NOT NULL DEFAULT '',
    webhook_url TEXT NOT NULL DEFAULT '',
    yandex_metrika_id TEXT NOT NULL DEFAULT '',
    map_mode TEXT NOT NULL DEFAULT 'lazy',
    booking_mode TEXT NOT NULL DEFAULT 'slots',
    enable_payment INTEGER NOT NULL DEFAULT 0,
    payment_shop_id TEXT NOT NULL DEFAULT '',
    payment_secret TEXT NOT NULL DEFAULT '',
    payment_sandbox INTEGER NOT NULL DEFAULT 1,
    enable_blog INTEGER NOT NULL DEFAULT 0,
    enable_tma INTEGER NOT NULL DEFAULT 0,
    enable_tracker INTEGER NOT NULL DEFAULT 0,
    enable_i18n INTEGER NOT NULL DEFAULT 1,
    default_lang TEXT NOT NULL DEFAULT 'ru',
    enable_floating_contact INTEGER NOT NULL DEFAULT 0,
    whatsapp_phone TEXT NOT NULL DEFAULT '',
    telegram_contact TEXT NOT NULL DEFAULT '',
    enable_cart INTEGER NOT NULL DEFAULT 0,
    header_variant TEXT NOT NULL DEFAULT 'classic',
    footer_variant TEXT NOT NULL DEFAULT 'classic',
    enable_live_chat INTEGER NOT NULL DEFAULT 0,
    enable_fleet INTEGER NOT NULL DEFAULT 0,
    enable_recent_activity INTEGER NOT NULL DEFAULT 0,
    client_admin_password TEXT NOT NULL DEFAULT '',
    client_admin_password_hash TEXT NOT NULL DEFAULT '',
    legal_entity TEXT NOT NULL DEFAULT '',
    legal_address TEXT NOT NULL DEFAULT '',
    inn TEXT NOT NULL DEFAULT '',
    kpp TEXT NOT NULL DEFAULT '',
    ogrn TEXT NOT NULL DEFAULT '',
    bank_name TEXT NOT NULL DEFAULT '',
    bank_bik TEXT NOT NULL DEFAULT '',
    bank_account TEXT NOT NULL DEFAULT '',
    bank_corr TEXT NOT NULL DEFAULT '',
    contact_email TEXT NOT NULL DEFAULT '',
    enable_careers INTEGER NOT NULL DEFAULT 1,
    enable_cookie_consent INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS site_content (
    id TEXT PRIMARY KEY,
    hero_badge TEXT NOT NULL DEFAULT '',
    hero_title TEXT NOT NULL DEFAULT '',
    hero_subtitle TEXT NOT NULL DEFAULT '',
    hero_title_b TEXT NOT NULL DEFAULT '',
    hero_subtitle_b TEXT NOT NULL DEFAULT '',
    hero_cta_b TEXT NOT NULL DEFAULT '',
    owner_name TEXT NOT NULL DEFAULT '',
    owner_title TEXT NOT NULL DEFAULT '',
    phone1 TEXT NOT NULL DEFAULT '',
    phone2 TEXT NOT NULL DEFAULT '',
    work_hours TEXT NOT NULL DEFAULT '',
    lunch_break TEXT NOT NULL DEFAULT '',
    days_off TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    address_sub TEXT NOT NULL DEFAULT '',
    map_iframe TEXT NOT NULL DEFAULT '',
    photo TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS pages (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL DEFAULT '',
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT '',
    is_published INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS page_blocks (
    id TEXT PRIMARY KEY,
    page_id TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    order_index INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    payload TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_page_blocks_order ON page_blocks (order_index);

CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price TEXT NOT NULL DEFAULT '',
    order_index INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    stock_count INTEGER NOT NULL DEFAULT 0,
    is_preorder INTEGER NOT NULL DEFAULT 0,
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_services_order ON services (order_index);

CREATE TABLE IF NOT EXISTS leads (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    booking_date TEXT NOT NULL DEFAULT '',
    booking_time TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'новая',
    payload TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_leads_created ON leads (created_at DESC);

CREATE TABLE IF NOT EXISTS articles (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    excerpt TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    cover_image TEXT NOT NULL DEFAULT '',
    author_name TEXT NOT NULL DEFAULT '',
    published_at TEXT NOT NULL DEFAULT '',
    is_published INTEGER NOT NULL DEFAULT 1,
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at TEXT NOT NULL DEFAULT '',
    views_count INTEGER NOT NULL DEFAULT 0,
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles (slug);
CREATE INDEX IF NOT EXISTS idx_articles_published ON articles (is_published, published_at DESC);

CREATE TABLE IF NOT EXISTS service_orders (
    id TEXT PRIMARY KEY,
    order_code TEXT NOT NULL UNIQUE,
    client_name TEXT NOT NULL DEFAULT '',
    client_phone TEXT NOT NULL DEFAULT '',
    item_title TEXT NOT NULL,
    current_stage TEXT NOT NULL DEFAULT 'received',
    total_amount REAL NOT NULL DEFAULT 0,
    is_paid INTEGER NOT NULL DEFAULT 0,
    stages_history TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_orders_code ON service_orders (order_code);

CREATE TABLE IF NOT EXISTS coupons (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    discount_type TEXT NOT NULL DEFAULT 'percent',
    discount_value REAL NOT NULL DEFAULT 0,
    min_order_amount REAL NOT NULL DEFAULT 0,
    valid_until TEXT NOT NULL DEFAULT '',
    usage_limit INTEGER NOT NULL DEFAULT 0,
    used_count INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons (code);

CREATE TABLE IF NOT EXISTS resources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    avatar TEXT NOT NULL DEFAULT '',
    active INTEGER NOT NULL DEFAULT 1,
    order_index INTEGER NOT NULL DEFAULT 0,
    service_ids TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS branches (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    work_hours TEXT NOT NULL DEFAULT '',
    lat REAL NOT NULL DEFAULT 0,
    lng REAL NOT NULL DEFAULT 0,
    metro TEXT NOT NULL DEFAULT '',
    services TEXT NOT NULL DEFAULT '',
    is_main INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS reviews (
    id TEXT PRIMARY KEY,
    author_name TEXT NOT NULL,
    rating INTEGER NOT NULL DEFAULT 5,
    comment TEXT NOT NULL,
    order_id TEXT NOT NULL DEFAULT '',
    photo TEXT NOT NULL DEFAULT '',
    is_published INTEGER NOT NULL DEFAULT 1,
    source TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS vacancies (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    department TEXT NOT NULL DEFAULT '',
    salary TEXT NOT NULL DEFAULT '',
    experience TEXT NOT NULL DEFAULT '',
    schedule TEXT NOT NULL DEFAULT '',
    responsibilities TEXT NOT NULL DEFAULT '[]',
    requirements TEXT NOT NULL DEFAULT '[]',
    benefits TEXT NOT NULL DEFAULT '[]',
    active INTEGER NOT NULL DEFAULT 1,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    sender TEXT NOT NULL,
    message TEXT NOT NULL,
    tg_msg_id INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_chat_session ON chat_messages (session_id, created_at ASC);

-- SCS Sessions store table for SQLite
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BLOB NOT NULL,
    expiry REAL NOT NULL
);
CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry);

-- +goose Down
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS vacancies;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS service_orders;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS leads;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS page_blocks;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS site_content;
DROP TABLE IF EXISTS site_settings;
