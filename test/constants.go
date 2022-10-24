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

package test

import (
	"fmt"
	"github.com/dapperlabs/flow-playground-api/model"
)

type Project struct {
	ID                   string
	Title                string
	Description          string
	Readme               string
	Seed                 int
	Persist              bool
	Version              string
	NumberOfAccounts     int
	TransactionTemplates []TransactionTemplate
	ScriptTemplates      []ScriptTemplate
	ContractTemplates    []ContractTemplate
	Secret               string
}

const MutationCreateProject = `
mutation($title: String!, $description: String!, $readme: String!, $seed: Int!, $numberOfAccounts: Int!, $transactionTemplates: [NewProjectTransactionTemplate!], $scriptTemplates: [NewProjectScriptTemplate!], $contractTemplates: [NewProjectContractTemplate!]) {
  createProject(input: { title: $title, description: $description, readme: $readme, seed: $seed, numberOfAccounts: $numberOfAccounts, transactionTemplates: $transactionTemplates, scriptTemplates: $scriptTemplates, contractTemplates: $contractTemplates }) {
    id
    title
		description
		readme
    seed
	numberOfAccounts
    transactionTemplates {
      id
      title
      script
      index
    }
    scriptTemplates {
      id
      title
      script
      index
    }
    contractTemplates {
      id
      title
      script
      index
    }
  }
}
`

type CreateProjectResponse struct {
	CreateProject Project
}

const QueryGetProject = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
  }
}
`

type GetProjectResponse struct {
	Project Project
}

const MutationUpdateProjectPersist = `
mutation($projectId: UUID!, $title: String!, $description: String!, $readme: String!, $persist: Boolean!) {
  updateProject(input: { id: $projectId, title: $title, description: $description, readme: $readme, persist: $persist }) {
    id
		title
		description
		readme
    persist
  }
}
`

type UpdateProjectResponse struct {
	UpdateProject struct {
		ID          string
		Title       string
		Description string
		Readme      string
		Persist     bool
	}
}

const QueryGetProjectTransactionTemplates = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    transactionTemplates {
      id
      script
      index
    }
  }
}
`

type GetProjectTransactionTemplatesResponse struct {
	Project struct {
		ID                   string
		TransactionTemplates []struct {
			ID     string
			Script string
			Index  int
		}
	}
}

const QueryGetProjectScriptTemplates = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    scriptTemplates {
      id
      script
      index
    }
  }
}
`

type GetProjectScriptTemplatesResponse struct {
	Project struct {
		ID              string
		ScriptTemplates []struct {
			ID     string
			Script string
			Index  int
		}
	}
}

type File struct {
	ID     string
	Title  string
	Script string
	Type   int
	Index  int
}

type ContractTemplate struct {
	ID     string
	Title  string
	Script string
	Index  int
}

type TransactionTemplate struct {
	ID     string
	Title  string
	Script string
	Index  int
}

const MutationCreateTransactionTemplate = `
mutation($projectId: UUID!, $title: String!, $script: String!) {
  createTransactionTemplate(input: { projectId: $projectId, title: $title, script: $script }) {
    id
    title
    script
    index
  }
}
`

type CreateTransactionTemplateResponse struct {
	CreateTransactionTemplate TransactionTemplate
}

const QueryGetTransactionTemplate = `
query($templateId: UUID!, $projectId: UUID!) {
  transactionTemplate(id: $templateId, projectId: $projectId) {
    id
    script
    index
  }
}
`

type GetTransactionTemplateResponse struct {
	TransactionTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationUpdateTransactionTemplateScript = `
mutation($templateId: UUID!, $projectId: UUID!, $script: String!) {
  updateTransactionTemplate(input: { id: $templateId, projectId: $projectId, script: $script }) {
    id
    script
    index
  }
}
`

const MutationUpdateTransactionTemplateIndex = `
mutation($templateId: UUID!, $projectId: UUID!, $index: Int!) {
  updateTransactionTemplate(input: { id: $templateId, projectId: $projectId, index: $index }) {
    id
    script
    index
  }
}
`

type UpdateTransactionTemplateResponse struct {
	UpdateTransactionTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationDeleteTransactionTemplate = `
