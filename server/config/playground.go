package config

import "time"

type PlaygroundConfig struct {
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
	ForceMigration             bool `default:"false"`
}

// GetPlaygroundConfig parses environment variables and returns a copy of the config
func GetPlaygroundConfig() PlaygroundConfig {
	var playgroundConf PlaygroundConfig
	getEnv("FLOW", &playgroundConf)
	return playgroundConf
}
