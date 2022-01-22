BEGIN;
CREATE TABLE appsflyer_data(
  id bigserial PRIMARY KEY,
  advertising_id varchar,
  raw_json jsonb NOT NULL DEFAULT '{}'::jsonb,

  created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);
CREATE INDEX appsflyer_data_advertising_id_idx ON appsflyer_data (advertising_id);
COMMIT;


