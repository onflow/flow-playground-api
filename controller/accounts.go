package controller

import (
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Accounts struct {
	store      storage.Store
	blockchain blockchain.Blockchain
}

func NewAccounts(
	store storage.Store,
	blockchain blockchain.Blockchain,
) *Accounts {
	return &Accounts{
		store:      store,
		blockchain: blockchain,
	}
}

func (a *Accounts) GetByID(ID uuid.UUID, projectID uuid.UUID) (*model.Account, error) {
	var acc model.InternalAccount

	err := a.store.GetAccount(model.NewProjectChildID(ID, projectID), &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	return acc.Export(), nil
}

func (a *Accounts) AllForProjectID(projectID uuid.UUID) ([]*model.Account, error) {
	var accounts []*model.InternalAccount

	err := a.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	// todo get account storage
	// a.blockchain.GetAccount() add storage to get account

	// todo revisit if this is needed
	exported := make([]*model.Account, len(accounts))
	for i, acc := range accounts {
		exported[i] = acc.Export()
	}

	return exported, nil
}
