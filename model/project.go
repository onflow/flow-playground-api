package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
)

type InternalProject struct {
	ID               uuid.UUID
	PrivateID        uuid.UUID
	PublicID         uuid.UUID
	ParentID         *uuid.UUID
	TransactionCount int
	Persist          bool
}

func (p *InternalProject) ExportPrivate() *Project {
	return &Project{
		ID:        p.ID,
		PrivateID: &p.PrivateID,
		PublicID:  p.PublicID,
		ParentID:  p.ParentID,
		Persist:   p.Persist,
		Mutable:   true,
	}
}

func (p *InternalProject) ExportPublicMutable() *Project {
	return &Project{
		ID:       p.ID,
		PublicID: p.PublicID,
		ParentID: p.ParentID,
		Persist:  p.Persist,
		Mutable:  true,
	}
}

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

type Project struct {
	ID        uuid.UUID
	PrivateID *uuid.UUID
	PublicID  uuid.UUID
	ParentID  *uuid.UUID
	Persist   bool
	Mutable   bool
}
