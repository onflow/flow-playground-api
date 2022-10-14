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
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"sync"
)

func newMutex() *mutex {
	return &mutex{
		mx:     &sync.RWMutex{},
		pMutex: map[uuid.UUID]*sync.RWMutex{},
	}
}

// todo mutex has been simplified to not remove mutexes from project ID map, this may grow with time but for now it removes complexity

// mutex contains locking logic for projects.
//
// this custom implementation of mutex creates per project ID mutex lock, this is needed because
// we need to restrict access to common resource (emulator) based on project ID, and we can not put a mutex lock
// on the emulator instance since it takes time to load the emulator in the first place, during which racing conditions may occur.
// Mutex keeps a map of mutex locks per project ID, and it also keeps a track of obtained locks per that ID so it can, after all
// the locks have been released remove that lock from the mutex map to not pollute memory.
type mutex struct {
	mx     *sync.RWMutex               // mutex for access to bellow maps
	pMutex map[uuid.UUID]*sync.RWMutex // per project mutexes
}

// load retrieves the mutex lock by the project ID and increase the usage counter.
func (m *mutex) load(uuid uuid.UUID) *sync.RWMutex {
	m.mx.Lock()
	defer m.mx.Unlock()

	if _, ok := m.pMutex[uuid]; !ok {
		m.pMutex[uuid] = &sync.RWMutex{}
	}

	return m.pMutex[uuid]
}

// remove returns the mutex lock by the project ID and decreases usage counter, deleting the map entry if at 0.
func (m *mutex) remove(uuid uuid.UUID) *sync.RWMutex {
	m.mx.Lock()
	defer m.mx.Unlock()

	mut, ok := m.pMutex[uuid]
	if !ok {
		sentry.CaptureMessage(fmt.Sprintf("trying to remove a mutex it doesn't exists, project ID: %s", uuid))
	}

	return mut
}
