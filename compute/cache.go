package compute

import (
	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
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

func (l *LedgerCache) GetOrCreate(
	id uuid.UUID,
	transactionCount int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
) (LedgerCacheItem, error) {
	if transactionCount == 0 {
		return LedgerCacheItem{
			ledger: make(Ledger),
			count:  0,
		}, nil
	}

	ledgerItem, ok := l.Get(id)
	if ok && ledgerItem.count == transactionCount {
		return ledgerItem, nil
	}

	ledger := make(Ledger)

	deltas, err := getRegisterDeltas()
	if err != nil {
		return LedgerCacheItem{}, errors.Wrap(err, "failed to load register deltas for project")
	}

	for _, delta := range deltas {
		ledger.ApplyDelta(delta.Delta)
	}

	ledgerItem = LedgerCacheItem{
		ledger: ledger,
		count:  transactionCount,
	}

	l.Set(id, ledgerItem)

	return ledgerItem, nil
}

func (l *LedgerCache) Set(id uuid.UUID, ledger LedgerCacheItem) {
	l.cache.Add(id, ledger)
}

func (l *LedgerCache) Clear() {
	l.cache.Purge()
}

func (l *LedgerCache) Delete(id uuid.UUID) {
	l.cache.Remove(id)
}
