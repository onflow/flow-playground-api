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
	"fmt"
	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
)

// newEmulatorCache returns a new instance of cache with provided capacity.
func newEmulatorCache(capacity int) *emulatorCache {
	c := lru.New(capacity)
	c.OnEvicted = func(key lru.Key, value interface{}) {
		fmt.Printf("Cache evicted emulator for project: %s - (%v)\n", key.(uuid.UUID).String(), key)
	}

	return &emulatorCache{
		cache: c,
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
func (c *emulatorCache) get(ID uuid.UUID) (*emulator, bool) {
	val, ok := c.cache.Get(ID)
	if !ok || val == nil {
		return nil, false
	}

	em := val.(emulator)
	return &em, true
}

// add new entry in the cache.
func (c *emulatorCache) add(ID uuid.UUID, emulator *emulator) {
	c.cache.Add(ID, *emulator)
}
