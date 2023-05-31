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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func Test_InstancePool(t *testing.T) {

	t.Run("get single instance", func(t *testing.T) {
		pool := newFlowKitPool(2)
		fk, err := pool.new()
		require.NoError(t, err)
		h, err := fk.getLatestBlockHeight()
		require.NoError(t, err)
		assert.Equal(t, 5, h) // confirm functioning flowKit
	})

	t.Run("drain out the pool", func(t *testing.T) {
		pool := newFlowKitPool(3)

		for i := 0; i < 5; i++ {
			fk, err := pool.new()
			require.NoError(t, err)
			h, err := fk.getLatestBlockHeight()
			require.NoError(t, err)
			assert.Equal(t, 5, h) // confirm functioning flowKit
		}
	})

	t.Run("concurrently access pool", func(t *testing.T) {
		pool := newFlowKitPool(5)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				fk, err := pool.new()
				require.NoError(t, err)
				h, err := fk.getLatestBlockHeight()
				require.NoError(t, err)
				assert.Equal(t, 5, h) // confirm functioning flowKit
				wg.Done()
			}()
		}

		wg.Wait()
	})
}
