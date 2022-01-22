package appsflyerdata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

const insertQuerySql string = "INSERT INTO public.appsflyer_data (advertising_id, raw_json) VALUES($1,$2)"

func CreateAppsflyerData(db *pgxpool.Pool, params map[string]interface{}) error {
	rawJson, _ := json.Marshal(params)
	dataJSON := &pgtype.JSONB{Bytes: rawJson, Status: pgtype.Present}
	_, err := db.Exec(context.Background(), insertQuerySql, params["advertising_id"], dataJSON)
	if err != nil {
		return fmt.Errorf("Cannot insert data to appsflyer_data: %w", err)
	}
	return nil
}
