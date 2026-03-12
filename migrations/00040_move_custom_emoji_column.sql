-- +goose Up
-- +goose StatementBegin
UPDATE users
SET emoji = '<tg-emoji emoji-id="' || custom_emoji_id || '">' || emoji || '</tg-emoji>'
WHERE custom_emoji_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE users
SET emoji = regexp_replace(emoji, '^<tg-emoji[^>]*>(.*)</tg-emoji>$', '\1')
WHERE emoji ~ '^<tg-emoji';
-- +goose StatementEnd