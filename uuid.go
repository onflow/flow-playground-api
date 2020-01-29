package flow_playground_api

import (
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func UnmarshalUUID(v interface{}) (id uuid.UUID, err error) {
	str, ok := v.(string)
	if !ok {
		return id, fmt.Errorf("ids must be strings")
	}

	err = id.UnmarshalText([]byte(str))
	if err != nil {
		return id, errors.Wrap(err, "failed to decode UUID")
	}

	return id, nil
}

func MarshalUUID(id uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		b, _ := id.MarshalText()
		w.Write(b)
	})
}
