-- +goose Up
-- SQL schema migration for SiteCraft (PostgreSQL dialect)

CREATE TABLE IF NOT EXISTS site_settings (
    id VARCHAR(36) PRIMARY KEY,
    site_name TEXT NOT NULL,
    tagline TEXT NOT NULL DEFAULT '',
    theme TEXT NOT NULL DEFAULT 'main',
    primary_color TEXT NOT NULL DEFAULT '',
    accent_color TEXT NOT NULL DEFAULT '',
    background_color TEXT NOT NULL DEFAULT '',
    font_family TEXT NOT NULL DEFAULT '',
    enable_a11y BOOLEAN NOT NULL DEFAULT TRUE,
    enable_alpine BOOLEAN NOT NULL DEFAULT FALSE,
    enable_turnstile BOOLEAN NOT NULL DEFAULT FALSE,
    turnstile_site_key TEXT NOT NULL DEFAULT '',
    turnstile_secret_key TEXT NOT NULL DEFAULT '',
    telegram_bot_token TEXT NOT NULL DEFAULT '',
    telegram_chat_id TEXT NOT NULL DEFAULT '',
    webhook_url TEXT NOT NULL DEFAULT '',
    yandex_metrika_id TEXT NOT NULL DEFAULT '',
    map_mode TEXT NOT NULL DEFAULT 'lazy',
    booking_mode TEXT NOT NULL DEFAULT 'slots',
    enable_payment BOOLEAN NOT NULL DEFAULT FALSE,
    payment_shop_id TEXT NOT NULL DEFAULT '',
    payment_secret TEXT NOT NULL DEFAULT '',
    payment_sandbox BOOLEAN NOT NULL DEFAULT TRUE,
    enable_blog BOOLEAN NOT NULL DEFAULT FALSE,
    enable_tma BOOLEAN NOT NULL DEFAULT FALSE,
    enable_tracker BOOLEAN NOT NULL DEFAULT FALSE,
    enable_i18n BOOLEAN NOT NULL DEFAULT TRUE,
    default_lang TEXT NOT NULL DEFAULT 'ru',
    enable_floating_contact BOOLEAN NOT NULL DEFAULT FALSE,
    whatsapp_phone TEXT NOT NULL DEFAULT '',
    telegram_contact TEXT NOT NULL DEFAULT '',
    enable_cart BOOLEAN NOT NULL DEFAULT FALSE,
    header_variant TEXT NOT NULL DEFAULT 'classic',
    footer_variant TEXT NOT NULL DEFAULT 'classic',
    enable_live_chat BOOLEAN NOT NULL DEFAULT FALSE,
    enable_fleet BOOLEAN NOT NULL DEFAULT FALSE,
    enable_recent_activity BOOLEAN NOT NULL DEFAULT FALSE,
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
    enable_careers BOOLEAN NOT NULL DEFAULT TRUE,
    enable_cookie_consent BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS site_content (
    id VARCHAR(36) PRIMARY KEY,
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pages (
    id VARCHAR(36) PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL DEFAULT '',
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT '',
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS page_blocks (
    id VARCHAR(36) PRIMARY KEY,
    page_id VARCHAR(36) NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    order_index INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    payload TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_page_blocks_order_pg ON page_blocks (order_index);

CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(36) PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price TEXT NOT NULL DEFAULT '',
    order_index INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    stock_count INTEGER NOT NULL DEFAULT 0,
    is_preorder BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_services_order_pg ON services (order_index);

CREATE TABLE IF NOT EXISTS leads (
    id VARCHAR(36) PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    booking_date TEXT NOT NULL DEFAULT '',
    booking_time TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'новая',
    payload TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_leads_created_pg ON leads (created_at DESC);

CREATE TABLE IF NOT EXISTS articles (
    id VARCHAR(36) PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    excerpt TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    cover_image TEXT NOT NULL DEFAULT '',
    author_name TEXT NOT NULL DEFAULT '',
    published_at TEXT NOT NULL DEFAULT '',
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TEXT NOT NULL DEFAULT '',
    views_count INTEGER NOT NULL DEFAULT 0,
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_articles_slug_pg ON articles (slug);
CREATE INDEX IF NOT EXISTS idx_articles_published_pg ON articles (is_published, published_at DESC);

CREATE TABLE IF NOT EXISTS service_orders (
    id VARCHAR(36) PRIMARY KEY,
    order_code TEXT NOT NULL UNIQUE,
    client_name TEXT NOT NULL DEFAULT '',
    client_phone TEXT NOT NULL DEFAULT '',
    item_title TEXT NOT NULL,
    current_stage TEXT NOT NULL DEFAULT 'received',
    total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    stages_history TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_orders_code_pg ON service_orders (order_code);

CREATE TABLE IF NOT EXISTS coupons (
    id VARCHAR(36) PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    discount_type TEXT NOT NULL DEFAULT 'percent',
    discount_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    min_order_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    valid_until TEXT NOT NULL DEFAULT '',
    usage_limit INTEGER NOT NULL DEFAULT 0,
    used_count INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_coupons_code_pg ON coupons (code);

CREATE TABLE IF NOT EXISTS resources (
    id VARCHAR(36) PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    avatar TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    order_index INTEGER NOT NULL DEFAULT 0,
    service_ids TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS branches (
    id VARCHAR(36) PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    work_hours TEXT NOT NULL DEFAULT '',
    lat DOUBLE PRECISION NOT NULL DEFAULT 0,
    lng DOUBLE PRECISION NOT NULL DEFAULT 0,
    metro TEXT NOT NULL DEFAULT '',
    services TEXT NOT NULL DEFAULT '',
    is_main BOOLEAN NOT NULL DEFAULT FALSE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reviews (
    id VARCHAR(36) PRIMARY KEY,
    author_name TEXT NOT NULL,
    rating INTEGER NOT NULL DEFAULT 5,
    comment TEXT NOT NULL,
    order_id TEXT NOT NULL DEFAULT '',
    photo TEXT NOT NULL DEFAULT '',
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    source TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS vacancies (
    id VARCHAR(36) PRIMARY KEY,
    title TEXT NOT NULL,
    department TEXT NOT NULL DEFAULT '',
    salary TEXT NOT NULL DEFAULT '',
    experience TEXT NOT NULL DEFAULT '',
    schedule TEXT NOT NULL DEFAULT '',
    responsibilities TEXT NOT NULL DEFAULT '[]',
    requirements TEXT NOT NULL DEFAULT '[]',
    benefits TEXT NOT NULL DEFAULT '[]',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id VARCHAR(36) PRIMARY KEY,
    session_id TEXT NOT NULL,
    sender TEXT NOT NULL,
    message TEXT NOT NULL,
    tg_msg_id INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_chat_session_pg ON chat_messages (session_id, created_at ASC);

-- SCS Sessions store table for Postgres
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS sessions_expiry_idx_pg ON sessions (expiry);

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
