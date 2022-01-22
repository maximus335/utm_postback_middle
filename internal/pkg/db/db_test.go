package db

import (
	"testing"

	"github.com/maximus335/utm_postback_middle/test/helpers/config"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)
	opts := Options{
		URL:             cfg.GetString("database.url"),
		ConnMaxLifetime: 0,
		MaxOpenConns:    1,
	}

	db, err := Connect(&opts)

	defer func() {
		db.Close()
	}()

	require.NoError(t, err)
}
