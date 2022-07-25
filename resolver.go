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

package playground

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/pkg/errors"
	"time"

	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/compute"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/migrate"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

const MaxAccounts = 5

type Resolver struct {
	version            *semver.Version
	store              storage.Store
	computer           *compute.Computer
	auth               *auth.Authenticator
	migrator           *migrate.Migrator
	projects           *controller.Projects
	scripts            *controller.Scripts
	lastCreatedProject *model.InternalProject
}

func NewResolver(
	version *semver.Version,
	store storage.Store,
	computer *compute.Computer,
	auth *auth.Authenticator,
) *Resolver {
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

	projects := controller.NewProjects(version, store, computer, MaxAccounts)
	scripts := controller.NewScripts(store, computer)
	migrator := migrate.NewMigrator(projects)

	return &Resolver{
		version:  version,
		store:    store,
		computer: computer,
		auth:     auth,
		migrator: migrator,
		projects: projects,
		scripts:  scripts,
	}
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Project() ProjectResolver {
	return &projectResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) TransactionExecution() TransactionExecutionResolver {
	return &transactionExecutionResolver{r}
}

func (r *Resolver) LastCreatedProject() *model.InternalProject {
	return r.lastCreatedProject
}

type mutationResolver struct {
	*Resolver
}

func (r *mutationResolver) CreateProject(ctx context.Context, input model.NewProject) (*model.Project, error) {
	user, err := r.auth.GetOrCreateUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get or create user")
	}

	proj, err := r.projects.Create(user, input)
	if err != nil {
		return nil, err
	}

	r.lastCreatedProject = proj

	return proj.ExportPublicMutable(), nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, input model.UpdateProject) (*model.Project, error) {
	var proj model.InternalProject

	err := r.projects.Get(input.ID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	err = r.projects.Update(input, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update project")
	}

	return proj.ExportPublicMutable(), nil
}

func (r *mutationResolver) UpdateAccount(ctx context.Context, input model.UpdateAccount) (*model.Account, error) {

	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	var acc model.InternalAccount

	err = r.store.GetAccount(model.NewProjectChildID(input.ID, proj.ID), &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	if input.DeployedCode == nil {
		err = r.store.UpdateAccount(input, &acc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update account")
		}

		return acc.Export(), nil
	}

	// Redeploy: clear all state
	if acc.DeployedCode != "" {
		err := r.projects.Reset(&proj)
		if err != nil {
			return nil, errors.Wrap(err, "failed to clear project state")
		}
	}

	address := acc.Address.ToFlowAddress()
	source := *input.DeployedCode
	contractName, err := getSourceContractName(source)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy account code")
	}

	tx := templates.AddAccountContract(address, templates.Contract{
		Name:   contractName,
		Source: source,
	})

	result, err := r.computer.ExecuteTransaction(
		proj.ID,
		proj.TransactionCount,
		func() ([]*model.RegisterDelta, error) {
			var deltas []*model.RegisterDelta
			err := r.store.GetRegisterDeltasForProject(proj.ID, &deltas)
			if err != nil {
				return nil, err
			}

			return deltas, nil
		},
		toTransactionBody(tx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy account code")
	}

	if result.Err != nil {
		return nil, errors.Wrap(result.Err, "failed to deploy account code")
	}

	states, err := r.getAccountStates(proj.ID, result.States)
	if err != nil {
		return nil, err
	}

	input.DeployedContracts = &[]string{contractName}

	err = r.store.UpdateAccountAfterDeployment(input, states, result.Delta, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update account")
	}

	return acc.Export(), nil
}

func getSourceContractName(code string) (string, error) {
	program, err := parser2.ParseProgram(code, nil)
	if err != nil {
		return "", err
	}
	return getProgramContractName(program)
}

func getProgramContractName(program *ast.Program) (string, error) {

	// The code may declare exactly one contract or one contract interface.

	var contractDeclarations []*ast.CompositeDeclaration
	var contractInterfaceDeclarations []*ast.InterfaceDeclaration

	for _, compositeDeclaration := range program.CompositeDeclarations() {
		if compositeDeclaration.CompositeKind == common.CompositeKindContract {
			contractDeclarations = append(contractDeclarations, compositeDeclaration)
		} else {
			return "", fmt.Errorf(
				"invalid %s: the code must declare exactly one contract or contract interface",
				compositeDeclaration.DeclarationKind().Name(),
			)
		}
	}

	for _, interfaceDeclaration := range program.InterfaceDeclarations() {
		if interfaceDeclaration.CompositeKind == common.CompositeKindContract {
			contractInterfaceDeclarations = append(contractInterfaceDeclarations, interfaceDeclaration)
		} else {
			return "", fmt.Errorf(
				"invalid %s: the code must declare exactly one contract or contract interface",
				interfaceDeclaration.DeclarationKind().Name(),
			)
		}
	}

	switch {
	case len(contractDeclarations) == 1 && len(contractInterfaceDeclarations) == 0:
		return contractDeclarations[0].Identifier.Identifier, nil
	case len(contractInterfaceDeclarations) == 1 && len(contractDeclarations) == 0:
		return contractInterfaceDeclarations[0].Identifier.Identifier, nil
	default:
		return "", errors.New(
			"the code must declare exactly one contract or contract interface",
		)
	}
}

func (r *mutationResolver) getAccountStates(
	projectID uuid.UUID,
	newStates compute.AccountStates,
) (map[uuid.UUID]model.AccountState, error) {
	var accounts []*model.InternalAccount

	err := r.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project accounts")
	}

	states := make(map[uuid.UUID]model.AccountState)

	for _, account := range accounts {
		stateDelta, ok := newStates[account.Address]
		if !ok {
			continue
		}

		state, err := account.State()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get account state")
		}

		for key, value := range stateDelta {
			state[key] = value
		}

		states[account.ID] = state
	}

	return states, nil
}

