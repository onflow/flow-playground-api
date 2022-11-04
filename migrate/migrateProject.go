package migrate

import (
	"github.com/google/uuid"
)

// migrateV1ProjectToV2 migrates a project from v1 to v2
//
// Also migrates the project user if needed
func (m *Migrator) migrateV1ProjectToV2(projectID uuid.UUID) error {
	_ = projectID
	return nil
	/*
		// TODO: implement GetV1Project for the old model
		v1Project, err := m.projects.GetV1Project(projectID)
		if err != nil {
			return errors.Wrap(err, "migration failed to get project")
		}

		// TODO: Migrate the project's user first (since it counts the number of v1 projects for user)
		_, err := m.MigrateV1UserToV2(v1Project.UserID)

		// 1. reset project state
		// TODO: Need to use the old project model? And then create a new project model to store in db!
		createdAccounts, err := m.projects.Reset(&model.Project{ID: projectID})
		if err != nil {
			return errors.Wrap(err, "migration failed to reset project state")
		}

		// 2. TODO: Add back GetAccountsForProject in order to retrieve the contracts + number of accounts
		//    TODO: Need the v1.0.0 account model to do migration
		var oldAccounts []*v1Account
		err = v1GetAccountsForProject(projectID, &oldAccounts)
		if err != nil {
			return errors.Wrap(err, "migration failed to get accounts")
		}

		// 3. Create contract files from old account draft codes
		var contractFiles []*model.File

		for i, account := range oldAccounts {
			contractFiles = append(contractFiles, &model.File{
				ID:        uuid.New(),
				ProjectID: projectID,
				Title:     "", // TODO: do we need to get the title here? Probably not?
				Type:      model.ContractFile,
				Index:     i,
				Script:    account.DraftCode,
			})
		}

		v2Project = model.Project{
			ID:                        projectID,
			UserID:                    v1Project.UserID,
			Secret:                    v1Project.Secret,
			PublicID:                  v1Project.PublicID,
			ParentID:                  v1Project.ParentID,
			Title:                     v1Project.Title,
			Description:               v1Project.Description,
			Readme:                    v1Project.Readme,
			Seed:                      v1Project.Seed,
			NumberOfAccounts:          len(oldAccounts),
			TransactionExecutionCount: 0, // TODO: just reset these?
			Persist:                   v1Project.Persist,
			CreatedAt:                 v1Project.CreatedAt,
			UpdatedAt:                 time.Now(),
			Version:                   V2,
			Mutable:                   false,
		}

		// 4. TODO: Convert transaction templates and script templates to files and add to DB
		//    TODO: Delete old transaction templates and script templates from DB

		// TODO: Delete old project model from DB and insert the new one

		return nil

	*/
}
