package playground

import (
	"context"
	"fmt"

	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go-sdk/templates"
	"github.com/dapperlabs/flow-go/language"
	"github.com/dapperlabs/flow-go/language/encoding"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/vm"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	store    storage.Store
	computer *vm.Computer
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

type mutationResolver struct {
	*Resolver
}

func (r *mutationResolver) CreateProject(ctx context.Context) (*model.Project, error) {
	proj := &model.Project{
		ID: uuid.New(),
	}

	err := r.store.InsertProject(proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store project")
	}

	// TODO: clean this up
	for i := 0; i < 3; i++ {
		acc := model.Account{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Index:     i,
		}

		script, _ := templates.CreateAccount(nil, nil)
		result, delta, err := r.computer.ExecuteTransaction(acc.ProjectID, string(script), nil)
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

		err = r.store.InsertAccount(&acc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store account")
		}
	}

	return proj, nil
}

func (r *mutationResolver) UpdateAccount(ctx context.Context, input model.UpdateAccount) (*model.Account, error) {
	var acc model.Account

	err := r.store.GetAccount(input.ID, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	// TODO: make deployment atomic
	if input.DeployedCode != nil {
		script := string(templates.UpdateAccountCode([]byte(*input.DeployedCode)))
		result, delta, err := r.computer.ExecuteTransaction(acc.ProjectID, script, []model.Address{acc.Address})
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
	}

	err = r.store.UpdateAccount(input, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update account")
	}

	return &acc, nil
}

func (r *mutationResolver) CreateTransactionTemplate(ctx context.Context, input model.NewTransactionTemplate) (*model.TransactionTemplate, error) {
	tpl := &model.TransactionTemplate{
		ID:        uuid.New(),
		ProjectID: input.ProjectID,
		Script:    input.Script,
	}

	var proj model.Project

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	err = r.store.InsertTransactionTemplate(tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store transaction template")
	}

	return tpl, nil
}

func (r *mutationResolver) UpdateTransactionTemplate(ctx context.Context, input model.UpdateTransactionTemplate) (*model.TransactionTemplate, error) {
	var tpl model.TransactionTemplate

	err := r.store.UpdateTransactionTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update transaction template")
	}

	return &tpl, nil
}

func (r *mutationResolver) DeleteTransactionTemplate(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	err := r.store.DeleteTransactionTemplate(id)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to delete transaction template")
	}

	return id, nil
}

func (r *mutationResolver) CreateTransactionExecution(
	ctx context.Context,
	input model.NewTransactionExecution,
) (*model.TransactionExecution, error) {
	var proj model.Project

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	result, delta, err := r.computer.ExecuteTransaction(input.ProjectID, input.Script, input.Signers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute transaction")
	}

	exe := model.TransactionExecution{
		ID:        uuid.New(),
		ProjectID: input.ProjectID,
		Script:    input.Script,
		Logs:      result.Logs,
	}

	if result.Error != nil {
		runtimeErr := result.Error.Error()
		exe.Error = &runtimeErr
	}

	if len(result.Events) > 0 {
		events := make([]model.Event, len(result.Events))
		for i, event := range result.Events {

			values := make([]*model.XDRValue, len(event.Fields))
			for j, field := range event.Fields {
				value, err := language.ConvertValue(field)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert event value")
				}

				encValue, err := encoding.Encode(value)
				if err != nil {
					return nil, errors.Wrap(err, "failed to encode event value")
				}

				values[j] = &model.XDRValue{
					Type:  value.Type().ID(),
					Value: fmt.Sprintf("%x", encValue),
				}
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

	var proj model.Project

	err := r.store.GetProject(input.ProjectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	err = r.store.InsertScriptTemplate(tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store script template")
	}

	return tpl, nil
}

func (r *mutationResolver) UpdateScriptTemplate(ctx context.Context, input model.UpdateScriptTemplate) (*model.ScriptTemplate, error) {
	var tpl model.ScriptTemplate

	err := r.store.UpdateScriptTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update script template")
	}

	return &tpl, nil
}

func (r *mutationResolver) CreateScriptExecution(ctx context.Context, input model.NewScriptExecution) (*model.ScriptExecution, error) {
	panic("not implemented")
}

func (r *mutationResolver) DeleteScriptTemplate(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	err := r.store.DeleteScriptTemplate(id)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to delete script template")
	}

	return id, nil
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) Accounts(ctx context.Context, obj *model.Project) ([]*model.Account, error) {
	var accs []*model.Account

	err := r.store.GetAccountsForProject(obj.ID, &accs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	return accs, nil
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
	var proj model.Project

	err := r.store.GetProject(id, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	return &proj, nil
}

func (r *queryResolver) Account(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	var acc model.Account

	err := r.store.GetAccount(id, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	return &acc, nil
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
