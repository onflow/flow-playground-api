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

package blockchain

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const accountsNumber = 5

var store storage.Store

func newStore() storage.Store {
	if store == nil {
		store = storage.NewSqlite()
	}
	return store
}

func newProjects() (*Projects, storage.Store) {
	store := newStore()
	chain := NewProjects(store, accountsNumber)

	return chain, store
}

func projectSeed() (*model.Project, []*model.File) {
	proj := &model.Project{
		ID:          uuid.New(),
		Secret:      uuid.New(),
		PublicID:    uuid.New(),
		ParentID:    nil,
		Seed:        123,
		Title:       "Test project title",
		Description: "we are the knights who say nii",
		Readme:      "we demand shrubbery",
		Persist:     false,
		Version:     semver.MustParse("1.0.0"),
	}

	files := make([]*model.File, 0)

	files = append(files, &model.TransactionTemplate{
		ID:        uuid.New(),
		ProjectID: proj.ID,
		Title:     "Transaction 1",
		Index:     0,
		Script:    "transaction {}",
	})

	files = append(files, &model.ScriptTemplate{
		ID:        uuid.New(),
		ProjectID: proj.ID,
		Title:     "Script 1",
		Index:     0,
		Script:    "pub fun main(): Int { return 42; }",
	})

	return proj, files
}

func newWithSeededProject() (*Projects, storage.Store, *model.Project, error) {
	projects, store := newProjects()
	proj, files := projectSeed()
	err := store.CreateProject(proj, files)

	return projects, store, proj, err
}

func Benchmark_LoadEmulator(b *testing.B) {
	projects, _, proj, _ := newWithSeededProject()

	// current run ~20 ms/op ~ 0.110s/op
	b.Run("without cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = projects.load(proj.ID)
			projects.flowKitCache.reset(proj.ID) // clear cache
		}
	})

	// current run ~15 ns/op
	b.Run("with cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = projects.load(proj.ID)
		}
	})
}

func Test_LoadFlowKit(t *testing.T) {

	t.Run("successful load of flowKit", func(t *testing.T) {
		projects, _, proj, err := newWithSeededProject()
		require.NoError(t, err)

		fk, err := projects.load(proj.ID)
		require.NoError(t, err)

		for i := 0; i < 4; i++ {
			_, err := fk.getAccount(flowsdk.HexToAddress(fmt.Sprintf("0x0%d", i+1)))
			require.NoError(t, err)
		}

		height, err := fk.getLatestBlockHeight()
		require.NoError(t, err)

		require.Equal(t, fk.initBlockHeight(), height)
	})

	t.Run("multiple loads with low cache", func(t *testing.T) {
		projects, store := newProjects()

		testProjs := make([]*model.Project, 150)

		for i := 0; i < len(testProjs); i++ {
			proj, files := projectSeed()
			err := store.CreateProject(proj, files)
			require.NoError(t, err)
			testProjs[i] = proj
		}

		for i := 0; i < len(testProjs); i++ {
			_, err := projects.load(testProjs[i].ID)
			require.NoError(t, err)
		}
	})

	t.Run("e2eTest stale cache", func(t *testing.T) {
		projects, store, proj, err := newWithSeededProject()
		require.NoError(t, err)

		_, err = projects.ExecuteTransaction(model.NewTransactionExecution{
			ProjectID: proj.ID,
			Script:    `transaction {}`,
			Signers:   nil,
			Arguments: nil,
		})
		require.NoError(t, err)

		fk, err := projects.load(proj.ID)
		require.NoError(t, err)

		// add another transaction directly to the database to simulate request coming from another replica
		err = store.InsertTransactionExecution(&model.TransactionExecution{
			BlockHeight: 2 + fk.initBlockHeight(),
			File: model.File{
				ID:        uuid.New(),
				ProjectID: proj.ID,
				Script: `transaction {
				execute {
					log("hello")
				}
			}`,
			},
		})
		require.NoError(t, err)

		fk, err = projects.load(proj.ID)
		require.NoError(t, err)

		latest, err := fk.getLatestBlockHeight()
		require.NoError(t, err)
		// there should be two blocks created, one from first execution and second from direct db execution from above
		assert.Equal(t, 2+fk.initBlockHeight(), latest)
	})

	// this tests that if another replica receives project reset, then this replica won't clear the cache,
	// so it needs to force-reset if it gets 0 executions from db even if emulator is on higher height
	t.Run("reset project on another replica", func(t *testing.T) {
		projects, store, proj, err := newWithSeededProject()
		require.NoError(t, err)

		_, err = projects.ExecuteTransaction(model.NewTransactionExecution{
			ProjectID: proj.ID,
			Script:    `transaction {}`,
			Signers:   nil,
			Arguments: nil,
		})
		require.NoError(t, err)

		err = store.ResetProjectState(proj)
		require.NoError(t, err)

		fk, err := projects.load(proj.ID)
		require.NoError(t, err)

		latest, err := fk.getLatestBlockHeight()
		require.NoError(t, err)
		assert.Equal(t, fk.initBlockHeight(), latest) // no exe since reset
	})
}

