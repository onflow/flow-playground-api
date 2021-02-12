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
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
)

type LedgerCache struct {
	cache *lru.ARCCache
}

func NewLedgerCache(size int) (*LedgerCache, error) {
	cache, err := lru.NewARC(size)
	if err != nil {
		return nil, err
	}

	return &LedgerCache{cache}, nil
}

type ledgerCacheItem struct {
	ledger Ledger
	index  int
}

func (l *LedgerCache) GetOrCreate(
	id uuid.UUID,
	index int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
) (Ledger, error) {
	if index == 0 {
		return make(Ledger), nil
	}

	ledgerItem, ok := l.get(id)
	if ok && ledgerItem.index == index {
		return ledgerItem.ledger, nil
	}

	ledger := make(Ledger)

	deltas, err := getRegisterDeltas()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load register deltas for project")
	}

	for _, delta := range deltas {
		ledger.ApplyDelta(delta.Delta)
	}

	l.Set(id, ledger, index)

	return ledger, nil
}

func (l *LedgerCache) get(id uuid.UUID) (ledgerCacheItem, bool) {
	ledger, ok := l.cache.Get(id)
	if !ok {
		return ledgerCacheItem{}, false
	}

	return ledger.(ledgerCacheItem), true
}

func (l *LedgerCache) Set(id uuid.UUID, ledger Ledger, index int) {
	ledgerItem := ledgerCacheItem{
		ledger: ledger,
		index:  index,
	}

	l.cache.Add(id, ledgerItem)
}

func (l *LedgerCache) Clear() {
	l.cache.Purge()
}

func (l *LedgerCache) Delete(id uuid.UUID) {
	l.cache.Remove(id)
}
