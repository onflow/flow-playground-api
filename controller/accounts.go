package controller

import (
	"fmt"

	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Accounts struct {
	store      storage.Store
	blockchain *blockchain.State
}

func NewAccounts(
	store storage.Store,
	blockchain *blockchain.State,
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

	// todo revisit if this is needed
	exported := make([]*model.Account, len(accounts))
	for i, acc := range accounts {
		exported[i] = acc.Export()

		// todo refactor think about defining a different account model, blockchain account or similar
		a, err := a.blockchain.GetAccount(projectID, acc.Address)
		if err != nil {
			return nil, err
		}

		exported[i].State = a.State
		exported[i].DeployedCode = a.DeployedCode
		exported[i].DeployedContracts = a.DeployedContracts
	}

	return exported, nil
}

func (a *Accounts) Update(input model.UpdateAccount) (*model.Account, error) {
	var acc model.InternalAccount

	if input.DraftCode != nil {
		err := a.store.UpdateAccount(input, &acc)
		if err != nil {
			return nil, err
		}

		return acc.Export(), nil
	}

	err := a.store.GetAccount(model.NewProjectChildID(input.ID, input.ProjectID), &acc)
	if err != nil {
		return nil, err
	}

	if input.DeployedCode == nil {
		return nil, fmt.Errorf("must provide either deployed code or draft code for update")
	}

	account, err := a.blockchain.GetAccount(input.ProjectID, acc.Address)
	if err != nil {
		return nil, err
	}

	if account.DeployedCode != "" {
		// todo reset state
	}

	account, err = a.blockchain.DeployContract(input.ProjectID, acc.Address, *input.DeployedCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy account code")
	}

	// todo refactor ofc
	returnAcc := acc.Export()
	returnAcc.DeployedCode = account.DeployedCode
	returnAcc.DeployedContracts = account.DeployedContracts
	return returnAcc, nil
}
