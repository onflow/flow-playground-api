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
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/require"

	"github.com/Masterminds/semver"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
)

func newProjects() (*Projects, *memory.Store) {
	store := memory.NewStore()
	chain := NewProjects(store, lru.New(128))

	return chain, store
}

func projectSeed() (*model.InternalProject, []*model.TransactionTemplate, []*model.ScriptTemplate) {
	proj := &model.InternalProject{
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

	txTpls := make([]*model.TransactionTemplate, 0)
	txTpls = append(txTpls, &model.TransactionTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: proj.ID,
		},
		Title:  "Transaction 1",
		Index:  0,
		Script: "transaction {}",
	})

	scriptTpls := make([]*model.ScriptTemplate, 0)
	scriptTpls = append(scriptTpls, &model.ScriptTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: proj.ID,
		},
		Title:  "Script 1",
		Index:  0,
		Script: "pub fun main(): Int { return 42; }",
	})

	return proj, txTpls, scriptTpls
}

func newWithSeededProject() (*Projects, *memory.Store, *model.InternalProject, error) {
	projects, store := newProjects()
	proj, txTpls, scriptTpls := projectSeed()
	err := store.CreateProject(proj, txTpls, scriptTpls)

	return projects, store, proj, err
}

func Test_LoadEmulator(t *testing.T) {

	t.Run("successful load of emulator", func(t *testing.T) {
		projects, _, proj, err := newWithSeededProject()
		require.NoError(t, err)

		emulator, err := projects.load(proj.ID)
		require.NoError(t, err)

		for i := 0; i < 4; i++ {
			_, _, err := emulator.getAccount(flowsdk.HexToAddress(fmt.Sprintf("0x0%d", i+1)))
			require.NoError(t, err)
		}

		block, err := emulator.getLatestBlock()
		require.NoError(t, err)

		require.Equal(t, uint64(0), block.Header.Height)
	})

	t.Run("multiple loads with low cache", func(t *testing.T) {
		projects, store := newProjects()
		projects.cache = lru.New(2)

		testProjs := make([]*model.InternalProject, 10)

		for i := 0; i < len(testProjs); i++ {
			proj, txTpls, scriptTpls := projectSeed()
			err := store.CreateProject(proj, txTpls, scriptTpls)
			require.NoError(t, err)
			testProjs[i] = proj
		}

		for i := 0; i < len(testProjs); i++ {
			_, err := projects.load(testProjs[i].ID)
			require.NoError(t, err)
		}
	})
}

func Benchmark_LoadEmulator(b *testing.B) {
	projects, _, proj, _ := newWithSeededProject()

	// current run ~110 000 000 ns/op ~ 0.110s/op
	b.Run("without cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = projects.load(proj.ID)
			projects.cache.Remove(proj.ID) // clear cache
		}
	})

	// current run ~15 ns/op
	b.Run("with cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = projects.load(proj.ID)
		}
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
		assert.Equal(t, `"hello"`, exe.Logs[0])
		assert.Equal(t, script, exe.Script)
		assert.Equal(t, []string{}, exe.Arguments)
		assert.Equal(t, signers, exe.Signers)

		var dbExe []*model.TransactionExecution
		err = store.GetTransactionExecutionsForProject(proj.ID, &dbExe)
		require.NoError(t, err)

		require.Len(t, dbExe, 1)
		assert.Equal(t, exe.ID, dbExe[0].ID)
		assert.Equal(t, script, dbExe[0].Script)
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

		executeAndAssert := func(exeLen int) {
			exe, err := projects.ExecuteTransaction(tx)
			require.NoError(t, err)
			require.Len(t, exe.Errors, 0)

			var dbExe []*model.TransactionExecution
			err = store.GetTransactionExecutionsForProject(proj.ID, &dbExe)
			require.NoError(t, err)

			require.Len(t, dbExe, exeLen)

			em, _ := projects.load(proj.ID)
			b, _ := em.getLatestBlock()
			require.Equal(t, b.Header.Height, uint64(exeLen))

			projects.cache.Remove(proj.ID)
		}

		for i := 0; i < 5; i++ {
			executeAndAssert(i + 1)
		}
	})

}

