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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func Test_Mutex(t *testing.T) {
	t.Skip()

}

func Test_ConcurrentAccess(t *testing.T) {
	mu := newMutex()
	testID := uuid.New()

	// simulate shared memory access
	shared := 0

	const subCount = 20
	wg := sync.WaitGroup{}
	wg.Add(subCount)

	uniques := make([]int, subCount)
	for i := 0; i < subCount; i++ {
		go func(x int) {
			mu.load(testID).Lock()
			defer mu.remove(testID).Unlock()

			shared += 1
			time.Sleep(time.Duration(rand.Intn(subCount)) * time.Millisecond) // make sure first routine lasts longer then to shortest
			uniques[x] = shared
			wg.Done()
		}(i)
	}

	wg.Wait()

	visited := make(map[int]bool)
	for _, u := range uniques {
		assert.False(t, visited[u])
		visited[u] = true
	}
}
