package playground

import (
	"context"

	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/middleware"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/vm"
)

const MaxAccounts = 4

type Resolver struct {
	store              storage.Store
	computer           *vm.Computer
	lastCreatedProject *model.InternalProject
}

func NewResolver(store storage.Store, computer *vm.Computer) *Resolver {
	return &Resolver{
		store:    store,
		computer: computer,
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
	proj := &model.InternalProject{
		ID:       uuid.New(),
		Secret:   uuid.New(),
		PublicID: uuid.New(),
		ParentID: input.ParentID,
		Seed:     input.Seed,
		Title:    input.Title,
		Persist:  false,
	}

	var (
		deltas    []delta.Delta
		regDeltas []*model.RegisterDelta
		accounts  []*model.InternalAccount
		ttpls     []*model.TransactionTemplate
		stpls     []*model.ScriptTemplate
	)

	for i := 0; i < MaxAccounts; i++ {
		acc := model.InternalAccount{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Index: i,
		}

		if i < len(input.Accounts) {
			acc.DraftCode = input.Accounts[i]
		}

		script, _ := templates.CreateAccount(nil, nil)
		result, delta, state, err := r.computer.ExecuteTransaction(
			acc.ProjectID,
			i,
			func() ([]*model.RegisterDelta, error) { return regDeltas, nil },
			string(script),
			[]model.Address{model.NewAddressFromBytes(flow.HexToAddress("01").Bytes())},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to deploy account code")
		}

		if result.Err != nil {
			return nil, errors.Wrap(result.Err, "failed to deploy account code")
		}

		deltas = append(deltas, delta)
		regDeltas = append(regDeltas, &model.RegisterDelta{
			ProjectID:         acc.ProjectID,
			Index:             i,
			Delta:             delta,
			IsAccountCreation: true,
		})

		addressValue := result.Events[0].Fields[0].(cadence.Address)
		address := model.NewAddressFromBytes(addressValue.Bytes())

		acc.Address = address

		acc.State = state[address]

		accounts = append(accounts, &acc)

	}

	for _, tpl := range input.TransactionTemplates {
		ttpl := &model.TransactionTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  tpl.Title,
			Script: tpl.Script,
		}

		ttpls = append(ttpls, ttpl)
	}

	for _, tpl := range input.ScriptTemplates {
		stpl := &model.ScriptTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  tpl.Title,
			Script: tpl.Script,
		}

		stpls = append(stpls, stpl)
	}

	err := r.store.CreateProject(proj, deltas, accounts, ttpls, stpls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project")
	}

	// add project to HTTP session
	if err := middleware.AddProjectToSession(ctx, proj); err != nil {
		return nil, errors.Wrap(err, "failed to save project in session")
	}

	r.lastCreatedProject = proj

	return proj.ExportPublicMutable(), nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, input model.UpdateProject) (*model.Project, error) {
	var proj model.InternalProject

	err := r.store.GetProject(input.ID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
	}

	err = r.store.UpdateProject(input, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update project")
	}

	return proj.ExportPublicMutable(), nil
}

func (r *mutationResolver) UpdateAccount(ctx context.Context, input model.UpdateAccount) (*model.Account, error) {

	var proj model.InternalProject

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
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

	transactionCount := proj.TransactionCount

	// Redeploy: clear all state
	if acc.DeployedCode != "" {
		var err error
		transactionCount, err = r.store.ClearProjectState(proj.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to clear project state")
		}

		r.computer.ClearCacheForProject(proj.ID)
	}

	script := string(templates.UpdateAccountCode([]byte(*input.DeployedCode)))
	result, delta, state, err := r.computer.ExecuteTransaction(
		proj.ID,
		transactionCount,
		func() ([]*model.RegisterDelta, error) {
			var deltas []*model.RegisterDelta
			err := r.store.GetRegisterDeltasForProject(proj.ID, &deltas)
			if err != nil {
				return nil, err
			}

			return deltas, nil
		},
		script,
		[]model.Address{acc.Address},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy account code")
	}

	if result.Err != nil {
		return nil, errors.Wrap(result.Err, "failed to deploy account code")
	}

	states, err := r.getAccountStates(proj.ID, state)
	if err != nil {
		return nil, err
	}

	contracts, err := parseDeployedContracts(result.Events)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse deployed contracts")
	}

	input.DeployedContracts = &contracts

	err = r.store.UpdateAccountAfterDeployment(input, states, delta, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update account")
	}

	return acc.Export(), nil
}