func Test_AccountCreation(t *testing.T) {
	t.Run("successful account creation", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		account, err := projects.CreateAccount(proj.ID)
		require.NoError(t, err)
		assert.Equal(t, "0000000000000005", account.Address.ToFlowAddress().String())
		assert.Equal(t, "", account.DraftCode)
		assert.Len(t, account.DeployedContracts, 0)
		assert.Equal(t, "", account.DeployedCode)
		assert.Equal(t, "", account.State)
		assert.Equal(t, proj.ID, account.ProjectID)

		var executions []*model.TransactionExecution
		err = store.GetTransactionExecutionsForProject(proj.ID, &executions)
		require.NoError(t, err)

		require.Len(t, executions, 1)
		assert.Len(t, executions[0].Errors, 0)
		assert.True(t, strings.Contains(executions[0].Script, "AuthAccount(payer: signer)"))
	})

	t.Run("multiple account creations, reset cache", func(t *testing.T) {
		projects, store, proj, _ := newWithSeededProject()

		createAndAssert := func(createNumber int) {
			account, err := projects.CreateAccount(proj.ID)
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("000000000000000%d", createNumber+4), account.Address.ToFlowAddress().String())

			projects.cache.Remove(proj.ID)

			var executions []*model.TransactionExecution
			err = store.GetTransactionExecutionsForProject(proj.ID, &executions)
			require.NoError(t, err)
			require.Len(t, executions, createNumber)
		}

		for i := 0; i < 5; i++ {
			createAndAssert(i + 1)
		}
	})
}

func Test_ConcurrentRequests(t *testing.T) {

	testConcurrently := func(
		numOfRequests int,
		request func(i int, ch chan any, wg *sync.WaitGroup, projects *Projects, proj *model.InternalProject),
		test func(ch chan any, proj *model.InternalProject),
	) {
		projects, _, proj, _ := newWithSeededProject()

		ch := make(chan any)
		var wg sync.WaitGroup

		wg.Add(numOfRequests)

		for i := 0; i < numOfRequests; i++ {
			go request(i, ch, &wg, projects, proj)
		}

		test(ch, proj)

		wg.Wait()
	}

	t.Run("concurrent account creation", func(t *testing.T) {
		const numOfRequests = 4

		createAccount := func(i int, ch chan any, wg *sync.WaitGroup, projects *Projects, proj *model.InternalProject) {
			defer wg.Done()

			acc, err := projects.CreateAccount(proj.ID)
			require.NoError(t, err)

			ch <- acc
		}

		testAccount := func(ch chan any, proj *model.InternalProject) {
			accounts := make([]*model.Account, 0)
			for a := range ch {
				account := a.(*model.Account)
				accounts = append(accounts, account)

				if len(accounts) == numOfRequests {
					close(ch)
				}
			}

			require.Len(t, accounts, numOfRequests)

			addresses := make([]string, numOfRequests)
			for i, acc := range accounts {
				assert.Equal(t, proj.ID, acc.ProjectID)
				addr := acc.Address.ToFlowAddress().String()
				assert.NotContains(t, addresses, addr) // make sure unique address is returned
				addresses[i] = addr
			}
		}

		t.Run("with cache", func(t *testing.T) {
			testConcurrently(numOfRequests, createAccount, testAccount)
		})

		t.Run("without cache", func(t *testing.T) {
			createAccountNoCache := func(i int, ch chan any, wg *sync.WaitGroup, projects *Projects, proj *model.InternalProject) {
				defer wg.Done()

				acc, err := projects.CreateAccount(proj.ID)
				require.NoError(t, err)

				projects.cache.Remove(proj.ID)

				ch <- acc
			}

			testConcurrently(numOfRequests, createAccountNoCache, testAccount)
		})

	})

}
