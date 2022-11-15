/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"github.com/kelseyhightower/envconfig"
	"log"
	"time"
)

// GetConfig parses environment variables and returns a copy of the config
func GetConfig() Config {
	if config == nil {
		config = &Config{}
		if err := envconfig.Process("FLOW", config); err != nil {
			log.Fatal(err)
		}
	}
	return *config
}

var config *Config = nil

// Config holds the environment variables for Playground
type Config struct {
	Platform                   Platform
	ForceMigration             bool          `default:"false"`
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
