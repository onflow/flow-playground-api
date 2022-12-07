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

// TelemetryConfig defines tracing configuration
type TelemetryConfig struct {
	// TracingEnabled determines whether to collect and export traces
	TracingEnabled bool `default:"false"`
	// TracingCollectorEndpoint is the OTEL collector endpoint to which traces should be sent
	TracingCollectorEndpoint string
}

var _ configGetter = &TelemetryConfig{}

func (c *TelemetryConfig) getConfig() {
	getEnv("TELEMETRY", c)
}
