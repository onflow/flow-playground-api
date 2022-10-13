package model

type File struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Index             int
	Address           Address `gorm:"serializer:json"`
	DraftCode         string
	DeployedCode      string   // todo drop this in db
	DeployedContracts []string `gorm:"serializer:json"`
	State             string
}
