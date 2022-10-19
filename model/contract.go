package model

import (
	"github.com/pkg/errors"
)

// ContractTemplate TODO: Why is this being re-declared in models_gen but script and trans are not?

type ContractTemplate = File

type ContractDeployment struct {
	File
	Address Address        `gorm:"serializer:json"`
	Errors  []ProgramError `gorm:"serializer:json"`
	Events  []Event        `gorm:"serializer:json"`
	Logs    []string       `gorm:"serializer:json"`
}

func (u *UpdateContractTemplate) Validate() error {
	if u.Title == nil && u.Index == nil && u.Script == nil {
		return errors.Wrap(missingValuesError, "title, index, script")
	}
	return nil
}
