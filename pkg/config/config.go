package config

// Config is the global config object of kache
type config struct {
	Port            string
	Addr            string
	Api             bool
	CacheStrategy   string
	MaxCacheBytes   int64
	DefaultReplicas int
}

var Config *config

func init() {
	Config = &config{
		CacheStrategy:   "lru",
		MaxCacheBytes:   200,
		DefaultReplicas: 5,
	}
}
