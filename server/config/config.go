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
)

// config holds all parsed environment variables
var config struct {
	envParsed  bool
	platform   PlatformConfig
	playground PlaygroundConfig
	sentry     SentryConfig
	database   DatabaseConfig
}

func GetPlatform() Platform {
	if !config.envParsed {
		parseConfig()
	}
	return config.platform.Type
}

func GetPlayground() PlaygroundConfig {
	if !config.envParsed {
		parseConfig()
	}
	return config.playground
}

func GetSentry() SentryConfig {
	if !config.envParsed {
		parseConfig()
	}
	return config.sentry
}

func GetDatabase() DatabaseConfig {
	if !config.envParsed {
		parseConfig()
	}
	return config.database
}

// parseConfig parses all environment variables into config
func parseConfig() {
	config.platform = GetPlatformConfig()
	config.playground = GetPlaygroundConfig()
	config.sentry = getSentryConfig()
	config.database = getDatabaseConfig()
	config.envParsed = true
}

// getEnv parses environment variables into dest pointer
func getEnv(name string, dest interface{}) {
	if err := envconfig.Process(name, dest); err != nil {
		log.Fatal(err)
	}
}
