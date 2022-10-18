/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _ Store = &SQL{}

const PostgreSQL = "postgresql"

// NewInMemory database, warning not concurrency safe, do not use for e2e tests
func NewInMemory() *SQL {
	return newSQL(sqlite.Open(":memory:"), logger.Warn)
}

func NewSqlite() *SQL {
	return newSQL(sqlite.Open("./e2e-db"), logger.Warn)
}

type DatabaseConfig struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     int
}

func NewPostgreSQL(conf *DatabaseConfig) *SQL {
	config := postgres.Config{
		DSN: fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			conf.Host,
			conf.User,
			conf.Password,
			conf.Name,
			conf.Port,
		),
	}

	return newSQL(postgres.New(config), logger.Error)
}

func newSQL(dial gorm.Dialector, level logger.LogLevel) *SQL {
	gormConf := &gorm.Config{
		Logger: logger.Default.LogMode(level),
	}

	db, err := gorm.Open(dial, gormConf)
	if err != nil {
		err := errors.Wrap(err, "failed to connect database")
		sentry.CaptureException(err)
		panic(err)
	}

	migrate(db)

	d, err := db.DB()
	if err != nil {
		panic(err)
	}
	d.SetMaxIdleConns(5) // we increase idle connection count due to nature of Playground API usage

	return &SQL{
		db: db,
	}
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.Project{},
		//&model.Account{},
		&model.File{},
		&model.ContractDeployment{},
		&model.ScriptExecution{},
		&model.TransactionExecution{},
		&model.User{},
	)
	if err != nil {
		err := errors.Wrap(err, "failed to migrate database")
		sentry.CaptureException(err)
		panic(err)
	}
}

type SQL struct {
	db *gorm.DB
}

func (s *SQL) GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error {
	//TODO implement me
	panic("implement me")
}

func (s *SQL) InsertScriptExecution(exe *model.ScriptExecution) error {
	//TODO implement me
	panic("implement me")
}

func (s *SQL) GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error {
	//TODO implement me
	panic("implement me")
}

func (s *SQL) InsertUser(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *SQL) GetUser(id uuid.UUID, user *model.User) error {
	return s.db.First(user, id).Error
}

func (s *SQL) CreateProject(proj *model.Project, files []*model.File) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(proj).Error; err != nil {
			return err
		}

		if len(files) > 0 {
			if err := tx.Create(files).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *SQL) UpdateProject(input model.UpdateProject, proj *model.Project) error {
	update := make(map[string]any)
	if input.Title != nil {
		update["title"] = *input.Title
	}
	if input.Description != nil {
		update["description"] = *input.Description
	}
	if input.Readme != nil {
		update["readme"] = *input.Readme
	}
	if input.Persist != nil {
		update["persist"] = *input.Persist
	}

	err := s.db.
		Model(&model.Project{ID: input.ID}).
		Updates(update).Error
	if err != nil {
		return err
	}

	return s.db.First(proj, input.ID).Error
}

func (s *SQL) UpdateProjectOwner(id, userID uuid.UUID) error {
	return s.db.
		Model(&model.Project{ID: id}).
		Updates(&model.Project{
			ID:     id,
			UserID: userID,
		}).Error
}

func (s *SQL) UpdateProjectVersion(id uuid.UUID, version *semver.Version) error {
	return s.db.
		Model(&model.Project{ID: id}).
		Updates(&model.Project{
			ID:      id,
			Version: version,
		}).Error
}

func (s *SQL) ResetProjectState(proj *model.Project) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(
			&model.TransactionExecution{},
			&model.TransactionExecution{
				File: model.File{ProjectID: proj.ID},
			},
		).Error
		if err != nil {
			return err
		}

		err = tx.Delete(
			&model.ScriptExecution{},
			&model.ScriptExecution{
				File: model.File{ProjectID: proj.ID},
			},
		).Error
		if err != nil {
			return err
		}

		err = tx.
			Model(&model.Project{ID: proj.ID}).
			Updates(map[string]any{ // need to use map due to zero value, see https://gorm.io/docs/update.html
				"TransactionExecutionCount": 0,
			}).Error

		return err
	})
}

