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

package e2eTest

import (
	"github.com/dapperlabs/flow-playground-api/model"
)

const initAccounts = 5
const addr1 = "0000000000000001"
const addr2 = "0000000000000002"
const addr3 = "0000000000000003"
const addr4 = "0000000000000004"
const addr5 = "0000000000000005"

type Project struct {
	ID                   string
	Title                string
	Description          string
	Readme               string
	Seed                 int
	Persist              bool
	Version              string
	NumberOfAccounts     int
	UpdatedAt            string
	Accounts             []Account
	TransactionTemplates []TransactionTemplate
	ScriptTemplates      []ScriptTemplate
	ContractTemplates    []ContractTemplate
	Secret               string
}

type Account struct {
	Address           string
	DeployedContracts []string
	State             string
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
	updatedAt
	accounts {
      address
      deployedContracts
      state
	}
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

type DeleteProjectResponse struct {
	DeleteProject string
}

const MutationDeleteProject = `
mutation($projectId: UUID!) {
  deleteProject(projectId: $projectId)
}
`

const QueryGetAccount = `
query($address: Address!, $projectId: UUID!) {
  account(address: $address, projectId: $projectId) {
    address
    deployedContracts
    state
  }
}
`

type GetAccountResponse struct {
	Account Account
}

const QueryGetProjectStorage = `
query($projectId: UUID!) {
  project(id: $projectId) {
	accounts {
		address
		deployedContracts
		state
	}
	transactionTemplates {
      id
      script
      index
    }
	scriptTemplates {
      id
      script
      index
	}
	contractTemplates {
      id
      script
      index
	}
  }
}
`

const QueryGetProject = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
	updatedAt
  }
}
`

type GetProjectResponse struct {
	Project Project
}

const QueryGetProjectList = `
query() {
  projectList() {
    projects {
      id
      title
    }
  }
}
`

type GetProjectListResponse struct {
	ProjectList struct {
		Projects []*Project
	}
}

const MutationUpdateProjectPersist = `
mutation($projectId: UUID!, $title: String!, $description: String!, $readme: String!, $persist: Boolean!) {
  updateProject(input: { id: $projectId, title: $title, description: $description, readme: $readme, persist: $persist }) {
    id
		title
		description
		readme
    updatedAt
    persist
  }
}
`

type UpdateProjectResponse struct {
	//UpdateProject Project

	UpdateProject struct {
		ID          string
		Title       string
		Description string
		Readme      string
		Persist     bool
		UpdatedAt   string
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

const QueryGetProjectContractTemplates = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    contractTemplates {
      id
      title
      script
      index
    }
  }
}
`

type GetProjectContractTemplatesResponse struct {
	Project struct {
		ID                string
		ContractTemplates []struct {
			ID     string
			Title  string
			Script string
			Index  int
		}
	}
}

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

type GetContractTemplateResponse struct {
	ContractTemplate struct {
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

const MutationUpdateContractTemplateScript = `
mutation($templateId: UUID!, $projectId: UUID!, $script: String!) {
  updateContractTemplate(input: { id: $templateId, projectId: $projectId, script: $script }) {
    id
    script
    index
  }
}
`

const MutationUpdateContractTemplateTitle = `
mutation($templateId: UUID!, $projectId: UUID!, $title: String) {
  updateContractTemplate(input: { id: $templateId, projectId: $projectId, title: $title }) {
    id
	title
    script
    index
  }
}
`

const MutationUpdateContractTemplateIndex = `
mutation($templateId: UUID!, $projectId: UUID!, $index: Int!) {
  updateContractTemplate(input: { id: $templateId, projectId: $projectId, index: $index }) {
    id
    script
    index
  }
}
`

type UpdateContractTemplateResponse struct {
	UpdateContractTemplate struct {
		ID     string
		Index  int
		Title  string
		Script string
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

const MutationDeleteContractTemplate = `
mutation($templateId: UUID!, $projectId: UUID!) {
  deleteContractTemplate(id: $templateId, projectId: $projectId)
}
`

type DeleteContractTemplateResponse struct {
	DeleteContractTemplate string
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

const MutationCreateContractDeployment = `
mutation($projectId: UUID!, $script: String!, $address: Address!, $arguments: [String!]) {
  createContractDeployment(input: {
	projectId: $projectId,
	script: $script,
	address: $address
	arguments: $arguments
  }) {
    id
	title
    script
    arguments
    address
	blockHeight
    errors {
      message
      startPosition { offset line column }
      endPosition { offset line column }
    }
    events {
      type
      values
    }
    logs
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
		ID          string
		Title       string
		Script      string
		Arguments   []string
		Address     string
		BlockHeight int
		Errors      []model.ProgramError
		Events      []struct {
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

type CreateContractTemplateResponse struct {
	CreateContractTemplate ContractTemplate
}

const QueryGetScriptTemplate = `
query($templateId: UUID!, $projectId: UUID!) {
  scriptTemplate(id: $templateId, projectId: $projectId) {
    id
    script
  }
}
`

const QueryGetContractTemplate = `
query($templateId: UUID!, $projectId: UUID!) {
  contractTemplate(id: $templateId, projectId: $projectId) {
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

const QueryGetFlowJson = `
query($projectId: UUID!) {
  flowJson(projectId: $projectId)
}
`

type GetFlowJsonResponse struct {
	FlowJson string
}

// todo add tests for:
// - failed transactions with successful transactions work (bootstrap works)??
// - assert we don't leak any internal model data to API
