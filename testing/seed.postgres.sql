-- Bruno fixtures for Postgres; mirrors the rows in data.sql. The schema comes
-- from Wakapi's own migrations, so only the data the API tests rely on is
-- inserted here (timestamps as timestamp, flags as boolean).
BEGIN;

INSERT INTO key_string_values (key, value)
VALUES ('imprint', 'no content here')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;

INSERT INTO users (id, api_key, email, location, password, created_at, last_logged_in_at,
                   share_data_max_days, share_editors, share_languages, share_projects, share_oss,
                   share_machines, is_admin, has_data, wakatime_api_key, reset_token, reports_weekly)
VALUES ('readuser', '33e7f538-0dce-4eba-8ffe-53db6814ed42', 'johndoe@example.org', 'Europe/Berlin',
        '$2a$10$93CAptdjLGRtc1D3xrZJcu8B/YBAPSjCZOHZRId.xpyrsLAeHOoA.',
        to_timestamp(1622205265.000) AT TIME ZONE 'UTC', to_timestamp(1622205274.178) AT TIME ZONE 'UTC',
        0, false, false, false, false, false, true, false, '', '', false),
       ('writeuser', 'f7aa255c-8647-4d0b-b90f-621c58fd580f', NULL, 'Europe/Berlin',
        '$2a$10$93CAptdjLGRtc1D3xrZJcu8B/YBAPSjCZOHZRId.xpyrsLAeHOoA.',
        to_timestamp(1622205296.000) AT TIME ZONE 'UTC', to_timestamp(1622205305.118) AT TIME ZONE 'UTC',
        7, false, false, true, false, false, false, true, '', '', false);

INSERT INTO api_keys (api_key, user_id, label, read_only)
VALUES ('1c91f670-2309-45fb-9d7e-738c766e85a6', 'writeuser', 'Full Access Key', false),
       ('774f7e16-b9a3-433e-ac68-7e28a82a50ca', 'writeuser', 'Read Only Key', true);

COMMIT;
