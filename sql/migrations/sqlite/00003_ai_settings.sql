-- +goose Up
ALTER TABLE site_settings ADD COLUMN ai_provider TEXT NOT NULL DEFAULT 'openai';
ALTER TABLE site_settings ADD COLUMN ai_base_url TEXT NOT NULL DEFAULT '';
ALTER TABLE site_settings ADD COLUMN ai_api_key_encrypted TEXT NOT NULL DEFAULT '';
ALTER TABLE site_settings ADD COLUMN ai_model_name TEXT NOT NULL DEFAULT '';
ALTER TABLE site_settings ADD COLUMN ai_proxy_url TEXT NOT NULL DEFAULT '';
ALTER TABLE site_settings ADD COLUMN ai_folder_id TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE site_settings DROP COLUMN ai_folder_id;
ALTER TABLE site_settings DROP COLUMN ai_proxy_url;
ALTER TABLE site_settings DROP COLUMN ai_model_name;
ALTER TABLE site_settings DROP COLUMN ai_api_key_encrypted;
ALTER TABLE site_settings DROP COLUMN ai_base_url;
ALTER TABLE site_settings DROP COLUMN ai_provider;