func (s *SQL) GetProject(id uuid.UUID, proj *model.Project) error {
	return s.db.First(proj, id).Error
}

/*
func (s *SQL) InsertAccount(acc *model.Account) error {
	return s.db.Save(acc).Error
}

func (s *SQL) InsertAccounts(accs []*model.Account) error {
	return s.db.Create(accs).Error
}

func (s *SQL) GetAccount(id, pID uuid.UUID, acc *model.Account) error {
	return s.db.First(acc, &model.Account{ID: id, ProjectID: pID}).Error
}

func (s *SQL) GetAccountsForProject(pID uuid.UUID, accs *[]*model.Account) error {
	return s.db.
		Where(&model.Account{ProjectID: pID}).
		Order("\"index\" asc").
		Find(accs).
		Error
}

func (s *SQL) DeleteAccount(id, pID uuid.UUID) error {
	return s.db.Delete(&model.Account{ID: id, ProjectID: pID}).Error
}
*/

/*
func (s *SQL) UpdateAccount(input model.UpdateAccount, acc *model.Account) error {
	// TODO: Remove accounts completely???
	update := make(map[string]any)
	if input.DraftCode != nil {
		update["draft_code"] = *input.DraftCode
	}

	err := s.db.Model(&model.Account{
		ID:        input.ID,
		ProjectID: input.ProjectID,
	}).Updates(update).Error
	if err != nil {
		return err
	}

	return s.db.First(acc, input.ID).Error
}
*/

func (s *SQL) InsertCadenceFile(file *model.File) error {
	var count int64
	err := s.db.Model(&model.File{}).
		Where("project_id", file.ProjectID).
		Count(&count).Error
	if err != nil {
		return err
	}

	file.Index = int(count)
	return s.db.Create(file).Error
}

func (s *SQL) UpdateCadenceFile(input model.UpdateFile, file *model.File) error {
	update := make(map[string]any)
	if input.Script != nil {
		update["script"] = *input.Script
	}
	if input.Title != nil {
		update["title"] = *input.Title
	}
	if input.Index != nil {
		update["index"] = *input.Index
	}

	err := s.db.
		Model(&model.File{
			ID:        input.ID,
			ProjectID: input.ProjectID,
		}).
		Updates(update).Error
	if err != nil {
		return err
	}

	return s.db.First(file, input.ID).Error
}

func (s *SQL) DeleteCadenceFile(id, pID uuid.UUID) error {
	return s.db.Delete(&model.File{ID: id, ProjectID: pID}).Error
}

func (s *SQL) GetFile(id, pID uuid.UUID, file *model.File) error {
	return s.db.First(file, &model.File{ID: id, ProjectID: pID}).Error
}

func (s *SQL) GetFilesForProject(pID uuid.UUID, files *[]*model.File, fileType model.FileType) error {
	return s.db.Where(&model.File{ProjectID: pID, Type: fileType}).Find(files).Error
}

func (s *SQL) GetAllFilesForProject(pID uuid.UUID, files *[]*model.File) error {
	return s.db.Where(&model.File{ProjectID: pID}).Find(files).Error
}

func (s *SQL) InsertTransactionExecution(exe *model.TransactionExecution) error {
	var proj model.Project
	if err := s.db.First(&proj, &model.Project{ID: exe.ProjectID}).Error; err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		exe.Index = proj.TransactionExecutionCount
		proj.TransactionExecutionCount += 1
		if err := tx.Save(proj).Error; err != nil {
			return err
		}

		if err := tx.Create(exe).Error; err != nil {
			return err
		}

		return nil
	})
}

/*
func (s *SQL) GetTransactionExecutionsForProject(pID uuid.UUID, exes *[]*model.TransactionExecution) error {
	return s.db.Where(&model.TransactionExecution{ProjectID: pID}).
		Order("\"index\" asc").
		Find(exes).
		Error
}
*/
