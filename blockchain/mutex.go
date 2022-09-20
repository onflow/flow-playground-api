package blockchain

import (
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"sync"
)

func newMutex() *mutex {
	return &mutex{}
}

// mutex contains locking logic for projects.
//
// this custom implementation of mutex creates per project ID mutex lock, this is needed because
// we need to restrict access to common resource (emulator) based on project ID, and we can not put a mutex lock
// on the emulator instance since it takes time to load the emulator in the first place, during which racing conditions may occur.
// Mutex keeps a map of mutex locks per project ID, and it also keeps a track of obtained locks per that ID so it can, after all
// the locks have been released remove that lock from the mutex map to not pollute memory.
type mutex struct {
	mu        sync.Map
	muCounter sync.Map
}

// load retrieves the mutex lock by the project ID and increase the usage counter.
func (m *mutex) load(uuid uuid.UUID) *sync.RWMutex {
	counter, _ := m.muCounter.LoadOrStore(uuid, 0)
	m.muCounter.Store(uuid, counter.(int)+1)

	mu, _ := m.mu.LoadOrStore(uuid, &sync.RWMutex{})
	return mu.(*sync.RWMutex)
}

// remove returns the mutex lock by the project ID and decreases usage counter, deleting the map entry if at 0.
func (m *mutex) remove(uuid uuid.UUID) *sync.RWMutex {
	mu, ok := m.mu.Load(uuid)
	if !ok {
		sentry.CaptureMessage("trying to access non-existing mutex")
	}

	counter, ok := m.muCounter.Load(uuid)
	if !ok {
		sentry.CaptureMessage("trying to access non-existing mutex counter")
	}

	if counter == 0 {
		m.mu.Delete(uuid)
		m.muCounter.Delete(uuid)
	} else {
		m.muCounter.Store(uuid, counter.(int)-1)
	}

	return mu.(*sync.RWMutex)
}
