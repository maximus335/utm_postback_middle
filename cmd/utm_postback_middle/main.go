package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/maximus335/utm_postback_middle/internal/app/utm_postback_middle"
	"github.com/maximus335/utm_postback_middle/internal/app/utm_postback_middle/config"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

func init() {
	flag.StringVar(&cfgFile, "config-path", "configs/utm_postback_middle.yml", "path to config file")
}

func main() {
	flag.Parse()

	viper.SetConfigFile(cfgFile)
	var configuration config.Configuration

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("ERROR! Cannot read config file")
		os.Exit(1)
	}

	err := viper.Unmarshal(&configuration)

	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	s := utm_postback_middle.New(&configuration)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
