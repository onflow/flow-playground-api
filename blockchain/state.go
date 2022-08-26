package blockchain

import (
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
)

// todo create instance pool as a possible optimization: we can pre-instantiate empty instances of emulators waiting around to be assigned to a project if init time will be proved to be an issue

// access cache and add to cache

// bootstrap emulator with transactions

type State struct {
	store *storage.Store
	// cache
}

func (s *State) bootstrap(ID uuid.UUID) {
	emulator, err := NewEmulator()
}

func (s *State) new(ID uuid.UUID) {
	emulator, err := NewEmulator()

}
