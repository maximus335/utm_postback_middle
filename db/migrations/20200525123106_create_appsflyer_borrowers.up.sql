BEGIN;
CREATE TABLE appsflyer_borrowers(
  id bigserial PRIMARY KEY,
  idfa_id varchar NOT NULL,
  appsflyer_id varchar NOT NULL,
  app_bundle varchar NOT NULL,
  mobile_phone varchar NOT NULL,
  os varchar NOT NULL
);
CREATE INDEX appsflyer_borrowers_mobile_phone_idx ON appsflyer_borrowers (mobile_phone);
COMMIT;