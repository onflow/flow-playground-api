package config

type DatabaseConfig struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     int
}

func getDatabaseConfig() DatabaseConfig {
	var databaseConfig DatabaseConfig
	getEnv("FLOW_DB", &databaseConfig)
	return databaseConfig
}
