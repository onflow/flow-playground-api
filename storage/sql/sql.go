package sql

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

var _ storage.Store = &SQL{}

func newSQL() *SQL {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// todo db.AutoMigrate()

	return &SQL{
		db: db,
	}
}

type SQL struct {
	db *gorm.DB
}

func (s *SQL) InsertUser(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *SQL) GetUser(id uuid.UUID, user *model.User) error {
	return s.db.First(user, id).Error
}

func (s *SQL) CreateProject(proj *model.Project, ttpl []*model.TransactionTemplate, stpl []*model.ScriptTemplate) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(proj).Error; err != nil {
			return err
		}
		if err := tx.Create(ttpl).Error; err != nil {
			return err
		}
		if err := tx.Create(stpl).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *SQL) UpdateProject(input model.UpdateProject, proj *model.Project) error {
	err := s.db.
		Model(proj).
		Updates(model.Project{
			Title:       *input.Title,
			Description: *input.Description,
			Readme:      *input.Readme,
			Persist:     *input.Persist,
		}).Error
	if err != nil {
		return err
	}

	return s.db.First(proj, input.ID).Error
}

func (s *SQL) UpdateProjectOwner(id, userID uuid.UUID) error {
	return s.db.
		Model(&model.Project{}).
		Updates(&model.Project{
			ID:     id,
			UserID: userID,
		}).Error
}

func (s *SQL) UpdateProjectVersion(id uuid.UUID, version *semver.Version) error {
	return s.db.
		Model(&model.Project{}).
		Updates(&model.Project{
			ID:      id,
			Version: version,
		}).Error
}

func (s *SQL) ResetProjectState(proj *model.Project) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(
			&model.TransactionExecution{},
			&model.TransactionExecution{ProjectID: proj.ID},
		).Error
		if err != nil {
			return err
		}

		err = tx.Delete(
			&model.ScriptExecution{},
			&model.ScriptExecution{ProjectID: proj.ID},
		).Error
		if err != nil {
			return err
		}

		err = tx.Model(proj).Updates(&model.Project{
			TransactionCount:          0,
			TransactionExecutionCount: 0,
			UpdatedAt:                 time.Now(),
		}).Error

		return err
	})
}

func (s *SQL) GetProject(id uuid.UUID, proj *model.Project) error {
	return s.db.First(proj, id).Error
}

func (s *SQL) InsertAccount(acc *model.Account) error {
	return s.db.Save(acc).Error
}

func (s *SQL) GetAccount(id, pID uuid.UUID, acc *model.Account) error {
	return s.db.First(acc, &model.Account{ID: id, ProjectID: pID}).Error
}

func (s *SQL) GetAccountsForProject(pID uuid.UUID, accs *[]*model.Account) error {
	return s.db.Where(&model.Account{ProjectID: pID}).Find(accs).Error
}

func (s *SQL) DeleteAccount(id, pID uuid.UUID) error {
	return s.db.Delete(&model.Account{ID: id, ProjectID: pID}).Error
}

func (s *SQL) UpdateAccount(input model.UpdateAccount, acc *model.Account) error {
	err := s.db.
		Model(acc).
		Updates(&model.Account{
			ID:        input.ID,
			ProjectID: input.ProjectID,
			DraftCode: *input.DraftCode,
		}).Error
	if err != nil {
		return err
	}

	return s.db.First(acc, input.ID).Error
}

func (s *SQL) InsertTransactionTemplate(tpl *model.TransactionTemplate) error {
	return s.db.Save(tpl).Error
}

func (s *SQL) UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error {
	err := s.db.
		Model(tpl).
		Updates(&model.TransactionTemplate{
			ID:        input.ID,
			ProjectID: input.ProjectID,
			Title:     *input.Title,
			Index:     *input.Index,
			Script:    *input.Script,
		}).Error
	if err != nil {
		return err
	}

	return s.db.First(tpl, input.ID).Error
}

func (s *SQL) GetTransactionTemplate(id, pID uuid.UUID, tpl *model.TransactionTemplate) error {
	return s.db.First(tpl, &model.TransactionTemplate{ID: id, ProjectID: pID}).Error
}

func (s *SQL) GetTransactionTemplatesForProject(pID uuid.UUID, tpls *[]*model.TransactionTemplate) error {
	return s.db.Where(&model.TransactionTemplate{ProjectID: pID}).Find(tpls).Error
}

func (s *SQL) DeleteTransactionTemplate(id, pID uuid.UUID) error {
	return s.db.Delete(&model.TransactionTemplate{ID: id, ProjectID: pID}).Error
}

func (s *SQL) InsertTransactionExecution(exe *model.TransactionExecution) error {
	return s.db.Save(exe).Error
}

func (s *SQL) GetTransactionExecutionsForProject(pID uuid.UUID, exes *[]*model.TransactionExecution) error {
	return s.db.Where(&model.TransactionExecution{ProjectID: pID}).Find(exes).Error
}

func (s *SQL) InsertScriptTemplate(tpl *model.ScriptTemplate) error {
	return s.db.Save(tpl).Error
}

func (s *SQL) UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error {
	err := s.db.
		Model(tpl).
		Updates(&model.ScriptTemplate{
			ID:        input.ID,
			ProjectID: input.ProjectID,
			Title:     *input.Title,
			Index:     *input.Index,
			Script:    *input.Script,
		}).Error
	if err != nil {
		return err
	}

	return s.db.First(tpl, input.ID).Error
}

func (s *SQL) GetScriptTemplate(id, pID uuid.UUID, tpl *model.ScriptTemplate) error {
	return s.db.First(tpl, &model.ScriptTemplate{ID: id, ProjectID: pID}).Error
}

func (s *SQL) GetScriptTemplatesForProject(pID uuid.UUID, tpls *[]*model.ScriptTemplate) error {
	return s.db.Where(&model.ScriptTemplate{ProjectID: pID}).Find(tpls).Error
}

func (s *SQL) DeleteScriptTemplate(id, pID uuid.UUID) error {
	return s.db.Delete(&model.ScriptTemplate{ID: id, ProjectID: pID}).Error
}

func (s *SQL) InsertScriptExecution(exe *model.ScriptExecution) error {
	return s.db.Save(exe).Error
}

func (s *SQL) GetScriptExecutionsForProject(pID uuid.UUID, exes *[]*model.ScriptExecution) error {
	return s.db.Where(&model.ScriptExecution{ProjectID: pID}).Find(exes).Error
}