mutation($templateId: UUID!, $projectId: UUID!) {
  deleteTransactionTemplate(id: $templateId, projectId: $projectId)
}
`

type DeleteTransactionTemplateResponse struct {
	DeleteTransactionTemplate string
}

const MutationCreateTransactionExecution = `
mutation($projectId: UUID!, $script: String!, $signers: [Address!], $arguments: [String!]) {
  createTransactionExecution(input: {
    projectId: $projectId,
    script: $script,
    arguments: $arguments,
    signers: $signers
  }) {
    id
    script
    errors {
      message
      startPosition { offset line column }
      endPosition { offset line column }
    }
    logs
    events {
      type
      values
    }
  }
}
`

type CreateTransactionExecutionResponse struct {
	CreateTransactionExecution struct {
		ID     string
		Script string
		Errors []model.ProgramError
		Logs   []string
		Events []struct {
			Type   string
			Values []string
		}
	}
}

const MutationCreateScriptExecution = `
mutation CreateScriptExecution($projectId: UUID!, $script: String!, $arguments: [String!]) {
  createScriptExecution(input: {
    projectId: $projectId,
    script: $script,
    arguments: $arguments
  }) {
    id
    script
    errors {
      message
      startPosition { offset line column }
      endPosition { offset line column }
    }
    logs
    value
  }
}
`

type CreateScriptExecutionResponse struct {
	CreateScriptExecution struct {
		ID     string
		Script string
		Errors []model.ProgramError
		Logs   []string
		Value  string
	}
}

const MutationCreateContractTemplate = `
mutation($projectId: UUID!, $title: String!, $script: String!) {
  createContractTemplate(input: { projectId: $projectId, title: $title, script: $script }) {
    id
    title
    script
    index
  }
}
`

const MutationCreateContractDeployment = `
mutation($projectId: UUID!, $script: String!, $address: Address!) {
  createContractDeployment(input: { projectId: $projectId, script: $script, address: $address }) {
    id
    script
    address
  }
}
`

const MutationCreateScriptTemplate = `
mutation($projectId: UUID!, $title: String!, $script: String!) {
  createScriptTemplate(input: { projectId: $projectId, title: $title, script: $script }) {
    id
    title
    script
    index
  }
}
`

type CreateContractDeploymentResponse struct {
	CreateContractDeployment struct {
		ID      string
		Script  string
		Address model.Address
		Errors  []model.ProgramError
		Events  []struct {
			Type   string
			Values []string
		}
		Logs []string
	}
}

type ScriptTemplate struct {
	ID     string
	Title  string
	Script string
	Index  int
}

type CreateScriptTemplateResponse struct {
	CreateScriptTemplate ScriptTemplate
}

const QueryGetScriptTemplate = `
query($templateId: UUID!, $projectId: UUID!) {
  scriptTemplate(id: $templateId, projectId: $projectId) {
    id
    script
  }
}
`

type GetScriptTemplateResponse struct {
	ScriptTemplate ScriptTemplate
}

const MutationUpdateScriptTemplateScript = `
mutation($templateId: UUID!, $projectId: UUID!, $script: String!) {
  updateScriptTemplate(input: { id: $templateId, projectId: $projectId, script: $script }) {
    id
    script
    index
  }
}
`

const MutationUpdateScriptTemplateIndex = `
mutation($templateId: UUID!, $projectId: UUID!, $index: Int!) {
  updateScriptTemplate(input: { id: $templateId, projectId: $projectId, index: $index }) {
    id
    script
    index
  }
}
`

type UpdateScriptTemplateResponse struct {
	UpdateScriptTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationDeleteScriptTemplate = `
mutation($templateId: UUID!, $projectId: UUID!) {
  deleteScriptTemplate(id: $templateId, projectId: $projectId)
}
`

type DeleteScriptTemplateResponse struct {
	DeleteScriptTemplate string
}

const initAccounts = 5

const counterContract = `
  pub contract Counting {

      pub event CountIncremented(count: Int)

      pub resource Counter {
          pub var count: Int

          init() {
              self.count = 0
          }

          pub fun add(_ count: Int) {
              self.count = self.count + count
              emit CountIncremented(count: self.count)
          }
      }

      pub fun createCounter(): @Counter {
          return <-create Counter()
      }
  }
