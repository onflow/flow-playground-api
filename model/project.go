package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type InternalProject struct {
	ID                        uuid.UUID
	UserID                    uuid.UUID
	Secret                    uuid.UUID
	PublicID                  uuid.UUID
	ParentID                  *uuid.UUID
	Title                     string
	Seed                      int
	TransactionCount          int
	TransactionExecutionCount int
	TransactionTemplateCount  int
	ScriptTemplateCount       int
	Persist                   bool
}

func (p *InternalProject) IsOwnedBy(userID uuid.UUID) bool {
	return p.UserID == userID
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

func ProjectNameKey(id uuid.UUID) *datastore.Key {
	return datastore.NameKey("Project", id.String(), nil)
}

func (p *InternalProject) NameKey() *datastore.Key {
	return ProjectNameKey(p.ID)
}

func (p *InternalProject) Load(ps []datastore.Property) error {
	tmp := struct {
		ID                        string
		UserID                    string
		Secret                    string
		PublicID                  string
		ParentID                  *string
		Title                     string
		Seed                      int
		TransactionCount          int
		TransactionExecutionCount int
		TransactionTemplateCount  int
		ScriptTemplateCount       int
		Persist                   bool
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := p.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	if tmp.UserID != "" {
		if err := p.UserID.UnmarshalText([]byte(tmp.UserID)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	}

	if tmp.Secret != "" {
		if err := p.Secret.UnmarshalText([]byte(tmp.Secret)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	}

	if err := p.PublicID.UnmarshalText([]byte(tmp.PublicID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	if tmp.ParentID != nil && len(*tmp.ParentID) != 0 {
		p.ParentID = new(uuid.UUID)
		if err := p.ParentID.UnmarshalText([]byte(*tmp.ParentID)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	} else {
		p.ParentID = nil
	}

	p.Title = tmp.Title
	p.Seed = tmp.Seed
	p.TransactionCount = tmp.TransactionCount
	p.TransactionExecutionCount = tmp.TransactionExecutionCount
	p.TransactionTemplateCount = tmp.TransactionTemplateCount
	p.ScriptTemplateCount = tmp.ScriptTemplateCount
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
			Name:  "UserID",
			Value: p.UserID.String(),
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
			Name:  "Title",
			Value: p.Title,
		},
		{
			Name:  "Seed",
			Value: p.Seed,
		},
		{
			Name:  "TransactionCount",
			Value: p.TransactionCount,
		},
		{
			Name:  "TransactionExecutionCount",
			Value: p.TransactionExecutionCount,
		},
		{
			Name:  "TransactionTemplateCount",
			Value: p.TransactionTemplateCount,
		},
		{
			Name:  "ScriptTemplateCount",
			Value: p.ScriptTemplateCount,
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
	Title    string
	Persist  bool
	Mutable  bool
}

type ProjectChildID struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
}

func NewProjectChildID(id uuid.UUID, projectID uuid.UUID) ProjectChildID {
	return ProjectChildID{ID: id, ProjectID: projectID}
}
