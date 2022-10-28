package model

import (
	"github.com/google/uuid"
	flowsdk "github.com/onflow/flow-go-sdk"
)

type Account struct {
	ProjectID         uuid.UUID
	Address           Address
	DeployedContracts []string
	State             string
}

func AccountFromFlow(account *flowsdk.Account, projectID uuid.UUID) *Account {
	contractNames := make([]string, 0)
	for name, _ := range account.Contracts {
		contractNames = append(contractNames, name)
	}

	return &Account{
		ProjectID:         projectID,
		Address:           NewAddressFromBytes(account.Address.Bytes()),
		DeployedContracts: contractNames,
	}
}

func (a *Account) Export() *Account {
	return &Account{
		ProjectID:         a.ProjectID,
		Address:           a.Address,
		DeployedContracts: a.DeployedContracts,
		State:             a.State,
	}
}
