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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_InstancePool(t *testing.T) {

	t.Run("get single instance", func(t *testing.T) {
		pool := newEmulatorPool(2)
		em, err := pool.new()
		require.NoError(t, err)
		h, err := em.getLatestBlockHeight()
		require.NoError(t, err)
		assert.Equal(t, 0, h) // confirm functioning emulator
	})

	t.Run("drain out the pool", func(t *testing.T) {
		pool := newEmulatorPool(3)

		for i := 0; i < 5; i++ {
			em, err := pool.new()
			require.NoError(t, err)
			h, err := em.getLatestBlockHeight()
			require.NoError(t, err)
			assert.Equal(t, 0, h) // confirm functioning emulator
		}
	})

	t.Run("concurrently access pool", func(t *testing.T) {
		pool := newEmulatorPool(5)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				em, err := pool.new()
				require.NoError(t, err)
				h, err := em.getLatestBlockHeight()
				require.NoError(t, err)
				assert.Equal(t, 0, h) // confirm functioning emulator
				wg.Done()
			}()
		}

		wg.Wait()
	})
}
