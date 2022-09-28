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
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func createExecutions(count int) []*model.TransactionExecution {
	executions := make([]*model.TransactionExecution, count)
	for i := 0; i < count; i++ {
		executions[i] = &model.TransactionExecution{
			ID:        uuid.New(),
			ProjectID: uuid.New(),
			Index:     i,
			Script:    fmt.Sprintf(`transaction { execute { log(%d) } }`, i),
		}
	}
	return executions
}

func Test_Cache(t *testing.T) {

	t.Run("returns cached emulator", func(t *testing.T) {
		testID := uuid.New()
		c := newEmulatorCache(2)

		em, err := newEmulator()
		require.NoError(t, err)

		c.add(testID, em)

		cacheEm, found := c.get(testID)
		require.True(t, found)

		cacheBlock, err := cacheEm.getLatestBlock()
		require.NoError(t, err)

		block, err := em.getLatestBlock()
		require.NoError(t, err)

		assert.Equal(t, block.ID(), cacheBlock.ID())
	})

	/* todo move to project test
	t.Run("returns cached emulator with executions", func(t *testing.T) {
		testID := uuid.New()
		c := newEmulatorCache(2)

		em, err := newEmulator()
		require.NoError(t, err)

		c.add(testID, em)

		executions := createExecutions(5)
		for _, exe := range executions {
			_, _, err := em.executeTransaction(exe.Script, exe.Arguments, nil)
			require.NoError(t, err)
		}

		cachedEm, found := c.get(testID)
		require.True(t, found)

		// cached emulator contains all the executions
		assert.Len(t, cacheExe, 0)
		// make sure emulators are same
		cacheBlock, _ := cachedEm.getLatestBlock()
		block, _ := em.getLatestBlock()
		assert.Equal(t, cacheBlock.ID(), block.ID())
	})

	t.Run("returns cached emulator with missing executions", func(t *testing.T) {
		testID := uuid.New()
		c := newEmulatorCache(2)

		em, err := newEmulator()
		require.NoError(t, err)

		c.add(testID, em)

		executions := createExecutions(5)

		for i, exe := range executions {
			if i == 3 {
				break // miss last two executions
			}
			_, _, err := em.executeTransaction(exe.Script, exe.Arguments, nil)
			require.NoError(t, err)
		}

		_, cacheExe, err := c.get(testID, executions)
		require.NoError(t, err)

		// cached emulator missed two executions
		assert.Len(t, cacheExe, 2)
		assert.Equal(t, 3, cacheExe[0].Index)
		assert.Equal(t, 4, cacheExe[1].Index)
	})

	*/
}