func Test_TransactionExecution(t *testing.T) {

	t.Run("successful transaction execution", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		script := `
			transaction {
				prepare (signer: AuthAccount) {} 
				execute {
					log("hello")
				}
			}`

		signers := []model.Address{
			model.NewAddressFromString("0x01"),
		}

		tx := model.NewTransactionExecution{
			ProjectID: proj.ID,
			Script:    script,
			Signers:   signers,
			Arguments: nil,
		}

		exe, err := projects.ExecuteTransaction(tx)
		require.NoError(t, err)
		require.Len(t, exe.Errors, 0)

		assert.Equal(t, proj.ID, exe.ProjectID)
		require.Len(t, exe.Logs, 1)
		assert.Equal(t, `{"level":"debug","message":"Cadence log: \"hello\""}`, exe.Logs[0])
		assert.Equal(t, script, exe.Script)
		assert.Equal(t, []string{}, exe.Arguments)
		assert.Equal(t, signers, exe.Signers)
		assert.Equal(t, 0, exe.Index)

		var dbExe []*model.TransactionExecution
		err = store.GetTransactionExecutionsForProject(proj.ID, &dbExe)
		require.NoError(t, err)

		require.Len(t, dbExe, 1)
		assert.Equal(t, exe.ID, dbExe[0].ID)
		assert.Equal(t, script, dbExe[0].Script)
	})

	t.Run("multiple transaction execution", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		script := `
			transaction {
				prepare (signer: AuthAccount) {} 
				execute {
					log("hello")
				}
			}`

		signers := []model.Address{
			model.NewAddressFromString("0x01"),
		}

		tx := model.NewTransactionExecution{
			ProjectID: proj.ID,
			Script:    script,
			Signers:   signers,
			Arguments: nil,
		}

		for i := 0; i < 5; i++ {
			exe, err := projects.ExecuteTransaction(tx)
			require.NoError(t, err)
			require.Len(t, exe.Errors, 0)

			assert.Equal(t, proj.ID, exe.ProjectID)
			require.Len(t, exe.Logs, 1)
			assert.Equal(t, `{"level":"debug","message":"Cadence log: \"hello\""}`, exe.Logs[0])
			assert.Equal(t, script, exe.Script)
			assert.Equal(t, []string{}, exe.Arguments)
			assert.Equal(t, signers, exe.Signers)
			assert.Equal(t, i, exe.Index)

			var dbExe []*model.TransactionExecution
			err = store.GetTransactionExecutionsForProject(proj.ID, &dbExe)
			require.NoError(t, err)

			require.Len(t, dbExe, i+1)
			assert.Equal(t, exe.ID, dbExe[i].ID)
			assert.Equal(t, script, dbExe[i].Script)
		}
	})

	t.Run("multiple transaction executions, reset cache", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		script := `
			transaction {
				prepare (signer: AuthAccount) {} 
				execute {
					log("hello")
				}
			}`

		signers := []model.Address{
			model.NewAddressFromString("0x01"),
		}

		tx := model.NewTransactionExecution{
			ProjectID: proj.ID,
			Script:    script,
			Signers:   signers,
			Arguments: nil,
		}

		fk, _ := projects.load(proj.ID)
		b, _ := fk.getLatestBlockHeight()
		assert.Equal(t, fk.initBlockHeight(), b)

		executeAndAssert := func(exeLen int) {
			exe, err := projects.ExecuteTransaction(tx)
			require.NoError(t, err)
			require.Len(t, exe.Errors, 0)

			var dbExe []*model.TransactionExecution
			err = store.GetTransactionExecutionsForProject(proj.ID, &dbExe)
			require.NoError(t, err)

			require.Len(t, dbExe, exeLen)

			fk, _ := projects.load(proj.ID)
			b, _ := fk.getLatestBlockHeight()
			require.Equal(t, exeLen, b-fk.initBlockHeight())

			projects.flowKitCache.reset(proj.ID)
		}

		for i := 0; i < 5; i++ {
			executeAndAssert(i + 1)
		}
	})

	t.Run("transaction with contract import and cache reset", func(t *testing.T) {
		projects, _, proj, _ := newWithSeededProject()

		scriptA := `
			pub contract HelloWorldA {
				pub var A: String
				pub init() { self.A = "HelloWorldA" }
			}`

		deployment, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), scriptA, nil)
		require.NoError(t, err)

		var deployments []*model.ContractDeployment
		err = projects.store.GetContractDeploymentsForProject(proj.ID, &deployments)
		assert.NoError(t, err)
		assert.Equal(t, deployment.Title, deployments[0].Title)

		acc, err := projects.GetAccount(proj.ID, model.NewAddressFromIndex(0))
		assert.NoError(t, err)

		assert.Equal(t, deployment.Title, acc.DeployedContracts[0])
		//assert.True(t, strings.Contains(acc.State, "HelloWorld")) //TODO: Account storage

		projects.flowKitCache.reset(proj.ID)

		script := `
			import HelloWorldA from 0x05
			transaction {
				prepare (signer: AuthAccount) {} 
				execute {
					log(HelloWorldA.A)
				}
			}`

		signers := []model.Address{model.NewAddressFromIndex(1)}

		tx := model.NewTransactionExecution{
			ProjectID: proj.ID,
			Script:    script,
			Signers:   signers,
			Arguments: nil,
		}

		exe, err := projects.ExecuteTransaction(tx)
		require.NoError(t, err)
		assert.Len(t, exe.Errors, 0)
	})

}

