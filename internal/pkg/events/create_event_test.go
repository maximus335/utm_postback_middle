package events

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maximus335/utm_postback_middle/internal/app/utm_postback_middle/config"
	"github.com/maximus335/utm_postback_middle/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEventFromKafka(t *testing.T) {
	db, err := helpers.ConnectDB()
	require.NoError(t, err)

	ms := helpers.MockServer{
		Status: 200,
	}

	server, err := helpers.StartMockServer(&ms)
	require.NoError(t, err)

	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			require.NoError(t, err)
		}
	}()

	insertParams := map[string]string{
		"idfa_id":      "idfa_id",
		"os":           "android",
		"app_bundle":   "app_bundle",
		"appsflyer_id": "appsflyer_id",
		"mobile_phone": "mobile_phone_test",
		"eventName":    "eventName",
		"eventValue":   "eventValue",
	}

	var id int
	err = db.QueryRow(
		context.Background(),
		insertBorrowerSql,
		insertParams["idfa_id"],
		insertParams["appsflyer_id"],
		insertParams["app_bundle"],
		insertParams["os"],
		insertParams["mobile_phone"],
	).Scan(&id)
	require.NoError(t, err)

	err = insertBorrowerApplication(db, id, "test_uid")
	require.NoError(t, err)

	config, err := helpers.CreateTestConfig()
	require.NoError(t, err)

	err = CreateEventFromKafka(db, "active", "test_uid", &config.Appsflyer)
	require.NoError(t, err)

}

func TestCreateEvent(t *testing.T) {
	db, err := helpers.ConnectDB()

	require.NoError(t, err)

	ms := helpers.MockServer{
		Status: 200,
	}

	server, err := helpers.StartMockServer(&ms)
	require.NoError(t, err)

	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			require.NoError(t, err)
		}
	}()

	config, err := helpers.CreateTestConfig()

	require.NoError(t, err)

	withIdfaIdWithoutBorrower(t, db, &config.Appsflyer)
	withIdfaIdWitBorrower(t, db, &config.Appsflyer)
}

func withIdfaIdWithoutBorrower(t *testing.T, db *pgxpool.Pool, config *config.AppsflyerConfiguration) {
	defer cleanUp(db, t)

	params := map[string]interface{}{
		"idfa_id":      "idfa_id",
		"os":           "android",
		"app_bundle":   "app_bundle",
		"appsflyer_id": "appsflyer_id",
		"mobile_phone": "mobile_phone_test",
		"eventName":    "eventName",
		"eventValue":   "eventValue",
		"uid":          "uid",
	}

	err := CreateEvent(db, params, config)
	require.NoError(t, err)

	rows, err := db.Query(
		context.Background(),
		"SELECT mobile_phone from appsflyer_borrowers WHERE mobile_phone = $1",
		params["mobile_phone"],
	)

	require.NoError(t, err)

	mobilePhones := make([]string, 0)

	for rows.Next() {
		var mobilePhone string
		err = rows.Scan(&mobilePhone)
		require.NoError(t, err)
		mobilePhones = append(mobilePhones, mobilePhone)
	}

	assert.Equal(t, params["mobile_phone"], mobilePhones[0])

	checkBorrowerApplications(t, db)
}

func withIdfaIdWitBorrower(t *testing.T, db *pgxpool.Pool, config *config.AppsflyerConfiguration) {
	defer cleanUp(db, t)

	insertParams := map[string]string{
		"idfa_id":      "idfa_id",
		"os":           "android",
		"app_bundle":   "app_bundle",
		"appsflyer_id": "appsflyer_id",
		"mobile_phone": "mobile_phone_test",
		"eventName":    "eventName",
		"eventValue":   "eventValue",
	}

	insertBorrower(db, t, insertParams)

	params := map[string]interface{}{
		"idfa_id":      "idfa_id",
		"os":           "ios",
		"app_bundle":   "app_bundle",
		"appsflyer_id": "appsflyer_id",
		"mobile_phone": "mobile_phone_test",
		"eventName":    "eventName",
		"eventValue":   "eventValue",
		"uid":          "uid",
	}

	err := CreateEvent(db, params, config)
	require.NoError(t, err)

	rows, err := db.Query(
		context.Background(),
		"SELECT os from appsflyer_borrowers WHERE mobile_phone = $1",
		insertParams["mobile_phone"],
	)

	require.NoError(t, err)

	oses := make([]string, 0)

	for rows.Next() {
		var os string
		err = rows.Scan(&os)
		require.NoError(t, err)
		oses = append(oses, os)
	}

	assert.Equal(t, params["os"], oses[0])

	checkBorrowerApplications(t, db)
}

func checkBorrowerApplications(t *testing.T, db *pgxpool.Pool) {
	var cnt int

	err := db.QueryRow(
		context.Background(),
		"SELECT COUNT(id) from appsflyer_borrower_applications WHERE application_uid = $1",
		"uid",
	).Scan(&cnt)

	require.NoError(t, err)

	assert.Equal(t, cnt, 1)
}
