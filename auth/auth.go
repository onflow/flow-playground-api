package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	legacyauth "github.com/dapperlabs/flow-playground-api/auth/legacy"
	"github.com/dapperlabs/flow-playground-api/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Authenticator struct {
	store       storage.Store
	sessionName string
}

const userIDKey = "userID"

func NewAuthenticator(store storage.Store, sessionName string) *Authenticator {
	return &Authenticator{
		store:       store,
		sessionName: sessionName,
	}
}

func (a *Authenticator) GetOrCreateUser(ctx context.Context) (*model.User, error) {
	session := sessions.Get(ctx, a.sessionName)

	var user *model.User
	var err error

	if session.IsNew {
		user, err = a.createNewUser()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new user")
		}

		session.Values[userIDKey] = user.ID.String()
	} else {
		user, err = a.getCurrentUser(session.Values[userIDKey].(string))
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

	session := sessions.Get(ctx, a.sessionName)

	if !session.IsNew {
		user, err = a.getCurrentUser(session.Values[userIDKey].(string))
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
		return nil, err
	}

	err = a.store.GetUser(userID, &user)
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

func (a *Authenticator) migrateLegacyProjectAccess(user *model.User, proj *model.InternalProject) (*model.User, error) {
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
		return nil, err
	}

	return user, nil
}
