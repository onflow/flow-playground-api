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
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// todo possible improvement is to also create default accounts as part of bootstrap

// newEmulatorPool creates new instance of instance pool with provided size.
func newEmulatorPool(size int) (*emulatorPool, error) {
	return &emulatorPool{
		instances: make(chan *emulator, size),
	}, nil
}

// emulatorPool is an instance pool that optimize slow init time of emulators and hence prepare them upfront.
//
// This is an optimization trick to avoid waiting for slow bootstrap time of a new emulator once it's needed,
// we instead prepare bootstrapped emulators ahead of time.
type emulatorPool struct {
	instances chan *emulator
}

// new returns a new emulator instance from the instance pool.
func (e *emulatorPool) new() (*emulator, error) {
	select {
	case em := <-e.instances:
		go e.create()
		return em, nil
	default: // in case pool gets emptied
		sentry.CaptureMessage("instance pool empty")
		return newEmulator()
	}
}

// add an emulator to internal instance pool, only to be used internally.
func (e *emulatorPool) add(em *emulator) {
	select {
	case e.instances <- em:
	default:
		// instance pool is full, shouldn't happen since we take one and create one, but deadlock prevention if it does
		return
	}
}

// create a new emulator for internal instance pool, only to be used internally.
func (e *emulatorPool) create() {
	em, err := newEmulator()
	if err != nil {
		sentry.CaptureException(errors.Wrap(err, "instance pool emulator creation failure"))
	}
	e.add(em)
}
