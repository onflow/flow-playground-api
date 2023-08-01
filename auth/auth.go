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

package auth

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	legacyauth "github.com/dapperlabs/flow-playground-api/auth/legacy"
	"github.com/dapperlabs/flow-playground-api/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

// An Authenticator manages user authentication for the Playground API.
//
// An authenticator instance can create new users, load existing users from session
// data, and check for project access.
type Authenticator struct {
	store       storage.Store
	sessionName string
}

// NewAuthenticator returns a new authenticator instance.
//
// User data is stored in and loaded from the provided store instance.
//
// The session name parameter is used as the cookie name when setting session
// data on the client.
func NewAuthenticator(store storage.Store, sessionName string) *Authenticator {
	return &Authenticator{
		store:       store,
		sessionName: sessionName,
	}
}

const userIDKey = "userID"

// GetOrCreateUser gets an existing user from the current session or creates a
// new user and session if a session does not already exist.
func (a *Authenticator) GetOrCreateUser(ctx context.Context) (*model.User, error) {
	session := sessions.Get(ctx, a.sessionName)

	var user *model.User
	var err error

	userLoaded := false

	if !session.IsNew {
		// Try to load existing user
		user, err = a.getCurrentUser(session.Values[userIDKey].(string))
		if err == nil {
			fmt.Printf("Failed to load user id %s from session\n", session.Values[userIDKey].(string))
			userLoaded = true
		}
	}

	if !userLoaded {
		// Create new user
		user, err = a.createNewUser()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new user")
		}

		session.Values[userIDKey] = user.ID.String()
	}

	err = sessions.Save(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update session")
	}

	return user, nil
}

// CheckProjectAccess returns an error if the current user is not authorized to mutate
// the provided project.
//
// This function checks for access using both the new and legacy authentication schemes. If
// a user has legacy access, their authentication is then migrated to use the new scheme.
func (a *Authenticator) CheckProjectAccess(ctx context.Context, proj *model.Project) error {
	var user *model.User
	var err error

	session := sessions.Get(ctx, a.sessionName)

	user, err = a.getCurrentUser(session.Values[userIDKey].(string))
	if err != nil {
		fmt.Println("Failed to get current user: ", err.Error())
		return errors.New("access denied")
	}

	if a.hasProjectAccess(user, proj) {
		err = sessions.Save(ctx, session)
		if err != nil {
			fmt.Println("Failed to Save Session: ", err.Error())
			return errors.New("access denied")
		}

		return nil
	}
	fmt.Println("User does not have Project Access")

	if a.hasLegacyProjectAccess(ctx, proj) {
		user, err = a.migrateLegacyProjectAccess(user, proj)
		if err != nil {
			return errors.New("access denied")
		}

		session.Values[userIDKey] = user.ID.String()

		err = sessions.Save(ctx, session)
		if err != nil {
			return errors.New("access denied")
		}

		return nil
	}

	return errors.New("access denied")
}

func (a *Authenticator) getCurrentUser(userIDStr string) (*model.User, error) {
	var user model.User
	var userID uuid.UUID

	err := userID.UnmarshalText([]byte(userIDStr))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal userIDStr")
	}

	err = a.store.GetUser(userID, &user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user from db")
	}

	return &user, nil
}

func (a *Authenticator) hasProjectAccess(user *model.User, proj *model.Project) bool {
	return user != nil && proj.IsOwnedBy(user.ID)
}

func (a *Authenticator) hasLegacyProjectAccess(ctx context.Context, proj *model.Project) bool {
	return legacyauth.ProjectInSession(ctx, proj)
}

func (a *Authenticator) migrateLegacyProjectAccess(user *model.User, proj *model.Project) (*model.User, error) {
	var err error

	if user == nil {
		user, err = a.createNewUser()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new user")
		}
	}

	err = a.store.UpdateProjectOwner(proj.ID, user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update project owner")
	}

	return user, nil
}

func (a *Authenticator) createNewUser() (*model.User, error) {
	user := &model.User{
		ID: uuid.New(),
	}

	err := a.store.InsertUser(user)
	if err != nil {
		return nil, errors.Wrap(err, "could not insert the user")
	}

	return user, nil
}
