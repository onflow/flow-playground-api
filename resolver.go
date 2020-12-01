package playground

import (
	"context"

	"github.com/Masterminds/semver"
	flowgo "github.com/dapperlabs/flow-go/model/flow"
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/compute"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/migrate"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

const MaxAccounts = 4

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

	tx := templates.UpdateAccountCode(address, []byte(*input.DeployedCode))

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

	contracts, err := parseDeployedContracts(result.Events)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse deployed contracts")
	}

	input.DeployedContracts = &contracts

	err = r.store.UpdateAccountAfterDeployment(input, states, result.Delta, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update account")
	}

	return acc.Export(), nil
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
		Script: input.Script,
		Logs:   result.Logs,
	}

	var states map[uuid.UUID]model.AccountState

	if result.Err != nil {
		runtimeErr := result.Err.Error()
		exe.Error = &runtimeErr
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

	exe, err := r.scripts.CreateExecution(&proj, input.Script)
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

func (r *transactionExecutionResolver) Signers(ctx context.Context, obj *model.TransactionExecution) ([]*model.Account, error) {
	panic("not implemented")
}

const AccountCodeUpdatedEvent = "flow.AccountCodeUpdated"

func parseDeployedContracts(events []cadence.Event) ([]string, error) {
	for _, event := range events {
		if event.Type().ID() == AccountCodeUpdatedEvent {
			arrayValue := event.Fields[2].(cadence.Array)

			contracts := make([]string, len(arrayValue.Values))

			for i, contractValue := range arrayValue.Values {
				contracts[i] = contractValue.(cadence.String).ToGoValue().(string)
			}

			return contracts, nil
		}
	}

	return nil, nil
}

func parseEvents(rtEvents []cadence.Event) ([]model.Event, error) {
	events := make([]model.Event, len(rtEvents))

	for i, event := range rtEvents {

		values := make([]string, len(event.Fields))
		for j, field := range event.Fields {
			enc, err := jsoncdc.Encode(field)
			if err != nil {
				return nil, errors.Wrap(err, "failed to encode to JSON-CDC")
			}

			values[j] = string(enc)
		}

		events[i] = model.Event{
			Type:   event.Type().ID(),
			Values: values,
		}
	}

	return events, nil
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
