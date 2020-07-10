package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	legacyauth "github.com/dapperlabs/flow-playground-api/auth/legacy"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/sessions"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Authenticator struct {
	store storage.Store
}

func NewAuthenticator(store storage.Store) *Authenticator {
	return &Authenticator{
		store: store,
	}
}

func (a *Authenticator) GetOrCreateUser(ctx context.Context) (*model.User, error) {
	session := sessions.Get(ctx, "flow-playground")

	var user *model.User
	var err error

	if session.IsNew {
		user, err = a.createNewUserWithSessionID(session.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new user")
		}
	} else {
		user, err = a.getCurrentUser(session.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load user from session")
		}
	}

	err = sessions.Save(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update session")
	}

	return user, nil
}

func (a *Authenticator) CheckProjectAccess(ctx context.Context, proj *model.InternalProject) error {
	var user *model.User
	var err error

	session := sessions.Get(ctx, "flow-playground")

	if !session.IsNew {
		user, err = a.getCurrentUser(session.ID)
		if err != nil {
			return errors.New("access denied")
		}
	}

	if a.hasProjectAccess(user, proj) {
		err = sessions.Save(ctx, session)
		if err != nil {
			return errors.New("access denied")
		}

		return nil
	}

	if a.hasLegacyProjectAccess(ctx, proj) {
		err = a.migrateLegacyProjectAccess(user, session.ID, proj)
		if err != nil {
			return errors.New("access denied")
		}

		err = sessions.Save(ctx, session)
		if err != nil {
			return errors.New("access denied")
		}

		return nil
	}

	return errors.New("access denied")
}

func (a *Authenticator) getCurrentUser(sessionID string) (*model.User, error) {
	var user model.User

	err := a.store.GetUserBySessionID(sessionID, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *Authenticator) hasProjectAccess(user *model.User, proj *model.InternalProject) bool {
	return user != nil && proj.IsOwnedBy(user.ID)
}

func (a *Authenticator) hasLegacyProjectAccess(ctx context.Context, proj *model.InternalProject) bool {
	return legacyauth.ProjectInSession(ctx, proj)
}

func (a *Authenticator) migrateLegacyProjectAccess(user *model.User, sessionID string, proj *model.InternalProject) error {
	var err error

	if user == nil {
		user, err = a.createNewUserWithSessionID(sessionID)
		if err != nil {
			return errors.Wrap(err, "failed to create new user")
		}
	}

	err = a.store.UpdateProjectOwner(proj.ID, user.ID)
	if err != nil {
		return errors.Wrap(err, "failed to update project owner")
	}

	return nil
}

func (a *Authenticator) createNewUserWithSessionID(sessionID string) (*model.User, error) {
	user := &model.User{
		ID:               uuid.New(),
		CurrentSessionID: &sessionID,
	}

	err := a.store.InsertUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