func (r *mutationResolver) getAccountStates(projectID uuid.UUID, state vm.AccountState) (map[uuid.UUID]map[string][]byte, error) {
	var accounts []*model.InternalAccount

	err := r.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project accounts")
	}

	states := make(map[uuid.UUID]map[string][]byte)

	for _, account := range accounts {
		stateDelta, ok := state[account.Address]
		if !ok {
			continue
		}

		for key, value := range stateDelta {
			account.State[key] = value
		}

		states[account.ID] = account.State
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

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
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

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
	}

	err = r.store.UpdateTransactionTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update transaction template")
	}

	return &tpl, nil
}

func (r *mutationResolver) DeleteTransactionTemplate(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (uuid.UUID, error) {
	var proj model.InternalProject

	err := r.store.GetProject(projectID, &proj)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return uuid.Nil, errors.New("access denied")
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

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
	}

	result, delta, state, err := r.computer.ExecuteTransaction(
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
		input.Script,
		input.Signers,
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

	var states map[uuid.UUID]map[string][]byte

	if result.Err != nil {
		runtimeErr := result.Err.Error()
		exe.Error = &runtimeErr
	} else {
		var err error
		states, err = r.getAccountStates(proj.ID, state)
		if err != nil {
			return nil, err
		}
	}

	events, err := parseEvents(result.Events)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse events")
	}

	exe.Events = events

	err = r.store.InsertTransactionExecution(&exe, states, delta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction execution record")
	}

	return &exe, nil
}

func (r *mutationResolver) CreateScriptTemplate(ctx context.Context, input model.NewScriptTemplate) (*model.ScriptTemplate, error) {
	tpl := &model.ScriptTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: input.ProjectID,
		},
		Title:  input.Title,
		Script: input.Script,
	}

	var proj model.InternalProject

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
	}

	err = r.store.InsertScriptTemplate(tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store script template")
	}

	return tpl, nil
}

func (r *mutationResolver) UpdateScriptTemplate(ctx context.Context, input model.UpdateScriptTemplate) (*model.ScriptTemplate, error) {
	var tpl model.ScriptTemplate

	var proj model.InternalProject

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
	}

	err = r.store.UpdateScriptTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update script template")
	}

	return &tpl, nil
}

func (r *mutationResolver) CreateScriptExecution(ctx context.Context, input model.NewScriptExecution) (*model.ScriptExecution, error) {
	var proj model.InternalProject

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if len(input.Script) == 0 {
		return nil, errors.New("cannot execute empty script")
	}

	result, err := r.computer.ExecuteScript(
		input.ProjectID,
		proj.TransactionCount,
		func() ([]*model.RegisterDelta, error) {
			var deltas []*model.RegisterDelta
			err := r.store.GetRegisterDeltasForProject(proj.ID, &deltas)
			if err != nil {
				return nil, err
			}

			return deltas, nil
		},
		input.Script,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute script")
	}

	exe := model.ScriptExecution{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: input.ProjectID,
		},
		Script: input.Script,
		Logs:   result.Logs,
	}

	if result.Err != nil {
		runtimeErr := result.Err.Error()
		exe.Error = &runtimeErr
	} else {
		enc, err := jsoncdc.Encode(result.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode to JSON-CDC")
		}

		exe.Value = string(enc)
	}

	err = r.store.InsertScriptExecution(&exe)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert script execution record")
	}

	return &exe, nil
}

func (r *mutationResolver) DeleteScriptTemplate(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (uuid.UUID, error) {
	var proj model.InternalProject

	err := r.store.GetProject(projectID, &proj)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return uuid.Nil, errors.New("access denied")
	}

	err = r.store.DeleteScriptTemplate(model.NewProjectChildID(id, projectID))
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to delete script template")
	}

	return id, nil
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

func (r *queryResolver) Project(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var proj model.InternalProject

	err := r.store.GetProject(id, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if middleware.ProjectInSession(ctx, &proj) {
		return proj.ExportPublicMutable(), nil
	}

	return proj.ExportPublicImmutable(), nil
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
