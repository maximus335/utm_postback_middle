package events

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
)

func cleanUp(db *pgxpool.Pool, t *testing.T) {
	_, err := db.Exec(context.Background(), "DELETE FROM appsflyer_borrowers;")
	require.NoError(t, err)
	_, err = db.Exec(context.Background(), "DELETE FROM appsflyer_borrower_applications;")
	require.NoError(t, err)
}

func insertBorrower(db *pgxpool.Pool, t *testing.T, data map[string]string) {
	_, er := db.Exec(context.Background(),
		"INSERT INTO public.appsflyer_borrowers(idfa_id, appsflyer_id, app_bundle, os, mobile_phone) VALUES($1,$2,$3,$4,$5)",
		data["idfa_id"],
		data["appsflyer_id"],
		data["app_bundle"],
		data["os"],
		data["mobile_phone"],
	)
	require.NoError(t, er)
}
