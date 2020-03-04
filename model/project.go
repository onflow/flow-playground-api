package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type InternalProject struct {
	ID                        uuid.UUID
	Secret                    uuid.UUID
	PublicID                  uuid.UUID
	ParentID                  *uuid.UUID
	TransactionCount          int
	TransactionExecutionCount int
	TransactionTemplateCount  int
	ScriptTemplateCount       int
	Persist                   bool
}

// ExportPublicMutable converts the internal project to its public representation
// and marks it as mutable.
func (p *InternalProject) ExportPublicMutable() *Project {
	return &Project{
		ID:       p.ID,
		PublicID: p.PublicID,
		ParentID: p.ParentID,
		Persist:  p.Persist,
		Seed:     p.Seed,
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
		Seed:     p.Seed,
		Mutable:  false,
	}
}

func (p *InternalProject) NameKey() *datastore.Key {
	return datastore.NameKey("Project", p.ID.String(), nil)
}

func (p *InternalProject) Load(ps []datastore.Property) error {
	tmp := struct {
		ID               string
		Secret           string
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
	if err := p.Secret.UnmarshalText([]byte(tmp.Secret)); err != nil {
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
			Name:  "Secret",
			Value: p.Secret.String(),
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
	Seed     int
	Persist  bool
	Mutable  bool
}
