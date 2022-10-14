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

func Test_Cache(t *testing.T) {

	t.Run("returns cached emulator", func(t *testing.T) {
		testID := uuid.New()
		c := newEmulatorCache(2)

		em, err := newEmulator()
		require.NoError(t, err)

		c.add(testID, em)

		cacheEm := c.get(testID)
		require.NotNil(t, cacheEm)

		cacheBlock, err := cacheEm.getLatestBlock()
		require.NoError(t, err)

		block, err := em.getLatestBlock()
		require.NoError(t, err)

		assert.Equal(t, block.ID(), cacheBlock.ID())
	})

	t.Run("returns cached emulator with executions", func(t *testing.T) {
		const numExecutions = 5

		var createExecutions = func(count int) []*model.TransactionExecution {
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

		testID := uuid.New()
		c := newEmulatorCache(2)

		em, err := newEmulator()
		require.NoError(t, err)

		// Add executions to emulator
		exes := createExecutions(numExecutions)
		for _, ex := range exes {
			_, _, err := em.executeTransaction(ex.Script, nil, nil)
			require.NoError(t, err)
		}

		latestBlock, err := em.getLatestBlock()
		require.NoError(t, err)

		assert.Equal(t, latestBlock.Header.Height, uint64(numExecutions))

		c.add(testID, em)

		cacheEm := c.get(testID)

		latestCacheBlock, err := cacheEm.getLatestBlock()
		require.NoError(t, err)

		// Verify cached emulator block height
		assert.Equal(t, latestCacheBlock.Header.Height, uint64(numExecutions))

		// Verify all cached emulator executions
		for i := 0; i <= numExecutions; i++ {
			block, err := em.blockchain.GetBlockByHeight(uint64(i))
			require.NoError(t, err)

			cacheBlock, err := em.blockchain.GetBlockByHeight(uint64(i))
			require.NoError(t, err)

			assert.Equal(t, block.ID(), cacheBlock.ID())
			assert.Equal(t, block.Checksum(), cacheBlock.Checksum())
		}
	})
}
