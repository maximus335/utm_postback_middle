package appsflyerdata

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
)

func cleanUp(db *pgxpool.Pool, t *testing.T) {
	_, err := db.Exec(context.Background(), "DELETE FROM appsflyer_data;")
	require.NoError(t, err)
}

func insertData(db *pgxpool.Pool, t *testing.T, d map[string]interface{}) {
	rawJson, _ := json.Marshal(d)
	dataJSON := &pgtype.JSONB{Bytes: rawJson, Status: pgtype.Present}

	_, er := db.Exec(context.Background(),
		"INSERT INTO public.appsflyer_data (advertising_id, raw_json) VALUES($1, $2)",
		"advertising_id",
		dataJSON,
	)
	require.NoError(t, er)
}
