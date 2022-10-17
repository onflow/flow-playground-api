/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"github.com/dapperlabs/flow-playground-api/adapter"
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

type Resolver struct {
	version  *semver.Version
	store    storage.Store
	auth     *auth.Authenticator
	migrator *migrate.Migrator
	projects *controller.Projects
	files    *controller.Files
	//scripts            *controller.Scripts
	//transactions       *controller.Transactions
	//accounts           *controller.Accounts // TODO: Remove accounts?
	lastCreatedProject *model.Project
}

func NewResolver(
	version *semver.Version,
	store storage.Store,
	auth *auth.Authenticator,
	blockchain *blockchain.Projects,
) *Resolver {
	projects := controller.NewProjects(version, store, blockchain)
	scripts := controller.NewScripts(store, blockchain)
	transactions := controller.NewTransactions(store, blockchain)
	accounts := controller.NewAccounts(store, blockchain)
	migrator := migrate.NewMigrator(store, projects)

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

func (r *Resolver) LastCreatedProject() *model.Project {
	return r.lastCreatedProject
}

type updateValidator interface {
	Validate() error
}

func validateUpdate(u updateValidator) error {
	return u.Validate()
}

type mutationResolver struct {
	*Resolver
}

func (r *mutationResolver) authorize(ctx context.Context, ID uuid.UUID) error {
	proj, err := r.projects.Get(ID)

	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}

	if err := r.auth.CheckProjectAccess(ctx, proj); err != nil {
		return err
	}

	return nil
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
	err := r.authorize(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := validateUpdate(&input); err != nil {
		return nil, err
	}

	proj, err := r.projects.Update(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update project")
	}

	return proj.ExportPublicMutable(), nil
}

func (r *mutationResolver) UpdateAccount(ctx context.Context, input model.UpdateAccount) (*model.Account, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	if err := validateUpdate(&input); err != nil {
		return nil, err
	}

	acc, err := r.accounts.Update(adapter.AccountFromAPI(input))
	if err != nil {
		return nil, err
	}

	return adapter.AccountToAPI(acc), nil
}

func (r *mutationResolver) CreateTransactionTemplate(ctx context.Context, input model.NewTransactionTemplate) (*model.TransactionTemplate, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	return r.files.CreateFile(input.ProjectID, model.NewFile(input), model.TransactionFile)
}

func (r *mutationResolver) UpdateTransactionTemplate(ctx context.Context, input model.UpdateTransactionTemplate) (*model.TransactionTemplate, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	if err := validateUpdate(&input); err != nil {
		return nil, err
	}

	return r.files.UpdateFile(model.UpdateFile(input))
}

func (r *mutationResolver) DeleteTransactionTemplate(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (uuid.UUID, error) {
	err := r.authorize(ctx, projectID)
	if err != nil {
		return uuid.UUID{}, err
	}

	err = r.files.DeleteFile(id, projectID)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *mutationResolver) CreateTransactionExecution(
	ctx context.Context,
	input model.NewTransactionExecution,
) (*model.TransactionExecution, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	exe, err := r.files.CreateTransactionExecution(adapter.TransactionFromAPI(input))
	if err != nil {
		return nil, err
	}

	return adapter.TransactionToAPI(exe), nil
}

func (r *mutationResolver) CreateScriptTemplate(ctx context.Context, input model.NewScriptTemplate) (*model.ScriptTemplate, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	file, err := r.files.CreateFile(input.ProjectID, model.NewFile(input), model.ScriptFile)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create script template")
	}

	return file, nil
}

func (r *mutationResolver) UpdateScriptTemplate(ctx context.Context, input model.UpdateScriptTemplate) (*model.ScriptTemplate, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	if err := validateUpdate(&input); err != nil {
		return nil, err
	}

	return r.files.UpdateFile(model.UpdateFile(input))
}

func (r *mutationResolver) DeleteScriptTemplate(
	ctx context.Context,
	id uuid.UUID,
	projectID uuid.UUID,
) (uuid.UUID, error) {
	err := r.authorize(ctx, projectID)
	if err != nil {
		return uuid.UUID{}, err
	}

	err = r.files.DeleteFile(id, projectID)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *mutationResolver) CreateScriptExecution(
	ctx context.Context,
	input model.NewScriptExecution,
) (*model.ScriptExecution, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	exe, err := r.files.CreateScriptExecution(adapter.ScriptFromAPI(input))
	if err != nil {
		return nil, err
	}

	return adapter.ScriptToAPI(exe), nil
}

/*
  createContractTemplate(input: NewContractTemplate!): ContractTemplate!
  updateContractTemplate(input: UpdateContractTemplate!): ContractTemplate!
  deleteContractTemplate(id: UUID!, projectId: UUID!): UUID!
  deployContract(input: NewContractDeployment!): ContractDeployment!
*/

func (r *mutationResolver) UpdateContractTemplate(
	ctx context.Context, input model.UpdateContractTemplate,
) (*model.ContractTemplate, error) {
	err := r.authorize(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	if err := validateUpdate(&input); err != nil {
		return nil, err
	}

	// TODO: Why is this different for contracts?! wtf
	return r.files.UpdateFile(model.UpdateFile(input)), nil
}

func (r *mutationResolver) DeleteContractTemplate(
	ctx context.Context,
	id uuid.UUID,
	projectID uuid.UUID,
) (uuid.UUID, error) {
	err := r.authorize(ctx, projectID) // TODO: need this?
	if err != nil {
		return uuid.UUID{}, err
	}

	err = r.files.DeleteFile(id, projectID)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *mutationResolver) DeployContract(
	ctx context.Context,
	input model.NewContractDeployment,
) (*model.ContractDeployment, error) {
	// TODO: Fix this crap
	err := r.authorize(ctx, input.ProjectID) // TODO: Need this?
	if err != nil {
		return nil, err
	}

	// TODO?!
	exe, err := r.files.DeployContract(adapter.ContractFromAPI(input))
	if err != nil {
		return nil, err
	}

	return adapter.ScriptToAPI(exe), nil
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) Accounts(_ context.Context, proj *model.Project) ([]*model.Account, error) {
	accounts, err := r.accounts.AllForProjectID(proj.ID)
	if err != nil {
		return nil, err
	}

	return adapter.AccountsToAPI(accounts), nil
}

func (r *projectResolver) TransactionTemplates(_ context.Context, proj *model.Project) ([]*model.TransactionTemplate, error) {
	files, err := r.files.GetFilesForProject(proj.ID, model.TransactionFile)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (r *projectResolver) TransactionExecutions(_ context.Context, proj *model.Project) ([]*model.TransactionExecution, error) {

	exes, err := r.files.AllExecutionsForProjectID(proj.ID)
	if err != nil {
		return nil, err
	}

	return adapter.TransactionsToAPI(exes), nil
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
	proj, err := r.projects.Get(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	// todo
	// only migrate if current user has access to this project
	migrated, err := r.migrator.MigrateProject(id, proj.Version, r.version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to migrate project")
	}

	// reload project if needed
	if migrated {
		proj, err = r.projects.Get(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get project")
		}
	}

	if err := r.auth.CheckProjectAccess(ctx, proj); err != nil {
		return proj.ExportPublicImmutable(), nil
	}

	return proj.ExportPublicMutable(), nil
}

func (r *queryResolver) Account(_ context.Context, id uuid.UUID, projectID uuid.UUID) (*model.Account, error) {
	acc, err := r.accounts.GetByID(id, projectID)
	if err != nil {
		return nil, err
	}

	return adapter.AccountToAPI(acc), nil
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
