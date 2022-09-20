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
	"testing"
)

func Test_Mutex(t *testing.T) {
	mutex := newMutex()

	testUuid := uuid.New()

	m := mutex.load(testUuid)
	m.Lock()

	v, _ := mutex.muCounter.Load(testUuid)
	assert.Equal(t, 1, v.(int))

	_, exists := mutex.mu.Load(testUuid)
	assert.True(t, exists)

	m1 := mutex.load(testUuid)
	locked := m1.TryLock()
	// should fail since we already have one lock
	assert.False(t, locked)

	v, _ = mutex.muCounter.Load(testUuid)
	assert.Equal(t, 2, v.(int))

	mutex.remove(testUuid).Unlock()

	v, _ = mutex.muCounter.Load(testUuid)
	assert.Equal(t, 1, v.(int))

	locked = m1.TryLock()
	assert.True(t, locked) // should succeed now

	mutex.remove(testUuid).Unlock()

	// after all locks are released there shouldn't be any counter left
	_, found := mutex.muCounter.Load(testUuid)
	assert.False(t, found)

	_, found = mutex.mu.Load(testUuid)
	assert.False(t, found)
}
