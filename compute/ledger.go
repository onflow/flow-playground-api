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

package compute

import (
	"github.com/onflow/flow-go/engine/execution/state/delta"
	"github.com/onflow/flow-go/model/flow"
)

type Ledger map[string]flow.RegisterEntry

func (l Ledger) NewView() *delta.View {
	return delta.NewView(func(owner, key string) ([]byte, error) {
		id := flow.RegisterID{
			Owner: owner,
			Key:   key,
		}
		return l[id.String()].Value, nil
	})
}

func (l Ledger) ApplyDelta(delta delta.Delta) {
	for id, value := range delta.Data {
		l[id] = value
		// TODO: support delete
	}
}
