package model

import (
	"github.com/google/uuid"
)

type InternalProject struct {
	ID               uuid.UUID
	PrivateID        uuid.UUID
	PublicID         uuid.UUID
	TransactionCount int
	Persist          bool
}

func (p *InternalProject) ExportPrivate() *Project {
	return &Project{
		ID:        p.ID,
		PrivateID: &p.PrivateID,
		PublicID:  p.PublicID,
		Persist:   p.Persist,
		Mutable:   true,
	}
}

func (p *InternalProject) ExportPublicMutable() *Project {
	return &Project{
		ID:       p.ID,
		PublicID: p.PublicID,
		Persist:  p.Persist,
		Mutable:  true,
	}
}

func (p *InternalProject) ExportPublicImmutable() *Project {
	return &Project{
		ID:       p.ID,
		PublicID: p.PublicID,
		Persist:  p.Persist,
		Mutable:  false,
	}
}

type Project struct {
	ID        uuid.UUID
	PrivateID *uuid.UUID
	PublicID  uuid.UUID
	Persist   bool
	Mutable   bool
}
