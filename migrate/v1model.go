package migrate

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
)

type v1_0_0Account struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Index             int
	Address           model.Address `gorm:"serializer:json"`
	DraftCode         string
	DeployedCode      string   // todo drop this in db
	DeployedContracts []string `gorm:"serializer:json"`
	State             string
}
