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

package storage

import (
	"errors"
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"time"
)

type Store interface {
	InsertUser(user *model.User) error
	GetUser(id uuid.UUID, user *model.User) error

	CreateProject(
		proj *model.Project,
		files []*model.File,
	) error
	UpdateProject(input model.UpdateProject, proj *model.Project) error
	UpdateProjectOwner(id, userID uuid.UUID) error
	UpdateProjectVersion(id uuid.UUID, version *semver.Version) error
	ResetProjectState(proj *model.Project) error
	GetProject(id uuid.UUID, proj *model.Project) error
	ProjectAccessed(id uuid.UUID) error
	GetAllProjectsForUser(userID uuid.UUID, proj *[]*model.Project) error
	GetProjectCountForUser(userID uuid.UUID, count *int64) error
	DeleteProject(id uuid.UUID) error

	GetStaleProjects(stale time.Duration, projs *[]*model.Project) error
	DeleteStaleProjects(stale time.Duration) error
	TotalProjectCount(totalProjects *int64) error

	InsertFile(file *model.File) error
	UpdateFile(input model.UpdateFile, file *model.File) error
	DeleteFile(id uuid.UUID, pID uuid.UUID) error
	GetFile(id uuid.UUID, pID uuid.UUID, file *model.File) error
	GetFilesForProject(projectID uuid.UUID, files *[]*model.File, fileType model.FileType) error
	GetAllFilesForProject(projectID uuid.UUID, files *[]*model.File) error

	InsertContractDeployment(deploy *model.ContractDeployment) error
	DeleteContractDeployment(deploy *model.ContractDeployment) error
	DeleteContractDeploymentByName(projectID uuid.UUID, address model.Address, contractName string) error
	InsertContractDeploymentWithExecution(deploy *model.ContractDeployment, exe *model.TransactionExecution) error
	GetContractDeploymentsForProject(projectID uuid.UUID, deployments *[]*model.ContractDeployment) error
	GetContractDeploymentOnAddress(projectID uuid.UUID, title string, address model.Address, deployment *model.ContractDeployment) error
	TruncateDeploymentsAndExecutionsByBlockHeight(projectID uuid.UUID, blockHeight int) error

	InsertTransactionExecution(exe *model.TransactionExecution) error
	GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error

	InsertScriptExecution(exe *model.ScriptExecution) error
	GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error

	Ping() error
}

var ErrNotFound = errors.New("entity not found")
