package blockchain

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"sync"
)

func newMutex() *mutex {
	return &mutex{
		mx:      &sync.RWMutex{},
		counter: map[uuid.UUID]int{},
		pMutex:  map[uuid.UUID]*sync.RWMutex{},
	}
}

// mutex contains locking logic for projects.
//
// this custom implementation of mutex creates per project ID mutex lock, this is needed because
// we need to restrict access to common resource (emulator) based on project ID, and we can not put a mutex lock
// on the emulator instance since it takes time to load the emulator in the first place, during which racing conditions may occur.
// Mutex keeps a map of mutex locks per project ID, and it also keeps a track of obtained locks per that ID so it can, after all
// the locks have been released remove that lock from the mutex map to not pollute memory.
type mutex struct {
	mx      *sync.RWMutex               // mutex for access to bellow maps
	pMutex  map[uuid.UUID]*sync.RWMutex // per project mutexes
	counter map[uuid.UUID]int           // per project counter
}

// load retrieves the mutex lock by the project ID and increase the usage counter.
func (m *mutex) load(uuid uuid.UUID) *sync.RWMutex {
	return m.mx
	m.mx.Lock()
	defer m.mx.Unlock()

	if _, ok := m.pMutex[uuid]; !ok {
		m.pMutex[uuid] = &sync.RWMutex{}
	}

	m.counter[uuid] += 1

	return m.pMutex[uuid]
}

// remove returns the mutex lock by the project ID and decreases usage counter, deleting the map entry if at 0.
func (m *mutex) remove(uuid uuid.UUID) *sync.RWMutex {
	return m.mx
	m.mx.Lock()
	defer m.mx.Unlock()

	mut, ok := m.pMutex[uuid]
	if !ok {
		sentry.CaptureMessage(fmt.Sprintf("trying to remove a mutex it doesn't exists, project ID: %s", uuid))
	}

	if m.counter[uuid] == 1 {
		delete(m.counter, uuid)
		delete(m.pMutex, uuid)
	} else {
		m.counter[uuid] -= 1
	}

	return mut
}
