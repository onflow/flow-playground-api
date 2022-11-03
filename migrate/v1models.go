package migrate

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
)

type v1Account struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Index             int
	Address           model.Address `gorm:"serializer:json"`
	DraftCode         string
	DeployedCode      string   // todo drop this in db
	DeployedContracts []string `gorm:"serializer:json"`
	State             string
}

type v1User struct {
	ID uuid.UUID
}
