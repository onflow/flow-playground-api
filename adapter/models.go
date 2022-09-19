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
	tx.Script = ContentAddressFromAPI(tx.Script)
	tx.Signers = addressesFromAPI(tx.Signers)

	for i, arg := range tx.Arguments {
		tx.Arguments[i] = ContentAddressFromAPI(arg)
	}

	return tx
}

func TransactionToAPI(tx *model.TransactionExecution) *model.TransactionExecution {
	tx.Script = contentAddressToAPI(tx.Script)
	tx.Signers = addressesToAPI(tx.Signers)

	for i, arg := range tx.Arguments {
		tx.Arguments[i] = ContentAddressFromAPI(arg)
	}

	for i, e := range tx.Events {
		for j, v := range e.Values {
			tx.Events[i].Values[j] = contentAddressToAPI(v)
		}
	}

	for i, e := range tx.Errors {
		tx.Errors[i].Message = contentAddressToAPI(e.Message)
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
	script.Script = ContentAddressFromAPI(script.Script)
	for i, a := range script.Arguments {
		script.Arguments[i] = ContentAddressFromAPI(a)
	}
	return script
}

func ScriptToAPI(script *model.ScriptExecution) *model.ScriptExecution {
	script.Script = contentAddressToAPI(script.Script)

	for i, e := range script.Errors {
		script.Errors[i].Message = contentAddressToAPI(e.Message)
	}

	for i, a := range script.Arguments {
		script.Arguments[i] = contentAddressToAPI(a)
	}

	script.Value = contentAddressToAPI(script.Value)

	return script
}

func AccountToAPI(account *model.Account) *model.Account {
	account.Address = addressToAPI(account.Address)
	account.DeployedCode = contentAddressToAPI(account.DeployedCode)

	account.State = stateToAPI(account.State)

	return account
}

func AccountsToAPI(accounts []*model.Account) []*model.Account {
	for i, a := range accounts {
		accounts[i] = AccountToAPI(a)
	}
	return accounts
}

func AccountFromAPI(account model.UpdateAccount) model.UpdateAccount {
	if account.DeployedCode != nil {
		adaptedCode := ContentAddressFromAPI(*account.DeployedCode)
		account.DeployedCode = &adaptedCode
	}
	return account
}
