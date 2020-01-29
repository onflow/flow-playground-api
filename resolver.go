package flow_playground_api

import (
	"context"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

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

func (r *mutationResolver) CreateTransactionTemplate(ctx context.Context, input NewTransactionTemplate) (*TransactionTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateTransactionTemplate(ctx context.Context, input UpdateTransactionTemplate) (*TransactionTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateTransactionExecution(ctx context.Context, input NewTransactionExecution) (*TransactionExecution, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateScriptTemplate(ctx context.Context, input NewScriptTemplate) (*ScriptTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateScriptTemplate(ctx context.Context, input UpdateScriptTemplate) (*ScriptTemplate, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateScriptExecution(ctx context.Context, input NewScriptExecution) (*ScriptExecution, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Accounts(ctx context.Context) ([]*Account, error) {
	panic("not implemented")
}
func (r *queryResolver) TransactionTemplates(ctx context.Context) ([]*TransactionTemplate, error) {
	panic("not implemented")
}
func (r *queryResolver) TransactionExecutions(ctx context.Context) ([]*TransactionExecution, error) {
	panic("not implemented")
}
func (r *queryResolver) ScriptTemplates(ctx context.Context) ([]*ScriptTemplate, error) {
	panic("not implemented")
}
func (r *queryResolver) ScriptExecutions(ctx context.Context) ([]*ScriptExecution, error) {
	panic("not implemented")
}

type scriptExecutionResolver struct{ *Resolver }

func (r *scriptExecutionResolver) Template(ctx context.Context, obj *ScriptExecution) (*ScriptTemplate, error) {
	panic("not implemented")
}

type transactionExecutionResolver struct{ *Resolver }

func (r *transactionExecutionResolver) Template(ctx context.Context, obj *TransactionExecution) (*TransactionTemplate, error) {
	panic("not implemented")
}
func (r *transactionExecutionResolver) PayerAccount(ctx context.Context, obj *TransactionExecution) (*Account, error) {
	panic("not implemented")
}
func (r *transactionExecutionResolver) SignerAccounts(ctx context.Context, obj *TransactionExecution) ([]*Account, error) {
	panic("not implemented")
}
