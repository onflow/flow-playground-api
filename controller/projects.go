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

package controller

import (
	"github.com/Masterminds/semver"
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go/engine/execution/state/delta"
	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/compute"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Projects struct {
	version     *semver.Version
	store       storage.Store
	computer    *compute.Computer
	numAccounts int
}

func NewProjects(
	version *semver.Version,
	store storage.Store,
	computer *compute.Computer,
	numAccounts int,
) *Projects {
	return &Projects{
		version:     version,
		store:       store,
		computer:    computer,
		numAccounts: numAccounts,
	}
}

func (p *Projects) Create(user *model.User, input model.NewProject) (*model.InternalProject, error) {
	proj := &model.InternalProject{
		ID:          uuid.New(),
		Secret:      uuid.New(),
		PublicID:    uuid.New(),
		ParentID:    input.ParentID,
		Seed:        input.Seed,
		Title:       input.Title,
		Description: input.Description,
		Readme:      input.Readme,
		Persist:     false,
		Version:     p.version,
	}

	accounts, deltas, err := p.createInitialAccounts(proj.ID, input.Accounts)
	if err != nil {
		return nil, err
	}

	ttpls := make([]*model.TransactionTemplate, len(input.TransactionTemplates))

	for i, tpl := range input.TransactionTemplates {
		ttpl := &model.TransactionTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  tpl.Title,
			Script: tpl.Script,
		}

		ttpls[i] = ttpl
	}

	stpls := make([]*model.ScriptTemplate, len(input.ScriptTemplates))

	for i, tpl := range input.ScriptTemplates {
		stpl := &model.ScriptTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  tpl.Title,
			Script: tpl.Script,
		}

		stpls[i] = stpl
	}

	proj.UserID = user.ID

	err = p.store.CreateProject(proj, deltas, accounts, ttpls, stpls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project")
	}

	return proj, nil
}

func (p *Projects) createInitialAccounts(
	projectID uuid.UUID,
	initialContracts []string,
) ([]*model.InternalAccount, []delta.Delta, error) {

	addresses, deltas, err := p.deployInitialAccounts(projectID)
	if err != nil {
		return nil, nil, err
	}

	accounts := make([]*model.InternalAccount, len(addresses))

	for i, address := range addresses {
		account := model.InternalAccount{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: projectID,
			},
			Index:   i,
			Address: address,
		}

		account.SetState(make(model.AccountState))

		if i < len(initialContracts) {
			account.DraftCode = initialContracts[i]
		}

		accounts[i] = &account
	}

	return accounts, deltas, nil
}

func (p *Projects) deployInitialAccounts(projectID uuid.UUID) ([]model.Address, []delta.Delta, error) {

	addresses := make([]model.Address, p.numAccounts)
	deltas := make([]delta.Delta, p.numAccounts)
	regDeltas := make([]*model.RegisterDelta, 0)

	for i := 0; i < p.numAccounts; i++ {

		payer := flow.HexToAddress("01")

		tx := templates.CreateAccount(nil, nil, payer)

		result, err := p.computer.ExecuteTransaction(
			projectID,
			i,
			func() ([]*model.RegisterDelta, error) { return regDeltas, nil },
			toTransactionBody(tx),
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to deploy account code")
		}

		if result.Err != nil {
			return nil, nil, errors.Wrap(result.Err, "failed to deploy account code")
		}

		deltas[i] = result.Delta

		regDeltas = append(regDeltas, &model.RegisterDelta{
			ProjectID: projectID,
			Index:     i,
			Delta:     result.Delta,
		})

		event := result.Events[0]

		eventPayload, err := jsoncdc.Decode(event.Payload)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to deploy account code")
		}

		addressValue := eventPayload.(cadence.Event).Fields[0].(cadence.Address)
		address := model.NewAddressFromBytes(addressValue.Bytes())

		addresses[i] = address
	}

	return addresses, deltas, nil
}

func (p *Projects) Get(id uuid.UUID, proj *model.InternalProject) error {
	err := p.store.GetProject(id, proj)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}

	return nil
}

func (p *Projects) Update(input model.UpdateProject, proj *model.InternalProject) error {
	err := p.store.UpdateProject(input, proj)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}

	return nil
}

func (p *Projects) UpdateVersion(id uuid.UUID, version *semver.Version) error {
	err := p.store.UpdateProjectVersion(id, version)
	if err != nil {
		return errors.Wrap(err, "failed to save project version")
	}

	return nil
}

func (p *Projects) Reset(proj *model.InternalProject) error {
	_, deltas, err := p.deployInitialAccounts(proj.ID)
	if err != nil {
		return err
	}

	err = p.store.ResetProjectState(deltas, proj)
	if err != nil {
		return err
	}

	p.computer.ClearCacheForProject(proj.ID)

	return nil
}

func toTransactionBody(tx *flow.Transaction) *flowgo.TransactionBody {
	txBody := flowgo.NewTransactionBody()
	txBody.SetScript(tx.Script)

	for _, authorizer := range tx.Authorizers {
		txBody.AddAuthorizer(flowgo.Address(authorizer))
	}

	for _, arg := range tx.Arguments {
		txBody.AddArgument(arg)
	}

	return txBody
}
