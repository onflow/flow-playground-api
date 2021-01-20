/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type User struct {
	ID uuid.UUID
}

func (u *User) NameKey() *datastore.Key {
	return datastore.NameKey("User", u.ID.String(), nil)
}

func (u *User) Load(ps []datastore.Property) error {
	tmp := struct {
		ID string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := u.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	return nil
}

func (u *User) Save() ([]datastore.Property, error) {
	return []datastore.Property{
		{
			Name:  "ID",
			Value: u.ID.String(),
		},
	}, nil
}