func (r *mutationResolver) CreateTransactionTemplate(ctx context.Context, input model.NewTransactionTemplate) (*model.TransactionTemplate, error) {
	tpl := &model.TransactionTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: input.ProjectID,
		},
		Title:  input.Title,
		Script: input.Script,
	}

	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	err = r.store.InsertTransactionTemplate(tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store transaction template")
	}

	return tpl, nil
}

func (r *mutationResolver) UpdateTransactionTemplate(ctx context.Context, input model.UpdateTransactionTemplate) (*model.TransactionTemplate, error) {
	var tpl model.TransactionTemplate

	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	err = r.store.UpdateTransactionTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update transaction template")
	}

	return &tpl, nil
}

func (r *mutationResolver) DeleteTransactionTemplate(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (uuid.UUID, error) {
	var proj model.InternalProject

	err := r.projects.Get(projectID, &proj)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return uuid.Nil, err
	}

	err = r.store.DeleteTransactionTemplate(model.NewProjectChildID(id, projectID))
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to delete transaction template")
	}

	return id, nil
}

func (r *mutationResolver) CreateTransactionExecution(
	ctx context.Context,
	input model.NewTransactionExecution,
) (*model.TransactionExecution, error) {
	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	tx := flow.NewTransaction().
		SetScript([]byte(input.Script))

	for i, argument := range input.Arguments {
		// Decode and then encode again to ensure the value is valid

		value, err := jsoncdc.Decode(nil, []byte(argument))
		if err == nil {
			err = tx.AddArgument(value)
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode argument %d", i))
		}
	}

	for _, authorizer := range input.Signers {
		tx.AddAuthorizer(authorizer.ToFlowAddress())
	}

	result, err := r.computer.ExecuteTransaction(
		proj.ID,
		proj.TransactionCount,
		func() ([]*model.RegisterDelta, error) {
			var deltas []*model.RegisterDelta
			err := r.store.GetRegisterDeltasForProject(proj.ID, &deltas)
			if err != nil {
				return nil, err
			}

			return deltas, nil
		},
		toTransactionBody(tx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute transaction")
	}

	exe := model.TransactionExecution{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: input.ProjectID,
		},
		Script:    input.Script,
		Arguments: input.Arguments,
		Logs:      result.Logs,
	}

	var states map[uuid.UUID]model.AccountState

	if result.Err != nil {
		exe.Errors = compute.ExtractProgramErrors(result.Err)
	} else {
		var err error
		states, err = r.getAccountStates(proj.ID, result.States)
		if err != nil {
			return nil, err
		}
	}

	events, err := parseEvents(result.Events)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse events")
	}

	exe.Events = events

	err = r.store.InsertTransactionExecution(&exe, states, result.Delta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction execution record")
	}

	return &exe, nil
}

func (r *mutationResolver) CreateScriptTemplate(ctx context.Context, input model.NewScriptTemplate) (*model.ScriptTemplate, error) {
	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	tpl, err := r.scripts.CreateTemplate(proj.ID, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create script template")
	}

	return tpl, nil
}

func (r *mutationResolver) UpdateScriptTemplate(ctx context.Context, input model.UpdateScriptTemplate) (*model.ScriptTemplate, error) {
	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	var tpl model.ScriptTemplate

	err = r.scripts.UpdateTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update script template")
	}

	return &tpl, nil
}