func Test_DeployContract(t *testing.T) {

	t.Run("deploy single contract", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		script := `pub contract HelloWorld {}`

		deployment, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), script, nil)
		require.NoError(t, err)
		assert.Equal(t, "HelloWorld", deployment.Title)

		var deployments []*model.ContractDeployment
		err = store.GetContractDeploymentsForProject(proj.ID, &deployments)
		require.NoError(t, err)
		require.Len(t, deployments, 1)

		deploy := deployments[0]
		assert.Equal(t, "flow.AccountContractAdded", deploy.Events[0].Type)
	})

	t.Run("multiple deploys with imports and cache reset", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		scriptA := `
			pub contract HelloWorldA {
				pub var A: String
				pub init() { self.A = "HelloWorldA" }
			}`

		scriptB := `
			import HelloWorldA from 0x05
			pub contract HelloWorldB {
				pub var B: String
				pub init() {
					self.B = "HelloWorldB"
					log(HelloWorldA.A) 
				}
			}`

		scriptC := `
			import HelloWorldA from 0x05
			import HelloWorldB from 0x06
			pub contract HelloWorldC {
				pub init() { 
					log(HelloWorldA.A)
					log(HelloWorldB.B)
				}
			}`

		deploy1, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), scriptA, nil)
		require.NoError(t, err)
		assert.Equal(t, "HelloWorldA", deploy1.Title)

		deploy2, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(1), scriptB, nil)
		require.NoError(t, err)
		assert.Equal(t, "HelloWorldB", deploy2.Title)

		var deployments []*model.ContractDeployment
		err = store.GetContractDeploymentsForProject(proj.ID, &deployments)
		require.NoError(t, err)
		require.Len(t, deployments, 2)

		projects.flowKitCache.reset(proj.ID)

		err = store.GetContractDeploymentsForProject(proj.ID, &deployments)
		require.NoError(t, err)
		require.Len(t, deployments, 2)

		_, err = projects.DeployContract(proj.ID, model.NewAddressFromIndex(2), scriptC, nil)
		require.NoError(t, err)

		err = store.GetContractDeploymentsForProject(proj.ID, &deployments)
		require.NoError(t, err)
		require.Len(t, deployments, 3)

		assert.Equal(t, "flow.AccountContractAdded", deployments[2].Events[0].Type)
		assert.Equal(t, "flow.AccountContractAdded", deployments[1].Events[0].Type)
		assert.Equal(t, "flow.AccountContractAdded", deployments[0].Events[0].Type)

		assert.Equal(t,
			`{"level":"debug","message":"Cadence log: \"HelloWorldA\""}`,
			deployments[1].Logs[0])

		assert.Equal(t,
			`{"level":"debug","message":"Cadence log: \"HelloWorldB\""}`,
			deployments[2].Logs[1])
	})

	t.Run("deploy single contract with arguments", func(t *testing.T) {
		projects, _, proj, _ := newWithSeededProject()

		const contract = `
		pub contract HelloWorld {
			pub var A: Int
			pub init(a: Int) { self.A = a }
		}`

		args := []string{
			`{"type":"Int","value":"42"}`,
		}

		_, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), contract, nil)
		require.Error(t, err)

		deployment, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), contract, args)
		require.NoError(t, err)
		require.Equal(t, args, deployment.Arguments)
	})

	t.Run("deploy contract with new import syntax", func(t *testing.T) {
		projects, _, proj, _ := newWithSeededProject()

		const contract = `
		pub contract HelloWorld {
			pub var A: Int
			pub init() { self.A = 5 }
		}`

		const importContract = `
		import "HelloWorld"

		pub contract Test {
			pub var B: Int
			pub init() { self.B = HelloWorld.A }
		}`

		_, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), contract, nil)
		require.NoError(t, err)

		_, err = projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), importContract, nil)
		require.NoError(t, err)
	})

	t.Run("import core contracts", func(t *testing.T) {
		projects, _, proj, _ := newWithSeededProject()

		const contract = `
		import "FungibleToken"
		import "MetadataViews"
		import "NonFungibleToken"
		
		pub contract Test {}`

		_, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), contract, nil)
		require.NoError(t, err)
	})
}

