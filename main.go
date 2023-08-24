package main

import (
	"fmt"

	"github.com/falldio/Kache/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("loading config file error: %s", err))
	}
	if err := viper.Unmarshal(config.Config); err != nil {
		log.Fatal(fmt.Errorf("unmarshaling conf failed, err: %s", err))
	}

	pflag.StringVarP(&config.Config.Port, "port", "p", "5658", "kache Port")
	pflag.BoolVarP(&config.Config.Api, "api", "a", true, "Start a api server?")
	pflag.StringVarP(&config.Config.CacheStrategy, "cache_strategy", "c", "lru", "Default cache strategy")
	pflag.Int64Var(&config.Config.MaxCacheBytes, "max_cache_bytes", 10, "Max byte size of the cache")
	pflag.IntVar(&config.Config.DefaultReplicas, "default_replicas", 5, "Replicas of the cache")
	pflag.Parse()
}
