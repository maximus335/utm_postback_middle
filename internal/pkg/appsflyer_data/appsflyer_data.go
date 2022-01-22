package appsflyerdata

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maximus335/utm_postback_middle/internal/pkg/snt"
	"github.com/maximus335/utm_postback_middle/internal/pkg/types"
)

const sqlAppsflyerData string = `select raw_json->>'media_source' as media_source,
	raw_json->>'campaign' as campaign,
	raw_json->>'af_ad' as af_ad,
	raw_json->>'af_adset' as af_adset,
	raw_json->>'af_sub1' as af_sub1,
	raw_json->>'af_sub2' as af_sub2,
	raw_json->>'af_siteid' as af_siteid,
	raw_json->>'af_channel' as af_channel,
	raw_json->>'install_app_store' as install_app_store
	from public.appsflyer_data where %s`

const sqlRawData string = `select raw_json from public.appsflyer_data where %s`

type AppsflyerData struct {
	AfSub1          snt.NullString `db:"af_sub1" json:"af_sub1"`
	AfSub2          snt.NullString `db:"af_sub2" json:"af_sub2"`
	AfAdset         snt.NullString `db:"af_adset" json:"af_adset"`
	Campaign        snt.NullString `db:"campaign" json:"campaign"`
	MediaSource     snt.NullString `db:"media_source" json:"media_source"`
	AfAd            snt.NullString `db:"af_ad" json:"af_ad"`
	AfSiteid        snt.NullString `db:"af_siteid" json:"af_siteid"`
	AfChannel       snt.NullString `db:"af_channel" json:"af_channel"`
	InstallAppStore snt.NullString `db:"install_app_store" json:"install_app_store"`
}

type RawData struct {
	RawJson types.JsonbType `db:"raw_json"`
}

func AllAppsflyerData(db *pgxpool.Pool, advertisingId, androidId, idfa string) ([]AppsflyerData, error) {
	paramsMap := map[string]string{"advertising_id": advertisingId, "android_id": androidId, "idfa": idfa}
	ads := []AppsflyerData{}
	conds := addWhereConditions(paramsMap)
	query := fmt.Sprintf(sqlAppsflyerData, conds)
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("Cannot select appsflyer data: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var ad AppsflyerData
		err = rows.Scan(
			&ad.MediaSource,
			&ad.Campaign,
			&ad.AfAd,
			&ad.AfAdset,
			&ad.AfSub1,
			&ad.AfSub2,
			&ad.AfSiteid,
			&ad.AfChannel,
			&ad.InstallAppStore,
		)
		if err != nil {
			return nil, fmt.Errorf("Cannot scan appsflyer data: %w", err)
		}
		ads = append(ads, ad)
	}
	return ads, nil
}

func RawAppsflyerData(db *pgxpool.Pool, advertisingId, androidId, idfa string) ([]types.JsonbType, error) {
	paramsMap := map[string]string{"advertising_id": advertisingId, "android_id": androidId, "idfa": idfa}
	data := []types.JsonbType{}
	conds := addWhereConditions(paramsMap)
	query := fmt.Sprintf(sqlRawData, conds)
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("Cannot select appsflyer data: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var rawJson types.JsonbType
		err = rows.Scan(&rawJson)
		if err != nil {
			return nil, fmt.Errorf("Cannot scan appsflyer data: %w", err)
		}
		data = append(data, rawJson)
	}
	return data, nil
}

func addWhereConditions(c map[string]string) string {
	s := ""
	for k, v := range c {
		if len(v) > 0 {
			if k == "advertising_id" {
				s = s + k + "=" + "'" + v + "'" + " and "
			} else {
				s = s + "raw_json->>" + "'" + k + "'" + "=" + "'" + v + "'" + " and "
			}
		}
	}
	return strings.TrimSuffix(s, " and ")
}