func Test_ScriptExecution(t *testing.T) {

	t.Run("single script execution", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		script := `pub fun main(): Int { 
			log("purpose")
			log("haha")
			log("test")
			return 42 
		}`

		scriptExe := model.NewScriptExecution{
			ProjectID: proj.ID,
			Script:    script,
			Arguments: nil,
		}

		exe, err := projects.ExecuteScript(scriptExe)
		require.NoError(t, err)
		assert.Len(t, exe.Errors, 0)
		assert.Equal(t, `{"level":"debug","message":"Cadence log: \"purpose\""}`, exe.Logs[0])
		assert.Equal(t, "42", exe.Value)
		assert.Equal(t, proj.ID, exe.ProjectID)

		var dbScripts []*model.ScriptExecution
		err = store.GetScriptExecutionsForProject(proj.ID, &dbScripts)
		require.NoError(t, err)

		require.Len(t, dbScripts, 1)
		assert.Equal(t, dbScripts[0].Script, script)
	})

	t.Run("script execution importing deployed contract, with cache reset", func(t *testing.T) {
		projects, _, proj, _ := newWithSeededProject()

		scriptA := `
			pub contract HelloWorldA {
				pub var A: String
				pub init() { self.A = "HelloWorldA" }
			}`

		_, err := projects.DeployContract(proj.ID, model.NewAddressFromIndex(0), scriptA, nil)
		require.NoError(t, err)

		script := `
			import HelloWorldA from 0x05
			pub fun main(): String { 
				return HelloWorldA.A
			}`

		scriptExe := model.NewScriptExecution{
			ProjectID: proj.ID,
			Script:    script,
			Arguments: nil,
		}

		exe, err := projects.ExecuteScript(scriptExe)
		require.NoError(t, err)
		assert.Equal(t, "\"HelloWorldA\"", exe.Value)
	})

	t.Run("script with arguments", func(t *testing.T) {
		projects, _, proj, _ := newWithSeededProject()

		script := `pub fun main(a: Int): Int { 
			return a
		}`

		scriptExe := model.NewScriptExecution{
			ProjectID: proj.ID,
			Script:    script,
			Arguments: []string{"{\"type\":\"Int\",\"value\":\"42\"}"},
		}

		exe, err := projects.ExecuteScript(scriptExe)
		require.NoError(t, err)
		assert.Equal(t, exe.Value, "42")
	})

}

func Benchmark_GetAccounts(b *testing.B) {
	projects, _, proj, _ := newWithSeededProject()

	addresses := make([]model.Address, 5)
	for i := 0; i < 5; i++ {
		addresses[i] = model.NewAddressFromIndex(i)
	}

	b.Run("get batch accounts", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := projects.GetAccounts(proj.ID, addresses)
			assert.NoError(b, err)
		}
	})
}
