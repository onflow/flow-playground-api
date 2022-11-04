package migrate

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"time"
)

type v1User struct {
	ID uuid.UUID
}

type v1Project struct {
	ID                        uuid.UUID
	UserID                    uuid.UUID
	Secret                    uuid.UUID
	PublicID                  uuid.UUID
	ParentID                  *uuid.UUID
	Title                     string
	Description               string
	Readme                    string
	Seed                      int
	TransactionExecutionCount int
	Persist                   bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	Version                   *semver.Version `gorm:"serializer:json"`
	Mutable                   bool
}

type v1Account struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Index             int
	Address           model.Address `gorm:"serializer:json"`
	DraftCode         string
	DeployedCode      string
	DeployedContracts []string `gorm:"serializer:json"`
	State             string
}

type v1TransactionTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Title     string
	Index     int
	Script    string
}

type v1ScriptTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Title     string
	Index     int
	Script    string
}
