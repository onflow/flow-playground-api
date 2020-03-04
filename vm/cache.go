package vm

import (
	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru"
)

type LedgerCacheItem struct {
	ledger Ledger
	count  int
}

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

func (l *LedgerCache) Get(id uuid.UUID) (LedgerCacheItem, bool) {
	ledger, ok := l.cache.Get(id)
	if !ok {
		return LedgerCacheItem{}, false
	}

	return ledger.(LedgerCacheItem), true
}

func (l *LedgerCache) Set(id uuid.UUID, ledger LedgerCacheItem) {
	l.cache.Add(id, ledger)
}

func (l *LedgerCache) Clear() {
	l.cache.Purge()
}
