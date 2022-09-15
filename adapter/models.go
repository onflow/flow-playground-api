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

package adapter

import "github.com/dapperlabs/flow-playground-api/model"

// models adapters compose different adapters in a single adapter.

func TransactionFromAPI(tx model.NewTransactionExecution) model.NewTransactionExecution {
	tx.Script = contentAddressesFromInput(tx.Script)
	tx.Signers = addressesFromInput(tx.Signers)

	for i, arg := range tx.Arguments {
		tx.Arguments[i] = contentAddressesFromInput(arg)
	}

	return tx
}

func TransactionToAPI(tx *model.TransactionExecution) *model.TransactionExecution {
	tx.Script = contentAddressToOutput(tx.Script)
	tx.Signers = addressesToOutput(tx.Signers)

	for i, arg := range tx.Arguments {
		tx.Arguments[i] = contentAddressesFromInput(arg)
	}

	for i, e := range tx.Events {
		for j, v := range e.Values {
			tx.Events[i].Values[j] = contentAddressToOutput(v)
		}
	}

	for i, e := range tx.Errors {
		tx.Errors[i].Message = contentAddressToOutput(e.Message)
	}

	return tx
}

func TransactionsToAPI(txs []*model.TransactionExecution) []*model.TransactionExecution {
	for i, tx := range txs {
		txs[i] = TransactionToAPI(tx)
	}
	return txs
}

func ScriptFromAPI(script model.NewScriptExecution) model.NewScriptExecution {
	script.Script = contentAddressesFromInput(script.Script)
	for i, a := range script.Arguments {
		script.Arguments[i] = contentAddressesFromInput(a)
	}
	return script
}

func ScriptToAPI(script *model.ScriptExecution) *model.ScriptExecution {
	script.Script = contentAddressToOutput(script.Script)

	for i, e := range script.Errors {
		script.Errors[i].Message = contentAddressToOutput(e.Message)
	}

	for i, a := range script.Arguments {
		script.Arguments[i] = contentAddressToOutput(a)
	}

	script.Value = contentAddressToOutput(script.Value)

	return script
}

func AccountToAPI(account *model.Account) *model.Account {
	account.Address = addressToOutput(account.Address)
	account.DeployedCode = contentAddressToOutput(account.DeployedCode)

	// todo storage adapter

	return account
}

func AccountsToAPI(accounts []*model.Account) []*model.Account {
	for i, a := range accounts {
		accounts[i] = AccountToAPI(a)
	}
	return accounts
}

func AccountFromAPI(account model.UpdateAccount) model.UpdateAccount {
	adaptedCode := contentAddressesFromInput(*account.DeployedCode)
	account.DeployedCode = &adaptedCode
	return account
}
