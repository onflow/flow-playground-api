package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type InternalProject struct {
	ID               uuid.UUID
	PrivateID        uuid.UUID
	PublicID         uuid.UUID
	ParentID         *uuid.UUID
	TransactionCount int
	Persist          bool
}

// ExportPublicMutable converts the internal project to its public representation
// and marks it as mutable.
func (p *InternalProject) ExportPublicMutable() *Project {
	return &Project{
		ID:       p.ID,
		PublicID: p.PublicID,
		ParentID: p.ParentID,
		Persist:  p.Persist,
		Mutable:  true,
	}
}

// ExportPublicImmutable converts the internal project to its public representation
// and marks it as immutable.
func (p *InternalProject) ExportPublicImmutable() *Project {
	return &Project{
		ID:       p.ID,
		PublicID: p.PublicID,
		ParentID: p.ParentID,
		Persist:  p.Persist,
		Mutable:  false,
	}
}

func (p *InternalProject) NameKey() *datastore.Key {
	return datastore.NameKey("Project", p.ID.String(), nil)
}

func (p *InternalProject) Load(ps []datastore.Property) error {
	tmp := struct {
		ID               string
		PrivateID        string
		PublicID         string
		ParentID         *string
		TransactionCount int
		Persist          bool
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := p.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	if err := p.PrivateID.UnmarshalText([]byte(tmp.PrivateID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	if err := p.PublicID.UnmarshalText([]byte(tmp.PublicID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	if tmp.ParentID != nil && len(*tmp.ParentID) != 0 {
		if err := p.ParentID.UnmarshalText([]byte(*tmp.ParentID)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	} else {
		p.ParentID = nil
	}

	p.TransactionCount = tmp.TransactionCount
	p.Persist = tmp.Persist
	return nil
}

func (p *InternalProject) Save() ([]datastore.Property, error) {
	parentID := new(string)
	if p.ParentID != nil {
		*parentID = (*p.ParentID).String()
	}

	return []datastore.Property{
		{
			Name:  "ID",
			Value: p.ID.String(),
		},
		{
			Name:  "PrivateID",
			Value: p.PrivateID.String(),
		},
		{
			Name:  "PublicID",
			Value: p.PublicID.String(),
		},
		{
			Name:  "ParentID",
			Value: parentID,
		},
		{
			Name:  "TransactionCount",
			Value: p.TransactionCount,
		},
		{
			Name:  "Persist",
			Value: p.Persist,
		},
	}, nil
}

type Project struct {
	ID       uuid.UUID
	PublicID uuid.UUID
	ParentID *uuid.UUID
	Persist  bool
	Mutable  bool
}
