package config

type config struct {
	Port          string
	Addr          string
	Api           bool
	CacheStrategy string
	MaxCacheBytes int64
}

var Config *config

func init() {
	Config = &config{
		CacheStrategy: "lru",
		MaxCacheBytes: 200,
	}
}
