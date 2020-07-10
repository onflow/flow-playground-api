package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/sessions"
	legacysessions "github.com/dapperlabs/flow-playground-api/sessions/legacy"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Authenticator struct {
	sessionManager *sessions.Manager
	store          storage.Store
}

func NewAuthenticator(sessionManager *sessions.Manager, store storage.Store) *Authenticator {
	return &Authenticator{
		sessionManager: sessionManager,
		store:          store,
	}
}

func (a *Authenticator) GetOrCreateUser(ctx context.Context) (*model.User, error) {
	sessionID, err := a.sessionManager.CurrentSessionID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load session")
	}

	if sessionID != uuid.Nil {
		var user model.User
		err := a.store.GetUserBySessionID(sessionID, &user)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load user from session")
		}

		err = a.sessionManager.SaveSession(ctx, sessionID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update session")
		}

		return &user, nil
	}

	sessionID = uuid.New()
	user, err := a.createNewUserWithSessionID(sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new user")
	}

	err = a.sessionManager.SaveSession(ctx, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update session")
	}

	return user, nil
}

func (a *Authenticator) CheckProjectAccess(ctx context.Context, proj *model.InternalProject) error {
	var user *model.User
	var err error

	user, err = a.getCurrentUser(ctx)
	if err != nil {
		return errors.New("access denied")
	}

	if a.hasProjectAccess(user, proj) {
		err = a.sessionManager.SaveSession(ctx, *user.CurrentSessionID)
		if err != nil {
			return errors.New("access denied")
		}

		return nil
	}

	if a.hasLegacyProjectAccess(ctx, proj) {
		sessionID, err := a.migrateLegacyProjectAccess(user, proj)
		if err != nil {
			return errors.New("access denied")
		}

		err = a.sessionManager.SaveSession(ctx, sessionID)
		if err != nil {
			return errors.New("access denied")
		}

		return nil
	}

	return errors.New("access denied")
}

func (a *Authenticator) getCurrentUser(ctx context.Context) (*model.User, error) {
	sessionID, err := a.sessionManager.CurrentSessionID(ctx)
	if err != nil {
		return nil, err
	}

	if sessionID != uuid.Nil {
		var user model.User

		err = a.store.GetUserBySessionID(sessionID, &user)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, nil
			}

			return nil, err
		}

		return &user, nil
	}

	return nil, nil
}

func (a *Authenticator) hasProjectAccess(user *model.User, proj *model.InternalProject) bool {
	return user != nil && proj.IsOwnedBy(user.ID)
}

func (a *Authenticator) hasLegacyProjectAccess(ctx context.Context, proj *model.InternalProject) bool {
	return legacysessions.ProjectInSession(ctx, proj)
}

func (a *Authenticator) migrateLegacyProjectAccess(user *model.User, proj *model.InternalProject) (uuid.UUID, error) {
	var err error

	if user == nil {
		sessionID := uuid.New()

		user, err = a.createNewUserWithSessionID(sessionID)
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "failed to create new user")
		}
	}

	err = a.store.UpdateProjectOwner(proj.ID, user.ID)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to update project owner")
	}

	return *user.CurrentSessionID, nil
}

func (a *Authenticator) createNewUserWithSessionID(sessionID uuid.UUID) (*model.User, error) {
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
