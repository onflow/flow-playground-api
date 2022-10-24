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

package adapter

// TODO: Do we still need these?
/*
type accountState struct {
	Private map[string]any
	Public  map[string]any
	Storage map[string]any
}

// stateToAPI removes any state values that are blockchain system values and not relevant to user usage of the playground.
func stateToAPI(state string) string {
	if state == "" {
		return "{}"
	}

	var accState accountState
	_ = json.Unmarshal([]byte(state), &accState) // state will always be valid JSON

	delete(accState.Public, "flowTokenBalance")
	delete(accState.Public, "flowTokenReceiver")
	delete(accState.Storage, "flowTokenVault")

	// return empty as empty object, no keys - v0 adapter
	if len(accState.Public)+len(accState.Private)+len(accState.Storage) == 0 {
		return "{}"
	}

	adaptedState, _ := json.Marshal(accState)
	return string(adaptedState)
}
*/
