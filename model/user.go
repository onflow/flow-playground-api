package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type User struct {
	ID               uuid.UUID
	CurrentSessionID *uuid.UUID
}

func (u *User) NameKey() *datastore.Key {
	return datastore.NameKey("User", u.ID.String(), nil)
}

func (u *User) Load(ps []datastore.Property) error {
	tmp := struct {
		ID               string
		CurrentSessionID *string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := u.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	if tmp.CurrentSessionID != nil && len(*tmp.CurrentSessionID) != 0 {
		u.CurrentSessionID = new(uuid.UUID)
		if err := u.CurrentSessionID.UnmarshalText([]byte(*tmp.CurrentSessionID)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	} else {
		u.CurrentSessionID = nil
	}

	return nil
}

func (u *User) Save() ([]datastore.Property, error) {
	currentSessionID := new(string)
	if u.CurrentSessionID != nil {
		*currentSessionID = (*u.CurrentSessionID).String()
	}

	return []datastore.Property{
		{
			Name:  "ID",
			Value: u.ID.String(),
		},
		{
			Name:  "CurrentSessionID",
			Value: currentSessionID,
		},
	}, nil
}
