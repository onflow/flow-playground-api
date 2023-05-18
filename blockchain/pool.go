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

// newFlowKitPool creates new instance of instance pool with provided size.
func newFlowKitPool(size int) *flowKitPool {
	pool := &flowKitPool{
		instances: make(chan *flowKit, size),
	}

	for i := 0; i < size; i++ {
		go pool.create()
	}

	return pool
}

// flowKitPool is an instance pool that optimize slow init time of emulators and hence prepare them upfront.
//
// This is an optimization trick to avoid waiting for slow bootstrap time of a new emulator once it's needed,
// we instead prepare bootstrapped emulators ahead of time.
type flowKitPool struct {
	instances chan *flowKit
}

// new returns a new emulator instance from the instance pool.
func (p *flowKitPool) new() (*flowKit, error) {
	select {
	case em := <-p.instances:
		go p.create()
		return em, nil
	default: // in case pool gets emptied
		sentry.CaptureMessage("instance pool empty")
		return newFlowkit()
	}
}

// add an emulator to internal instance pool, only to be used internally.
func (p *flowKitPool) add(fk *flowKit) {
	select {
	case p.instances <- fk:
	default:
		// instance pool is full, shouldn't happen since we take one and create one, but deadlock prevention if it does
		return
	}
}

// create a new emulator for internal instance pool, only to be used internally.
func (p *flowKitPool) create() {
	em, err := newFlowkit()
	if err != nil {
		sentry.CaptureException(errors.Wrap(err, "instance pool emulator creation failure"))
		return
	}
	p.add(em)
}