func (r *mutationResolver) DeleteScriptTemplate(
	ctx context.Context,
	id uuid.UUID,
	projectID uuid.UUID,
) (uuid.UUID, error) {
	var proj model.InternalProject

	err := r.projects.Get(projectID, &proj)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return uuid.Nil, err
	}

	err = r.scripts.DeleteTemplate(id, projectID)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *mutationResolver) CreateScriptExecution(
	ctx context.Context,
	input model.NewScriptExecution,
) (*model.ScriptExecution, error) {
	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	exe, err := r.scripts.CreateExecution(&proj, input.Script, input.Arguments)
	if err != nil {
		return nil, err
	}

	return exe, nil
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) Accounts(ctx context.Context, obj *model.Project) ([]*model.Account, error) {
	var accs []*model.InternalAccount

	err := r.store.GetAccountsForProject(obj.ID, &accs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	exportedAccs := make([]*model.Account, len(accs))

	for i, acc := range accs {
		exported, err := acc.ExportWithJSONState()
		if err != nil {
			return nil, errors.Wrap(err, "failed to export")
		}

		exportedAccs[i] = exported
	}

	return exportedAccs, nil
}

func (r *projectResolver) TransactionTemplates(ctx context.Context, obj *model.Project) ([]*model.TransactionTemplate, error) {
	var tpls []*model.TransactionTemplate

	err := r.store.GetTransactionTemplatesForProject(obj.ID, &tpls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction templates")
	}

	return tpls, nil
}

func (r *projectResolver) TransactionExecutions(ctx context.Context, obj *model.Project) ([]*model.TransactionExecution, error) {
	var exes []*model.TransactionExecution

	err := r.store.GetTransactionExecutionsForProject(obj.ID, &exes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction executions")
	}

	return exes, nil
}

func (r *projectResolver) ScriptTemplates(ctx context.Context, obj *model.Project) ([]*model.ScriptTemplate, error) {
	var tpls []*model.ScriptTemplate

	err := r.store.GetScriptTemplatesForProject(obj.ID, &tpls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get script templates")
	}

	return tpls, nil
}

func (r *projectResolver) ScriptExecutions(ctx context.Context, obj *model.Project) ([]*model.ScriptExecution, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) PlaygroundInfo(ctx context.Context) (*model.PlaygroundInfo, error) {
	return &model.PlaygroundInfo{
		APIVersion:     *r.version,
		CadenceVersion: *semver.MustParse(cadence.Version),
	}, nil
}

func (r *queryResolver) Project(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var proj model.InternalProject

	err := r.projects.Get(id, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return proj.ExportPublicImmutable(), nil
	}

	// only migrate if current user has access to this project

	migrated, err := r.migrator.MigrateProject(id, proj.Version, r.version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to migrate project")
	}

	// reload project if needed

	if migrated {
		err := r.projects.Get(id, &proj)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get project")
		}
	}

	return proj.ExportPublicMutable(), nil
}

func (r *queryResolver) Account(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (*model.Account, error) {
	var acc model.InternalAccount

	err := r.store.GetAccount(model.NewProjectChildID(id, projectID), &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	exported, err := acc.ExportWithJSONState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to export")
	}

	return exported, nil
}

func (r *queryResolver) TransactionTemplate(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (*model.TransactionTemplate, error) {
	var tpl model.TransactionTemplate

	err := r.store.GetTransactionTemplate(model.NewProjectChildID(id, projectID), &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction template")
	}

	return &tpl, nil
}

func (r *queryResolver) ScriptTemplate(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (*model.ScriptTemplate, error) {
	var tpl model.ScriptTemplate

	err := r.store.GetScriptTemplate(model.NewProjectChildID(id, projectID), &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get script template")
	}

	return &tpl, nil
}

type transactionExecutionResolver struct{ *Resolver }

func (*transactionExecutionResolver) Signers(_ context.Context, _ *model.TransactionExecution) ([]*model.Account, error) {
	panic("not implemented")
}

func parseEvents(flowEvents []flowgo.Event) ([]model.Event, error) {
	events := make([]model.Event, len(flowEvents))

	for i, event := range flowEvents {
		parsedEvent, err := parseEvent(event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse event")
		}
		events[i] = parsedEvent
	}

	return events, nil
}

func parseEvent(event flowgo.Event) (model.Event, error) {
	payload, err := jsoncdc.Decode(nil, event.Payload)
	if err != nil {
		return model.Event{}, errors.Wrap(err, "failed to decode event payload (JSON-CDC)")
	}

	fields := payload.(cadence.Event).Fields
	values := make([]string, len(fields))
	for j, field := range fields {
		enc, err := jsoncdc.Encode(field)
		if err != nil {
			return model.Event{}, errors.Wrap(err, "failed to encode event field to JSON-CDC")
		}

		values[j] = string(enc)
	}

	return model.Event{
		Type:   string(event.Type),
		Values: values,
	}, nil
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
