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
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/getsentry/sentry-go"
	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// newCache returns a new instance of cache with provided capacity.
func newCache(capacity int) *cache {
	return &cache{
		cache: lru.New(capacity),
	}
}

type cache struct {
	cache *lru.Cache
}

// reset the cache for the ID.
func (c *cache) reset(ID uuid.UUID) {
	c.cache.Remove(ID)
}

// get returns a cached emulator if exists, but also checks if it's stale.
//
// based on the executions the function receives it compares that to the emulator block height, since
// one execution is always one block it can compare the heights to the length. If it finds some executions
// that are not part of emulator it returns that subset, so they can be applied on top.
func (c *cache) get(
	ID uuid.UUID,
	executions []*model.TransactionExecution,
) (blockchain, []*model.TransactionExecution, error) {
	val, ok := c.cache.Get(ID)
	if !ok {
		return nil, executions, nil
	}

	emulator := val.(blockchain)
	latest, err := emulator.getLatestBlock()
	if err != nil {
		return nil, nil, errors.Wrap(err, "cache failure")
	}

	// this should never happen, sanity check
	if int(latest.Header.Height) > len(executions) {
		err := fmt.Errorf("cache failure, block height is higher than executions count")
		sentry.CaptureException(err)
		return nil, nil, err
	}

	// this will return only executions that are missing from the emulator
	return emulator, executions[latest.Header.Height:], nil
}

// add new entry in the cache.
func (c *cache) add(ID uuid.UUID, emulator blockchain) {
	c.cache.Add(ID, emulator)
}
