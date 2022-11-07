package migrate

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

// TODO: Will we even be able to query the SQL DB for old projects after we connect and it hits migrateDB
// TODO: which sets the tables to v2 tables??

// migrateV1ProjectToV2 migrates a project from v1 to v2
func (m *Migrator) migrateV1ProjectToV2(db *gorm.DB, projectID uuid.UUID) error {
	var v1Proj v1Project
	err := GetV1Project(db, projectID, &v1Proj)
	if err != nil {
		return errors.Wrap(err, "migration failed to get project")
	}

	// Migrate v1 account draft code to v2 contract files
	v2ContractFiles, err := m.migrateV1AccountsToV2(db, projectID)
	if err != nil {
		return errors.Wrap(err, "migration failed to migrate v1 accounts to v2")
	}

	v2Project := model.Project{
		ID:                        projectID,
		UserID:                    v1Proj.UserID,
		Secret:                    v1Proj.Secret,
		PublicID:                  v1Proj.PublicID,
		ParentID:                  v1Proj.ParentID,
		Title:                     v1Proj.Title,
		Description:               v1Proj.Description,
		Readme:                    v1Proj.Readme,
		Seed:                      v1Proj.Seed,
		NumberOfAccounts:          5,
		TransactionExecutionCount: 0,
		Persist:                   true,
		CreatedAt:                 v1Proj.CreatedAt,
		UpdatedAt:                 time.Now(),
		Version:                   V2,
		Mutable:                   false,
	}

	// TODO: Convert transaction templates and script templates to files and add to DB
	// TODO: Delete old transaction templates and script templates from DB

	// TODO: Clean up DB - delete v1Project, delete v1Accounts, delete v1 transaction + script templates

	/*
		err = m.store.DeleteV1Project(projectID)
		if err != nil {
			return errors.Wrap(err, "migration failed to delete v1 project")
		}
	*/

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

// migrateV1AccountsToV2 converts old account draft codes to contract templates
func (m *Migrator) migrateV1AccountsToV2(db *gorm.DB, projectID uuid.UUID) ([]*model.File, error) {
	var v1Accounts []*v1Account
	err := v1GetAccountsForProject(db, projectID, &v1Accounts)
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
