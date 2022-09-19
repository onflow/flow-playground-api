package config

import (
	"github.com/kelseyhightower/envconfig"
	"log"
	"time"
)

var conf *Config = nil

type Config struct {
	Port                       int           `default:"8080"`
	Debug                      bool          `default:"false"`
	AllowedOrigins             []string      `default:"http://localhost:3000"`
	SessionAuthKey             string        `default:"428ce08c21b93e5f0eca24fbeb0c7673"`
	SessionMaxAge              time.Duration `default:"157680000s"`
	SessionCookiesSecure       bool          `default:"true"`
	SessionCookiesHTTPOnly     bool          `default:"true"`
	SessionCookiesSameSiteNone bool          `default:"false"`
	LedgerCacheSize            int           `default:"128"`
	PlaygroundBaseURL          string        `default:"http://localhost:3000"`
	StorageBackend             string
}

func GetConfig() Config {
	if conf == nil {
		conf = &Config{}
		if err := envconfig.Process("FLOW", conf); err != nil {
			log.Fatal(err)
		}
	}
	return *conf
}
