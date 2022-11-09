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

package blockchain

import (
	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
)

// newEmulatorCache returns a new instance of cache with provided capacity.
func newEmulatorCache(capacity int) *emulatorCache {
	return &emulatorCache{
		cache: lru.New(capacity),
	}
}

// emulatorCache caches the emulator state.
//
// In the environment where multiple replicas maintain it's own cache copy it can get into multiple states:
// - it can get stale because replica A receives transaction execution 1, and replica B receives transaction execution 2,
//   then replica A needs to apply missed transaction execution 2 before continuing
// - it can be outdated because replica A receives project reset, which clears all executions and the cache, but replica B
//   doesn't receive that request so on next run it receives 0 executions but cached emulator contains state from previous
//   executions that wasn't cleared
type emulatorCache struct {
	cache *lru.Cache
}

// reset the cache for the ID.
func (c *emulatorCache) reset(ID uuid.UUID) {
	c.cache.Remove(ID)
}

// get returns a cached emulator if exists, but also checks if it's stale.
func (c *emulatorCache) get(ID uuid.UUID) *emulator {
	val, ok := c.cache.Get(ID)
	if !ok {
		return nil
	}

	return val.(*emulator)
}

// add new entry in the cache.
func (c *emulatorCache) add(ID uuid.UUID, emulator *emulator) {
	c.cache.Add(ID, emulator)
}
