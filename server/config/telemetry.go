package config

type TelemetryConfig struct {
	TracingEnabled bool `default:"true"`
}

var _ configGetter = &DatabaseConfig{}

func (c *TelemetryConfig) getConfig() {
	getEnv("TELEMETRY", c)
}
