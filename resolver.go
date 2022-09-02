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

	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/migrate"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"
)

const MaxAccounts = 5

type Resolver struct {
	version            *semver.Version
	store              storage.Store
	auth               *auth.Authenticator
	migrator           *migrate.Migrator
	projects           *controller.Projects
	scripts            *controller.Scripts
	transactions       *controller.Transactions
	accounts           *controller.Accounts
	lastCreatedProject *model.InternalProject
}

func NewResolver(
	version *semver.Version,
	store storage.Store,
	auth *auth.Authenticator,
	blockchain *blockchain.State,
) *Resolver {
	projects := controller.NewProjects(version, store, MaxAccounts, blockchain)
	scripts := controller.NewScripts(store, blockchain)
	transactions := controller.NewTransactions(store, blockchain)
	accounts := controller.NewAccounts(store, blockchain)
	migrator := migrate.NewMigrator(projects)

	return &Resolver{
		version:      version,
		store:        store,
		auth:         auth,
		migrator:     migrator,
		projects:     projects,
		scripts:      scripts,
		transactions: transactions,
		accounts:     accounts,
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
	// todo refactor auth check
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

	return r.accounts.Update(input)
}

func (r *mutationResolver) CreateTransactionTemplate(ctx context.Context, input model.NewTransactionTemplate) (*model.TransactionTemplate, error) {
	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	return r.transactions.CreateTemplate(input.ProjectID, input.Title, input.Script)
}

func (r *mutationResolver) UpdateTransactionTemplate(ctx context.Context, input model.UpdateTransactionTemplate) (*model.TransactionTemplate, error) {
	var proj model.InternalProject

	err := r.projects.Get(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, &proj); err != nil {
		return nil, err
	}

	return r.transactions.UpdateTemplate(input)
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

	err = r.transactions.DeleteTemplate(id, projectID)
	if err != nil {
		return id, err
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

	return r.transactions.CreateTransactionExecution(input)
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

	return r.scripts.UpdateTemplate(input)
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

func (r *projectResolver) Accounts(_ context.Context, proj *model.Project) ([]*model.Account, error) {
	return r.accounts.AllForProjectID(proj.ID)
}

func (r *projectResolver) TransactionTemplates(_ context.Context, proj *model.Project) ([]*model.TransactionTemplate, error) {
	return r.transactions.AllTemplatesForProjectID(proj.ID)
}

func (r *projectResolver) TransactionExecutions(_ context.Context, proj *model.Project) ([]*model.TransactionExecution, error) {
	return r.transactions.AllExecutionsForProjectID(proj.ID)
}

func (r *projectResolver) ScriptTemplates(_ context.Context, proj *model.Project) ([]*model.ScriptTemplate, error) {
	return r.scripts.AllTemplatesForProjectID(proj.ID)
}

func (r *projectResolver) ScriptExecutions(_ context.Context, _ *model.Project) ([]*model.ScriptExecution, error) {
	// todo implement
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) PlaygroundInfo(_ context.Context) (*model.PlaygroundInfo, error) {
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

func (r *queryResolver) Account(_ context.Context, id uuid.UUID, projectID uuid.UUID) (*model.Account, error) {
	return r.accounts.GetByID(id, projectID)
}

func (r *queryResolver) TransactionTemplate(_ context.Context, id uuid.UUID, projectID uuid.UUID) (*model.TransactionTemplate, error) {
	return r.transactions.TemplateByID(id, projectID)
}

func (r *queryResolver) ScriptTemplate(_ context.Context, id uuid.UUID, projectID uuid.UUID) (*model.ScriptTemplate, error) {
	return r.scripts.TemplateByID(id, projectID)
}

type transactionExecutionResolver struct{ *Resolver }

func (*transactionExecutionResolver) Signers(_ context.Context, _ *model.TransactionExecution) ([]*model.Account, error) {
	panic("not implemented")
}
