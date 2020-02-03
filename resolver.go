package flow_playground_api

import (
	"context"

	"github.com/dapperlabs/flow-playground-api/model"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) ScriptExecution() ScriptExecutionResolver {
	return &scriptExecutionResolver{r}
}
func (r *Resolver) TransactionExecution() TransactionExecutionResolver {
	return &transactionExecutionResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateTransactionTemplate(ctx context.Context, input NewTransactionTemplate) (*model.TransactionTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateTransactionTemplate(ctx context.Context, input UpdateTransactionTemplate) (*model.TransactionTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateTransactionExecution(ctx context.Context, input NewTransactionExecution) (*model.TransactionExecution, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateScriptTemplate(ctx context.Context, input NewScriptTemplate) (*model.ScriptTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateScriptTemplate(ctx context.Context, input UpdateScriptTemplate) (*model.ScriptTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateScriptExecution(ctx context.Context, input NewScriptExecution) (*model.ScriptExecution, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Accounts(ctx context.Context) ([]*model.Account, error) {
	panic("not implemented")
}
func (r *queryResolver) TransactionTemplates(ctx context.Context) ([]*model.TransactionTemplate, error) {
	panic("not implemented")
}
func (r *queryResolver) TransactionExecutions(ctx context.Context) ([]*model.TransactionExecution, error) {
	panic("not implemented")
}
func (r *queryResolver) ScriptTemplates(ctx context.Context) ([]*model.ScriptTemplate, error) {
	panic("not implemented")
}
func (r *queryResolver) ScriptExecutions(ctx context.Context) ([]*model.ScriptExecution, error) {
	panic("not implemented")
}

type scriptExecutionResolver struct{ *Resolver }

func (r *scriptExecutionResolver) Template(ctx context.Context, obj *model.ScriptExecution) (*model.ScriptTemplate, error) {
	panic("not implemented")
}

type transactionExecutionResolver struct{ *Resolver }

func (r *transactionExecutionResolver) Template(ctx context.Context, obj *model.TransactionExecution) (*model.TransactionTemplate, error) {
	panic("not implemented")
}
func (r *transactionExecutionResolver) PayerAccount(ctx context.Context, obj *model.TransactionExecution) (*model.Account, error) {
	panic("not implemented")
}
func (r *transactionExecutionResolver) SignerAccounts(ctx context.Context, obj *model.TransactionExecution) ([]*model.Account, error) {
	panic("not implemented")
}