`

// generateAddTwoToCounterScript generates a script that increments a counter.
// If no counter exists, it is created.
func generateAddTwoToCounterScript(counterAddress string) string {
	return fmt.Sprintf(
		`
            import 0x%s

            transaction {

                prepare(signer: AuthAccount) {
                    if signer.borrow<&Counting.Counter>(from: /storage/counter) == nil {
                        signer.save(<-Counting.createCounter(), to: /storage/counter)
                    }

                    signer.borrow<&Counting.Counter>(from: /storage/counter)!.add(2)
                }
            }
        `,
		counterAddress,
	)
}

/*
func TestContractImport(t *testing.T) {
	t.Parallel()
	c := newClient()

	project := createProject(t, c)

	accountA := project.Accounts[0]
	accountB := project.Accounts[1]

	contractA := `
	pub contract HelloWorldA {
		pub var A: String
		pub init() { self.A = "HelloWorldA" }
	}`

	contractB := `
	import HelloWorldA from 0x01
	pub contract HelloWorldB {
		pub init() {
			log(HelloWorldA.A)
		}
	}`

	var respA UpdateAccountResponse

	err := c.Post(
		MutationUpdateAccountDeployedCode,
		&respA,
		client2.Var("projectId", project.ID),
		client2.Var("accountId", accountA.ID),
		client2.Var("code", contractA),
		client2.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
	assert.Equal(t, contractA, respA.UpdateAccount.DeployedCode)

	var respB UpdateAccountResponse

	err = c.Post(
		MutationUpdateAccountDeployedCode,
		&respB,
		client2.Var("projectId", project.ID),
		client2.Var("accountId", accountB.ID),
		client2.Var("code", contractB),
		client2.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
}

// TODO implement these tests on flow accounts still?
func TestAccountStorage(t *testing.T) {
	t.Parallel()
	c := newClient()

	project := createProject(t, c)
	account := project.Accounts[0]

	var accResp GetAccountResponse

	err := c.Post(
		QueryGetAccount,
		&accResp,
		client2.Var("projectId", project.ID),
		client2.Var("accountId", account.ID),
	)
	require.NoError(t, err)

	assert.Equal(t, account.ID, accResp.Account.ID)
	assert.Equal(t, `{}`, accResp.Account.State)

	var resp CreateTransactionExecutionResponse

	const script = `
		transaction {
		  prepare(signer: AuthAccount) {
			  	signer.save("storage value", to: /storage/storageTest)
 				signer.link<&String>(/public/publicTest, target: /storage/storageTest)
				signer.link<&String>(/private/privateTest, target: /storage/storageTest)
		  }
   		}`

	err = c.Post(
		MutationCreateTransactionExecution,
		&resp,
		client2.Var("projectId", project.ID),
		client2.Var("script", script),
		client2.Var("signers", []string{account.Address}),
		client2.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	err = c.Post(
		QueryGetAccount,
		&accResp,
		client2.Var("projectId", project.ID),
		client2.Var("accountId", account.ID),
	)
	require.NoError(t, err)

	assert.Equal(t, account.ID, accResp.Account.ID)
	assert.NotEmpty(t, accResp.Account.State)

	type accountStorage struct {
		Private map[string]any
		Public  map[string]any
		Storage map[string]any
	}

	var accStorage accountStorage
	err = json.Unmarshal([]byte(accResp.Account.State), &accStorage)
	require.NoError(t, err)

	assert.Equal(t, "storage value", accStorage.Storage["storageTest"])
	assert.NotEmpty(t, accStorage.Private["privateTest"])
	assert.NotEmpty(t, accStorage.Public["publicTest"])

	assert.NotContains(t, accStorage.Public, "flowTokenBalance")
	assert.NotContains(t, accStorage.Public, "flowTokenReceiver")
	assert.NotContains(t, accStorage.Storage, "flowTokenVault")
}
*/

// todo add tests for:
// - failed transactions with successful transactions work (bootstrap works)??
// - assert we don't leak any internal model data to API
