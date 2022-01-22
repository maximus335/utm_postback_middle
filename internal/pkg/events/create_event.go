package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maximus335/utm_postback_middle/internal/app/utm_postback_middle/config"
	"github.com/maximus335/utm_postback_middle/internal/pkg/types"
)

const insertBorrowerSql string = `INSERT INTO
    public.appsflyer_borrowers(idfa_id, appsflyer_id, app_bundle, os, mobile_phone)
    VALUES($1,$2,$3,$4,$5) RETURNING id`

const selectBorrowerSql string = `SELECT appsflyer_id, idfa_id, app_bundle, os
	FROM public.appsflyer_borrowers WHERE mobile_phone=$1`

const selectBorrowerIDSql string = `SELECT coalesce((SELECT id
    FROM public.appsflyer_borrowers WHERE mobile_phone=$1), 0)`

const updateBorrowerSql string = `UPDATE public.appsflyer_borrowers
    SET(idfa_id, appsflyer_id, app_bundle, os)=($1, $2, $3, $4)
    WHERE mobile_phone=$5`

const findBorrowerIDSql string = `SELECT appsflyer_borrower_id from public.appsflyer_borrower_applications
    WHERE application_uid=$1`

const findBorrowerSql string = `SELECT appsflyer_id, idfa_id, app_bundle, os
    FROM public.appsflyer_borrowers WHERE id=$1`

const insertBorrAppl string = `INSERT INTO
	public.appsflyer_borrower_applications(appsflyer_borrower_id, application_uid) VALUES($1,$2)`

const selectBorrAppsCountSql string = `SELECT COUNT(id)
    FROM public.appsflyer_borrower_applications WHERE application_uid=$1`

var fullRequiredParams = []string{"appsflyer_id", "idfa_id", "os", "app_bundle", "mobile_phone", "eventName", "uid"}
var shortRequiredParams = []string{"eventName", "mobile_phone"}
var eventMap = map[string]string{
	"active": "server_has_contract",
}

type appsflyerAndroidParams struct {
	AppsflyerId   string      `json:"appsflyer_id"`
	AdvertisingId string      `json:"advertising_id"`
	AfEventsApi   string      `json:"af_events_api"`
	EventValue    interface{} `json:"eventValue,omitempty"`
	EventName     string      `json:"eventName"`
}

type appsflyerIosParams struct {
	AppsflyerId string      `json:"appsflyer_id"`
	Idfa        string      `json:"idfa"`
	AfEventsApi string      `json:"af_events_api"`
	EventValue  interface{} `json:"eventValue"`
	EventName   string      `json:"eventName"`
}

type borrower struct {
	IdfaId      string
	AppsflyerId string
	AppBundle   string
	Os          string
}

func CreateEventFromKafka(db *pgxpool.Pool, event string, application_uid string, apiConfig *config.AppsflyerConfiguration) error {
	var appsflyer_borrower_id int
	err := db.QueryRow(context.Background(), findBorrowerIDSql, application_uid).Scan(&appsflyer_borrower_id)

	if err != nil {
		return &types.NotFoundError{Message: "Borrower appplication not found"}
	}

	borr, err := getBorrowerById(db, appsflyer_borrower_id)

	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"eventName": eventMap[event],
	}

	apiParams := generateApiParams(borr.AppsflyerId, borr.IdfaId, borr.Os, params)

	err = requestToAppsflyer(&apiParams, borr.AppBundle, apiConfig)

	if err != nil {
		return err
	}
	return nil
}

func CreateEvent(db *pgxpool.Pool, params map[string]interface{}, apiConfig *config.AppsflyerConfiguration) error {
	if _, ok := params["idfa_id"]; ok {
		err := checkParams(params, fullRequiredParams)
		if err != nil {
			return err
		}

		err = createBorrowerLink(db, params)
		if err != nil {
			return err
		}

		apiParams := generateApiParams(
			params["appsflyer_id"].(string),
			params["idfa_id"].(string),
			params["os"].(string),
			params,
		)

		err = requestToAppsflyer(&apiParams, params["app_bundle"].(string), apiConfig)
		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		err := checkParams(params, shortRequiredParams)
		if err != nil {
			return err
		}

		borr, err := getBorrower(db, params["mobile_phone"].(string))
		if err != nil {
			return err
		}

		apiParams := generateApiParams(borr.AppsflyerId, borr.IdfaId, borr.Os, params)

		err = requestToAppsflyer(&apiParams, borr.AppBundle, apiConfig)
		if err != nil {
			return err
		}
		return nil
	}
}

