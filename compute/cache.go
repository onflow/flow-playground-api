package compute

import (
	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru"
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
