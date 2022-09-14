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

package controller

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/go-chi/render"
	"github.com/onflow/cadence"
)

type UtilsHandler struct{}

func NewUtilsHandler() *UtilsHandler {
	return &UtilsHandler{}
}

func (u *UtilsHandler) VersionHandler(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, struct {
		Version string `json:"version"`
	}{
		cadence.Version,
	})
}

// Backward compatibility address adapters section.
//
// Because new blockchain execution is done using the emulator, it takes up first X accounts as service accounts, so if we
// want to keep the same address space for the user we need to translate addresses coming from the user to the backend and vice-versa.
// todo temp workaround to prevent API breaking changes, remove this in the v2.
// We can avoid doing translations of address in the next version of playground we can start the address space at 0x0a.

const numberOfServiceAccounts = 4
const addressLength = 8

// contentAddressesFromInput converts addresses found in content from the user input.
func contentAddressesFromInput(input string) string {
	return contentAdapter(input, true)
}

// contentAddressesFromInput converts addresses found in content to the user output.
func contentAddressToOutput(input string) string {
	return contentAdapter(input, false)
}

func contentAdapter(input string, fromInput bool) string {
	r := regexp.MustCompile(`0x0*(\d+)`)

	// we must use this logic since if we parse the address to Address type
	// it outputs it in standard format which might be different to the input format
	for _, addressMatch := range r.FindAllStringSubmatch(input, -1) {
		original := addressMatch[0]
		addr, _ := strconv.Atoi(addressMatch[1])

		if fromInput {
			addr = addr + numberOfServiceAccounts
		} else if addr > numberOfServiceAccounts { // don't convert if service address, shouldn't happen
			addr = addr - numberOfServiceAccounts
		}

		replaced := strings.ReplaceAll(original, addressMatch[1], fmt.Sprintf("%d", addr))
		input = strings.ReplaceAll(input, original, replaced)
	}

	return input
}

// todo temp workaround to prevent API breaking changes, remove this in the v2.
// addressFromInput converts the address from the user input and shifts it for number of service accounts.
func addressFromInput(address model.Address) model.Address {
	var b model.Address // create a copy
	copy(b[:], address[:])
	b[len(b)-1] = b[len(b)-1] + numberOfServiceAccounts
	return b
}

func addressesFromInput(addresses []model.Address) []model.Address {
	for i, address := range addresses {
		addresses[i] = addressFromInput(address)
	}
	return addresses
}

// todo temp workaround to prevent API breaking changes, remove this in the v2.
// addressFromInput converts the address to the user output by subtracting the number of service accounts.
func addressToOutput(address model.Address) model.Address {
	var b model.Address
	copy(b[addressLength-len(address):], address[:])
	b[len(b)-1] = b[len(b)-1] - numberOfServiceAccounts
	return b
}

func addressesToOutput(addresses []model.Address) []model.Address {
	for i, address := range addresses {
		addresses[i] = addressToOutput(address)
	}
	return addresses
}

func transactionAdapterFromInput(tx model.NewTransactionExecution) model.NewTransactionExecution {
	tx.Script = contentAddressesFromInput(tx.Script)
	tx.Signers = addressesFromInput(tx.Signers)

	for i, arg := range tx.Arguments {
		tx.Arguments[i] = contentAddressesFromInput(arg)
	}

	return tx
}

func transactionAdapterToOutput(tx *model.TransactionExecution) *model.TransactionExecution {
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

func transactionsAdapterToOutput(txs []*model.TransactionExecution) []*model.TransactionExecution {
	for i, tx := range txs {
		txs[i] = transactionAdapterToOutput(tx)
	}
	return txs
}

func scriptAdapterFromInput(script model.NewScriptExecution) model.NewScriptExecution {
	script.Script = contentAddressesFromInput(script.Script)
	for i, a := range script.Arguments {
		script.Arguments[i] = contentAddressesFromInput(a)
	}
	return script
}

func scriptAdapterToOutput(script *model.ScriptExecution) *model.ScriptExecution {
	script.Script = contentAddressToOutput(script.Script)

	for i, e := range script.Errors {
		script.Errors[i].Message = contentAddressToOutput(e.Message)
	}

	for i, a := range script.Arguments {
		script.Arguments[i] = contentAddressToOutput(a)
	}

	return script
}

func accountAdapterToOutput(account *model.Account) *model.Account {
	account.Address = addressToOutput(account.Address)
	account.DeployedCode = contentAddressToOutput(account.DeployedCode)
	// todo check state
	return account
}
