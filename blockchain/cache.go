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
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
)

// newLruCache wraps creating a new lru cache in error handling
func newLruCache(capacity int) *lru.Cache {
	cache, err := lru.New(capacity)
	if err != nil {
		sentry.CaptureException(err)
		return nil
	}
	return cache
}

// emulatorCache caches the emulator state.
//
// In the environment where multiple replicas maintain its own cache copy it can get into multiple states:
// - it can get stale because replica A receives transaction execution 1, and replica B receives transaction execution 2,
//   then replica A needs to apply missed transaction execution 2 before continuing
// - it can be outdated because replica A receives project reset, which clears all executions and the cache, but replica B
//   doesn't receive that request so on next run it receives 0 executions but cached emulator contains state from previous
//   executions that wasn't cleared
type emulatorCache struct {
	capacity int
	cache    *lru.Cache
}

// newEmulatorCache returns a new instance of emulatorCache with provided capacity.
func newEmulatorCache(capacity int) *emulatorCache {
	return &emulatorCache{
		capacity: capacity,
		cache:    newLruCache(capacity),
	}
}

// reset the cached emulator for the ID.
func (c *emulatorCache) reset(ID uuid.UUID) {
	if c.cache == nil {
		return
	}
	c.cache.Remove(ID)
}

// get returns a cached emulator for specified ID if it exists
func (c *emulatorCache) get(ID uuid.UUID) *emulator {
	if c.cache == nil {
		return nil
	}

	val, ok := c.cache.Get(ID)
	if !ok {
		return nil
	}

	return val.(*emulator)
}

// add new emulator to the cache.
func (c *emulatorCache) add(ID uuid.UUID, emulator *emulator) {
	if c.cache == nil {
		// Try to initialize new cache
		c.cache = newLruCache(c.capacity)
		if c.cache == nil {
			return
		}
	}

	c.cache.Add(ID, emulator)
}
