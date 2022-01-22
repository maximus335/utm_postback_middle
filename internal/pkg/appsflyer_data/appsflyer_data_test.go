package appsflyerdata

import (
	"testing"

	"github.com/maximus335/utm_postback_middle/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllAppsflyerData(t *testing.T) {
	db, err := helpers.ConnectDB()

	require.NoError(t, err)

	rawData := map[string]interface{}{
		"af_sub1":           "af_sub1",
		"af_sub2":           "af_sub2",
		"af_adset":          "af_adset",
		"campaign":          "campaign",
		"media_source":      "media_source",
		"af_ad":             "af_ad",
		"af_siteid":         "af_siteid",
		"af_channel":        "af_channel",
		"install_app_store": "install_app_store",
		"android_id":        "android_id",
		"idfa":              "idfa",
	}

	insertData(db, t, rawData)
	defer cleanUp(db, t)

	result, err := AllAppsflyerData(db, "advertising_id", "android_id", "idfa")
	require.NoError(t, err)

	assert.Equal(t, "af_sub1", result[0].AfSub1.String)
	assert.Equal(t, "af_sub2", result[0].AfSub2.String)
	assert.Equal(t, "af_ad", result[0].AfAd.String)
	assert.Equal(t, "campaign", result[0].Campaign.String)
	assert.Equal(t, "media_source", result[0].MediaSource.String)
	assert.Equal(t, "af_adset", result[0].AfAdset.String)
	assert.Equal(t, "af_siteid", result[0].AfSiteid.String)
	assert.Equal(t, "af_channel", result[0].AfChannel.String)
}

func TestRawAppsflyerData(t *testing.T) {
	db, err := helpers.ConnectDB()

	require.NoError(t, err)

	rawData := map[string]interface{}{
		"test2": "test2",
		"test1": "test1",
	}

	insertData(db, t, rawData)
	defer cleanUp(db, t)

	result, err := RawAppsflyerData(db, "advertising_id", "", "")
	require.NoError(t, err)

	assert.Equal(t, "test1", result[0]["test1"])
	assert.Equal(t, "test2", result[0]["test2"])
}
