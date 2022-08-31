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
