package config

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Load returns parsed test config.
func Load() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(GetPath())

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	return v, nil
}

// GetPath returns expanded path to test config.
func GetPath() string {
	_, filename, _, _ := runtime.Caller(0) // nolint:dogsled
	return strings.Replace(filepath.Dir(filename), "test/helpers/config", "configs/test.yml", 1)
}
