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

// newCache returns a new instance of cache with provided capacity.
func newCache(capacity int) *emulatorCache {
	return &emulatorCache{
		cache: lru.New(capacity),
	}
}

type emulatorCache struct {
	cache *lru.Cache
}

// reset the cache for the ID.
func (c *emulatorCache) reset(ID uuid.UUID) {
	c.cache.Remove(ID)
}

// get returns a cached emulator if exists, but also checks if it's stale.
//
// based on the executions the function receives it compares that to the emulator block height, since
// one execution is always one block it can compare the heights to the length. If it finds some executions
// that are not part of emulator it returns that subset, so they can be applied on top.
func (c *emulatorCache) get(ID uuid.UUID) blockchain {
	val, ok := c.cache.Get(ID)
	if !ok {
		return nil
	}
	return val.(blockchain)
}

// add new entry in the cache.
func (c *emulatorCache) add(ID uuid.UUID, emulator blockchain) {
	c.cache.Add(ID, emulator)
}
