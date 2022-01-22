BEGIN;
CREATE TABLE appsflyer_borrower_applications(
  id bigserial PRIMARY KEY,
  application_uid varchar NOT NULL,
  appsflyer_borrower_id bigint REFERENCES appsflyer_borrowers(id) ON DELETE CASCADE
);
CREATE INDEX appsflyer_borrower_applications_uid_idx ON appsflyer_borrower_applications (application_uid);
COMMIT;