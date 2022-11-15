package config

type Platform string

const (
	Local      Platform = "LOCAL"
	Staging    Platform = "STAGING"
	Production Platform = "PRODUCTION"
)
