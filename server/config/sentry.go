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

type SentryConfig struct {
	Dsn              string `default:"https://e8ff473e48aa4962b1a518411489ec5d@o114654.ingest.sentry.io/6398442"`
	Debug            bool   `default:"true"`
	AttachStacktrace bool   `default:"true"`
}

func getSentryConfig() SentryConfig {
	var sentryConfig SentryConfig
	getEnv("SENTRY", &sentryConfig)
	return sentryConfig
}