func checkParams(params map[string]interface{}, schema []string) error {
	fails := make([]string, 0)

	for _, param := range schema {
		if value, ok := params[param]; !ok || value == nil {
			fails = append(fails, param)
		}
	}

	if len(fails) > 0 {
		result := strings.Join(fails, ", ")
		return &types.ValidationParamsError{Message: "Must be exist: " + result}
	}

	return nil
}

func generateApiParams(appsflyerId string, idfaId string, os string, params map[string]interface{}) interface{} {
	var apiParams interface{}
	switch os {
	case "android":
		apiParams = appsflyerAndroidParams{
			AppsflyerId:   appsflyerId,
			AdvertisingId: idfaId,
			AfEventsApi:   "true",
			EventName:     params["eventName"].(string),
			EventValue:    params["eventValue"],
		}
	case "ios":
		apiParams = appsflyerIosParams{
			AppsflyerId: appsflyerId,
			Idfa:        idfaId,
			AfEventsApi: "true",
			EventName:   params["eventName"].(string),
			EventValue:  params["eventValue"],
		}
	}
	return apiParams
}

func getBorrowerById(db *pgxpool.Pool, id int) (*borrower, error) {
	var borr borrower
	err := db.QueryRow(context.Background(), findBorrowerSql, id).Scan(
		&borr.AppsflyerId,
		&borr.IdfaId,
		&borr.AppBundle,
		&borr.Os,
	)
	if err != nil {
		return nil, &types.NotFoundError{Message: "Borrower not found"}
	} else {
		return &borr, nil
	}
}

func getBorrower(db *pgxpool.Pool, mobile_phone string) (*borrower, error) {
	var borr borrower
	err := db.QueryRow(context.Background(), selectBorrowerSql, mobile_phone).Scan(
		&borr.AppsflyerId,
		&borr.IdfaId,
		&borr.AppBundle,
		&borr.Os,
	)
	if err != nil {
		return nil, &types.NotFoundError{Message: "Borrower not found"}
	} else {
		return &borr, nil
	}
}

func createBorrowerLink(db *pgxpool.Pool, params map[string]interface{}) error {
	var id int
	err := db.QueryRow(context.Background(), selectBorrowerIDSql, params["mobile_phone"].(string)).Scan(&id)

	if err != nil {
		return fmt.Errorf("Cannot select id appsflyer borrower: %w", err)
	}

	if id > 0 {
		_, err = db.Exec(
			context.Background(),
			updateBorrowerSql,
			params["idfa_id"],
			params["appsflyer_id"],
			params["app_bundle"],
			params["os"],
			params["mobile_phone"],
		)

		if err != nil {
			return fmt.Errorf("Cannot insert data to appsflyer_borrowers: %w", err)
		}

		err = insertBorrowerApplication(db, id, params["uid"].(string))

		if err != nil {
			return err
		}
	} else {
		err := db.QueryRow(
			context.Background(),
			insertBorrowerSql,
			params["idfa_id"],
			params["appsflyer_id"],
			params["app_bundle"],
			params["os"],
			params["mobile_phone"],
		).Scan(&id)

		if err != nil {
			return fmt.Errorf("Cannot insert data to appsflyer_borrowers: %w", err)
		}

		err = insertBorrowerApplication(db, id, params["uid"].(string))

		if err != nil {
			return err
		}
	}
	return nil
}

func insertBorrowerApplication(db *pgxpool.Pool, borrower_id int, uid string) error {
	var cnt int
	err := db.QueryRow(context.Background(), selectBorrAppsCountSql, uid).Scan(&cnt)
	if err != nil {
		return fmt.Errorf("Cannot select count of appsflyer_borrower_applications: %w", err)
	}

	if cnt > 0 {
		return nil
	}

	_, err = db.Exec(
		context.Background(),
		insertBorrAppl,
		borrower_id,
		uid,
	)

	if err != nil {
		return fmt.Errorf("Cannot insert data to appsflyer_borrower_applications: %w", err)
	}

	return nil
}

func requestToAppsflyer(params *interface{}, bundle string, apiConfig *config.AppsflyerConfiguration) error {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	url := apiConfig.Url + bundle
	jsonParams, _ := json.Marshal(params)
	body := bytes.NewBuffer(jsonParams)
	request, err := http.NewRequest("POST", url, body)
	request.Header.Set("Content-type", "application/json")
	request.Header.Set("authentication", apiConfig.ApiKey)
	if err != nil {
		return fmt.Errorf("Cant make request: %w", err)
	}

	resp, err := client.Do(request)

	if err != nil {
		return fmt.Errorf("Cant do request: %w", err)
	}

	if resp.StatusCode == 200 {
		return nil
	} else {
		return fmt.Errorf("Appsflyer bad response: code %v", resp.StatusCode)
	}
}
