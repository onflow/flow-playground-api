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
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
)

// newEmulatorCache returns a new instance of cache with provided capacity.
func newEmulatorCache(capacity int) *emulatorCache {
	emCache := &emulatorCache{
		capacity: capacity,
		cache:    nil,
	}
	_ = emCache.initializeCache()
	return emCache
}

// setCache sets emulatorCache to a new lru cache and returns true if successful
func (c *emulatorCache) initializeCache() bool {
	var onEvicted = func(key interface{}, value interface{}) {
		fmt.Printf("Cache evicted emulator for project: %s - (%v)\n",
			key.(uuid.UUID).String(), key)
	}

	cache, err := lru.NewWithEvict(c.capacity, onEvicted)
	if err != nil {
		c.cache = nil
		sentry.CaptureException(err)
		return false
	}

	c.cache = cache
	return true
}

// checkCache return true if cache is accessible, or we can reset it successfully
func (c *emulatorCache) checkCache() bool {
	return c.cache != nil || c.initializeCache()
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
	capacity int
	cache    *lru.Cache
}

// reset the cache for the ID.
func (c *emulatorCache) reset(ID uuid.UUID) {
	if !c.checkCache() {
		return
	}
	c.cache.Remove(ID)
}

// get returns a cached emulator if exists, but also checks if it's stale.
func (c *emulatorCache) get(ID uuid.UUID) *emulator {
	if !c.checkCache() {
		return nil
	}
	val, ok := c.cache.Get(ID)
	if !ok {
		return nil
	}

	return val.(*emulator)
}

// add new entry in the cache.
func (c *emulatorCache) add(ID uuid.UUID, emulator *emulator) {
	if !c.checkCache() {
		return
	}
	c.cache.Add(ID, emulator)
}
