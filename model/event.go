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

package model

import (
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

func EventsFromFlow(flowEvents []flow.Event) ([]Event, error) {
	events := make([]Event, len(flowEvents))

	for i, event := range flowEvents {
		parsedEvent, err := parseEvent(event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse event")
		}
		events[i] = parsedEvent
	}

	return events, nil
}

func parseEvent(event flow.Event) (Event, error) {
	values := make([]string, len(event.Value.Fields))
	for j, field := range event.Value.Fields {
		encoded, _ := jsoncdc.Encode(field)
		values[j] = string(encoded)
	}

	return Event{
		Type:   event.Type,
		Values: values,
	}, nil
}
