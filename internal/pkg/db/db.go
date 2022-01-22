// Package db implements helpers for database management
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Options represents connect options
type Options struct {
	URL             string
	ConnMaxLifetime time.Duration
	MaxOpenConns    int32
}

// Connect connects to database, sets logger, and pool options
func Connect(opts *Options) (*pgxpool.Pool, error) {
	connConfig, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse database url: %w", err)
	}

	connConfig.ConnConfig.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
	}
	connConfig.ConnConfig.PreferSimpleProtocol = true
	connConfig.MaxConnLifetime = opts.ConnMaxLifetime
	connConfig.MaxConns = opts.MaxOpenConns

	db, err := pgxpool.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	//connStr := stdlib.RegisterConnConfig(connConfig)
	//
	//db, err := sqlx.Connect("pgx", connStr)
	//if err != nil {
	//	return nil, fmt.Errorf("cannot connect to database: %w", err)
	//}
	//
	//db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	//db.SetMaxIdleConns(opts.MaxIdleConns)
	//db.SetMaxOpenConns(opts.MaxOpenConns)

	return db, nil
}
