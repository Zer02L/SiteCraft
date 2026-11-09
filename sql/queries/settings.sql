-- name: GetSiteSettings :one
SELECT * FROM site_settings
LIMIT 1;

-- name: UpsertSiteSettings :one
INSERT INTO site_settings (
    id, site_name, tagline, theme, primary_color, accent_color, background_color, font_family,
    enable_a11y, enable_alpine, enable_turnstile, turnstile_site_key, turnstile_secret_key,
    telegram_bot_token, telegram_chat_id, webhook_url, yandex_metrika_id, map_mode, booking_mode,
    enable_payment, payment_shop_id, payment_secret, payment_sandbox,
    enable_blog, enable_tma, enable_tracker, enable_i18n, default_lang,
    enable_floating_contact, whatsapp_phone, telegram_contact, enable_cart,
    header_variant, footer_variant, enable_live_chat, enable_fleet, enable_recent_activity,
    client_admin_password, client_admin_password_hash,
    legal_entity, legal_address, inn, kpp, ogrn,
    bank_name, bank_bik, bank_account, bank_corr, contact_email,
    enable_careers, enable_cookie_consent, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO UPDATE SET
    site_name = excluded.site_name,
    tagline = excluded.tagline,
    theme = excluded.theme,
    primary_color = excluded.primary_color,
    accent_color = excluded.accent_color,
    background_color = excluded.background_color,
    font_family = excluded.font_family,
    enable_a11y = excluded.enable_a11y,
    enable_alpine = excluded.enable_alpine,
    enable_turnstile = excluded.enable_turnstile,
    turnstile_site_key = excluded.turnstile_site_key,
    turnstile_secret_key = excluded.turnstile_secret_key,
    telegram_bot_token = excluded.telegram_bot_token,
    telegram_chat_id = excluded.telegram_chat_id,
    webhook_url = excluded.webhook_url,
    yandex_metrika_id = excluded.yandex_metrika_id,
    map_mode = excluded.map_mode,
    booking_mode = excluded.booking_mode,
    enable_payment = excluded.enable_payment,
    payment_shop_id = excluded.payment_shop_id,
    payment_secret = excluded.payment_secret,
    payment_sandbox = excluded.payment_sandbox,
    enable_blog = excluded.enable_blog,
    enable_tma = excluded.enable_tma,
    enable_tracker = excluded.enable_tracker,
    enable_i18n = excluded.enable_i18n,
    default_lang = excluded.default_lang,
    enable_floating_contact = excluded.enable_floating_contact,
    whatsapp_phone = excluded.whatsapp_phone,
    telegram_contact = excluded.telegram_contact,
    enable_cart = excluded.enable_cart,
    header_variant = excluded.header_variant,
    footer_variant = excluded.footer_variant,
    enable_live_chat = excluded.enable_live_chat,
    enable_fleet = excluded.enable_fleet,
    enable_recent_activity = excluded.enable_recent_activity,
    client_admin_password = excluded.client_admin_password,
    client_admin_password_hash = excluded.client_admin_password_hash,
    legal_entity = excluded.legal_entity,
    legal_address = excluded.legal_address,
    inn = excluded.inn,
    kpp = excluded.kpp,
    ogrn = excluded.ogrn,
    bank_name = excluded.bank_name,
    bank_bik = excluded.bank_bik,
    bank_account = excluded.bank_account,
    bank_corr = excluded.bank_corr,
    contact_email = excluded.contact_email,
    enable_careers = excluded.enable_careers,
    enable_cookie_consent = excluded.enable_cookie_consent,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetSiteContent :one
SELECT * FROM site_content
LIMIT 1;

-- name: UpsertSiteContent :one
INSERT INTO site_content (
    id, hero_badge, hero_title, hero_subtitle, hero_title_b, hero_subtitle_b, hero_cta_b,
    owner_name, owner_title, phone1, phone2, work_hours, lunch_break, days_off,
    address, address_sub, map_iframe, photo, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO UPDATE SET
    hero_badge = excluded.hero_badge,
    hero_title = excluded.hero_title,
    hero_subtitle = excluded.hero_subtitle,
    hero_title_b = excluded.hero_title_b,
    hero_subtitle_b = excluded.hero_subtitle_b,
    hero_cta_b = excluded.hero_cta_b,
    owner_name = excluded.owner_name,
    owner_title = excluded.owner_title,
    phone1 = excluded.phone1,
    phone2 = excluded.phone2,
    work_hours = excluded.work_hours,
    lunch_break = excluded.lunch_break,
    days_off = excluded.days_off,
    address = excluded.address,
    address_sub = excluded.address_sub,
    map_iframe = excluded.map_iframe,
    photo = excluded.photo,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;
