package playground

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go-sdk/templates"
	"github.com/dapperlabs/flow-go/language"
	"github.com/dapperlabs/flow-go/language/runtime"

	encoding2 "github.com/dapperlabs/flow-playground-api/encoding"
	"github.com/dapperlabs/flow-playground-api/middleware"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/vm"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

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
		Persist:  false,
	}

	err := r.store.InsertProject(proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store project")
	}

	for i := 0; i < MaxAccounts; i++ {
		acc := model.InternalAccount{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Index:     i,
		}

		if i < len(input.Accounts) {
			acc.DraftCode = input.Accounts[i]
		}

		script, _ := templates.CreateAccount(nil, nil)
		result, delta, state, err := r.computer.ExecuteTransaction(acc.ProjectID, string(script), nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to deploy account code")
		}

		if result.Error != nil {
			return nil, errors.Wrap(result.Error, "failed to deploy account code")
		}

		err = r.store.InsertRegisterDelta(acc.ProjectID, delta)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store register delta")
		}

		value, _ := language.ConvertValue(result.Events[0].Fields[0])
		addressValue := value.(language.Address)

		address := model.Address(flow.BytesToAddress(addressValue.Bytes()))

		acc.Address = address

		acc.State = state[address]

		err = r.store.InsertAccount(&acc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store account")
		}
	}

	for _, script := range input.TransactionTemplates {
		tpl := &model.TransactionTemplate{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Script:    script,
		}

		err = r.store.InsertTransactionTemplate(tpl)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store transaction template")
		}
	}

	for _, script := range input.ScriptTemplates {
		tpl := &model.ScriptTemplate{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Script:    script,
		}

		err = r.store.InsertScriptTemplate(tpl)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store script template")
		}
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
	var acc model.InternalAccount

	err := r.store.GetAccount(input.ID, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	var proj model.InternalProject

	err = r.store.GetProject(acc.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return nil, errors.New("access denied")
	}

	// TODO: make deployment atomic
	if input.DeployedCode != nil {

		script := string(templates.UpdateAccountCode([]byte(*input.DeployedCode)))
		result, delta, state, err := r.computer.ExecuteTransaction(acc.ProjectID, script, []model.Address{acc.Address})
		if err != nil {
			return nil, errors.Wrap(err, "failed to deploy account code")
		}

		if result.Error != nil {
			return nil, errors.Wrap(result.Error, "failed to deploy account code")
		}

		err = r.updateAccountState(proj.ID, state)
		if err != nil {
			return nil, err
		}

		err = r.store.InsertRegisterDelta(acc.ProjectID, delta)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store register delta")
		}

		contracts, err := parseDeployedContracts(result.Events)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse deployed contracts")
		}

		input.DeployedContracts = &contracts
	}

	err = r.store.UpdateAccount(input, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update account")
	}

	return acc.Export(), nil
}

func (r *mutationResolver) updateAccountState(projectID uuid.UUID, state vm.AccountState) error {
	var accounts []*model.InternalAccount

	err := r.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return errors.Wrap(err, "failed to get project accounts")
	}

	for _, account := range accounts {
		for key, value := range state[account.Address] {
			account.State[key] = value
		}

		err := r.store.UpdateAccountState(account.ID, account.State)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *mutationResolver) CreateTransactionTemplate(ctx context.Context, input model.NewTransactionTemplate) (*model.TransactionTemplate, error) {
	tpl := &model.TransactionTemplate{
		ID:        uuid.New(),
		ProjectID: input.ProjectID,
		Script:    input.Script,
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

	err := r.store.GetTransactionTemplate(input.ID, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction template")
	}

	var proj model.InternalProject

	err = r.store.GetProject(tpl.ProjectID, &proj)
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

func (r *mutationResolver) DeleteTransactionTemplate(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var tpl model.TransactionTemplate

	err := r.store.GetTransactionTemplate(id, &tpl)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get transaction template")
	}

	var proj model.InternalProject

	err = r.store.GetProject(tpl.ProjectID, &proj)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return uuid.Nil, errors.New("access denied")
	}

	err = r.store.DeleteTransactionTemplate(id)
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

	result, delta, state, err := r.computer.ExecuteTransaction(proj.ID, input.Script, input.Signers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute transaction")
	}

	exe := model.TransactionExecution{
		ID:        uuid.New(),
		ProjectID: proj.ID,
		Script:    input.Script,
		Logs:      result.Logs,
	}

	if result.Error != nil {
		runtimeErr := result.Error.Error()
		exe.Error = &runtimeErr
	} else {
		err = r.updateAccountState(proj.ID, state)
		if err != nil {
			return nil, err
		}
	}

	if len(result.Events) > 0 {
		events := make([]model.Event, len(result.Events))
		for i, event := range result.Events {

			values := make([]string, len(event.Fields))
			for j, field := range event.Fields {

				value, err := encoding2.ConvertValue(field)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert event value")
				}

				encoded, err := json.Marshal(value)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON encode event value")
				}

				values[j] = string(encoded)
			}

			events[i] = model.Event{
				Type:   string(event.Type.ID()),
				Values: values,
			}
		}

		exe.Events = events
	}

	err = r.store.InsertTransactionExecution(&exe, delta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction execution record")
	}

	return &exe, nil
}

