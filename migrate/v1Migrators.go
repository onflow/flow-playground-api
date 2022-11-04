package migrate

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

// migrateV1AccountsToV2 converts old account draft codes to contract templates
func (m *Migrator) migrateV1AccountsToV2(projectID uuid.UUID) ([]*model.File, error) {
	// TODO: Add v1GetAccountsForProject in order to retrieve the draft codes
	var v1Accounts []*v1Account
	err := v1GetAccountsForProject(projectID, &v1Accounts)
	if err != nil {
		return nil, errors.Wrap(err, "migration failed to get accounts")
	}

	// Create contract files from old account draft codes
	var contractFiles []*model.File

	for i, account := range v1Accounts {
		contractFiles = append(contractFiles, &model.File{
			ID:        uuid.New(),
			ProjectID: projectID,
			Title:     "",
			Type:      model.ContractFile,
			Index:     i,
			Script:    account.DraftCode,
		})
	}

	// TODO: Delete v1Accounts from database

	return contractFiles, nil
}

// migrateV1ProjectToV2 migrates a project from v1 to v2
//
// Also migrates the project user if needed
func (m *Migrator) migrateV1ProjectToV2(projectID uuid.UUID) error {
	// TODO: implement GetV1Project for the old model
	v1Project, err := m.projects.GetV1Project(projectID)
	if err != nil {
		return errors.Wrap(err, "migration failed to get project")
	}

	v2ContractFiles, err := m.migrateV1AccountsToV2(projectID)
	if err != nil {
		return errors.Wrap(err, "migration failed to migrate v1 accounts to v2")
	}

	v2Project := model.Project{
		ID:                        projectID,
		UserID:                    v1Project.UserID,
		Secret:                    v1Project.Secret,
		PublicID:                  v1Project.PublicID,
		ParentID:                  v1Project.ParentID,
		Title:                     v1Project.Title,
		Description:               v1Project.Description,
		Readme:                    v1Project.Readme,
		Seed:                      v1Project.Seed,
		NumberOfAccounts:          5,
		TransactionExecutionCount: 0, // TODO: just reset these?
		Persist:                   v1Project.Persist,
		CreatedAt:                 v1Project.CreatedAt,
		UpdatedAt:                 time.Now(),
		Version:                   V2,
		Mutable:                   false,
	}

	// 4. TODO: Convert transaction templates and script templates to files and add to DB
	//    TODO: Delete old transaction templates and script templates from DB

	// TODO: Need to create a DeleteV1Project method?!? This will probably fail right?
	err = m.store.DeleteProject(projectID)
	if err != nil {
		return errors.Wrap(err, "migration failed to delete v1 project")
	}

	// Compile v2Project files from contract, transaction, and script templates
	v2ProjectFiles := make([]*model.File, 0)
	for _, contractTemplate := range v2ContractFiles {
		v2ProjectFiles = append(v2ProjectFiles, contractTemplate)
	}

	// TODO: Add transaction and script templates
	v2ProjectFiles = append(v2ProjectFiles, v2TransactionTemplates)
	v2ProjectFiles = append(v2ProjectFiles, v2ScriptTemplates)

	// Insert new project into DB
	err = m.store.CreateProject(&v2Project, v2ProjectFiles)

	return nil

}
