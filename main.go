package main

import (
	"errors"
	"fmt"
	"net/http"

	kache "github.com/falldio/Kache/pkg"
	"github.com/falldio/Kache/pkg/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *kache.Group {
	return kache.NewGroup("scores", 2<<10, kache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, g *kache.Group, r *gin.Engine) {
	peers := kache.NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("kache is running at ", addr)
	r.GET(kache.DefaultBasePath, peers.ServeHTTP)
}

func startAPIServer(g *kache.Group, r *gin.Engine) {
	r.GET("/api", func(c *gin.Context) {
		key := c.Query("key")
		view, err := g.Get(key)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New(""))
		}
		c.Data(http.StatusOK, "application/octet-stream", view.ByteSlice())
	})
}

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
	pflag.Parse()

	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	r := gin.Default()
	if config.Config.Api {
		startAPIServer(group, r)
	}
	addr := fmt.Sprintf("%s:%s", config.Config.Addr, config.Config.Port)
	startCacheServer(addr, addrs, group, r)
	r.Run(addr)
}
