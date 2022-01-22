package helpers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maximus335/utm_postback_middle/internal/app/utm_postback_middle/config"
	"github.com/maximus335/utm_postback_middle/internal/pkg/db"
	configTest "github.com/maximus335/utm_postback_middle/test/helpers/config"
	"github.com/spf13/viper"
)

// ConnectDB connects to test database
func ConnectDB() (*pgxpool.Pool, error) {
	opts, err := LoadDBOptions()
	if err != nil {
		return nil, fmt.Errorf("cannot load DB config: %w", err)
	}

	return db.Connect(opts)
}

// LoadDBOptions builds db.Options from test config
func LoadDBOptions() (*db.Options, error) {
	v, err := configTest.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}

	opts := db.Options{
		URL:             v.GetString("database.Url"),
		ConnMaxLifetime: time.Duration(v.GetInt("database.ConnMaxLifetime")) * time.Second,
		MaxOpenConns:    v.GetInt32("database.MaxOpenConns"),
	}

	return &opts, nil
}

type MockServer struct {
	Status int
}

func (ms *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	switch ms.Status {
	case 200:
		w.WriteHeader(http.StatusOK)
	case 500:
		w.WriteHeader(http.StatusInternalServerError)
	case 404:
		w.WriteHeader(http.StatusNotFound)
	}
}

func StartMockServer(ms *MockServer) (*http.Server, error) {
	server := &http.Server{Addr: ":8099", Handler: ms}
	var err error
	go func() {
		err = server.ListenAndServe()
	}()

	return server, err
}

func CreateTestConfig() (*config.Configuration, error) {
	viper.SetConfigFile(configTest.GetPath())
	var configuration config.Configuration

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	err := viper.Unmarshal(&configuration)

	if err != nil {
		return nil, err
	}

	return &configuration, nil
}
