/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
		w.Write([]byte("\""))
		w.Write(b)
		w.Write([]byte("\""))
	})
}
