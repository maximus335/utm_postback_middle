package appsflyerdata

import (
	"context"
	"testing"

	"github.com/maximus335/utm_postback_middle/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAppsflyerData(t *testing.T) {
	db, err := helpers.ConnectDB()

	require.NoError(t, err)

	rawParams := map[string]interface{}{
		"advertising_id": "advertising_id",
		"af_sub1":        "af_sub1",
		"af_sub2":        "af_sub2",
		"af_ad":          "af_ad",
		"campaign":       "campaign",
		"media_source":   "media_source",
		"af_adset":       "af_adset",
		"af_siteid":      "af_siteid",
	}

	defer cleanUp(db, t)

	advertisingIds := make([]string, 0)

	err = CreateAppsflyerData(db, rawParams)

	require.NoError(t, err)

	rows, err := db.Query(context.Background(), "SELECT advertising_id from appsflyer_data WHERE advertising_id = $1", rawParams["advertising_id"])

	require.NoError(t, err)

	for rows.Next() {
		var advertisingId string
		err = rows.Scan(&advertisingId)
		require.NoError(t, err)
		advertisingIds = append(advertisingIds, advertisingId)
	}

	assert.Equal(t, rawParams["advertising_id"], advertisingIds[0])
}