func (r *mutationResolver) CreateScriptTemplate(ctx context.Context, input model.NewScriptTemplate) (*model.ScriptTemplate, error) {
	tpl := &model.ScriptTemplate{
		ID:        uuid.New(),
		ProjectID: input.ProjectID,
		Script:    input.Script,
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

	err := r.store.GetScriptTemplate(input.ID, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get script template")
	}

	var proj model.InternalProject

	err = r.store.GetProject(tpl.ProjectID, &proj)
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

	result, err := r.computer.ExecuteScript(input.ProjectID, input.Script)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute script")
	}

	exe := model.ScriptExecution{
		ID:        uuid.New(),
		ProjectID: input.ProjectID,
		Script:    input.Script,
		Logs:      result.Logs,
	}

	if result.Error != nil {
		runtimeErr := result.Error.Error()
		exe.Error = &runtimeErr
	} else {
		value, err := encoding2.ConvertValue(result.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert value")
		}

		enc, err := json.Marshal(value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode to JSON")
		}

		exe.Value = string(enc)
	}

	err = r.store.InsertScriptExecution(&exe)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert script execution record")
	}

	return &exe, nil
}

func (r *mutationResolver) DeleteScriptTemplate(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var tpl model.ScriptTemplate

	err := r.store.GetScriptTemplate(id, &tpl)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get script template")
	}

	var proj model.InternalProject

	err = r.store.GetProject(tpl.ProjectID, &proj)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get project")
	}

	if !middleware.ProjectInSession(ctx, &proj) {
		return uuid.Nil, errors.New("access denied")
	}

	err = r.store.DeleteScriptTemplate(id)
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

func (r *queryResolver) Account(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	var acc model.InternalAccount

	err := r.store.GetAccount(id, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	exported, err := acc.ExportWithJSONState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to export")
	}

	return exported, nil
}

func (r *queryResolver) TransactionTemplate(ctx context.Context, id uuid.UUID) (*model.TransactionTemplate, error) {
	var tpl model.TransactionTemplate

	err := r.store.GetTransactionTemplate(id, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction template")
	}

	return &tpl, nil
}

func (r *queryResolver) ScriptTemplate(ctx context.Context, id uuid.UUID) (*model.ScriptTemplate, error) {
	var tpl model.ScriptTemplate

	err := r.store.GetScriptTemplate(id, &tpl)
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

func parseDeployedContracts(events []runtime.Event) ([]string, error) {
	for _, event := range events {
		if event.Type.ID() == AccountCodeUpdatedEvent {
			value, err := language.ConvertValue(event.Fields[2])
			if err != nil {
				return nil, err
			}
			arrayValue := value.(language.VariableSizedArray)

			contracts := make([]string, len(arrayValue.Values))

			for i, contractValue := range arrayValue.Values {
				contracts[i] = contractValue.(language.String).ToGoValue().(string)
			}

			return contracts, nil
		}
	}

	return nil, nil
}
