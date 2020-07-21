package model

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

func UnmarshalVersion(v interface{}) (version semver.Version, err error) {
	str, ok := v.(string)
	if !ok {
		return version, fmt.Errorf("versions must be strings")
	}

	err = json.Unmarshal([]byte(str), &version)
	if err != nil {
		return version, errors.Wrap(err, "failed to unmarshal versino")
	}

	return version, nil
}

func MarshalVersion(version semver.Version) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		enc := json.NewEncoder(w)
		_ = enc.Encode(&version)
	})
}
