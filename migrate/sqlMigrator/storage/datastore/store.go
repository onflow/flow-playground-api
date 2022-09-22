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

package datastore

import (
	"context"
	"time"

	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/storage"

	"cloud.google.com/go/datastore"
	"github.com/Masterminds/semver"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/model"
)

type DatastoreEntity interface {
	NameKey() *datastore.Key
}

// Config is the configuration required to connect to Datastore.
type Config struct {
	DatastoreProjectID string
	DatastoreTimeout   time.Duration
}

const (
	defaultTimeout = time.Second * 5
)

var _ storage.Store = &Datastore{}

type Datastore struct {
	conf     *Config
	dsClient *datastore.Client
}

// NewDatastore initializes and returns a Datastore.
//
// This function will return an error if it fails to connect to Datastore.
func NewDatastore(
	ctx context.Context,
	conf *Config,
) (*Datastore, error) {
	if conf.DatastoreProjectID == "" {
		return nil, errors.New("missing env var: DATASTORE_PROJECT_ID")
	}
	if conf.DatastoreTimeout == 0 {
		conf.DatastoreTimeout = defaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, conf.DatastoreTimeout)
	defer cancel()
	dsClient, err := datastore.NewClient(ctx, conf.DatastoreProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Datastore")
	}

	return &Datastore{
		conf:     conf,
		dsClient: dsClient,
	}, nil
}

// Helper functions, wrapping all datastore functions with a timeout
// ===
func (d *Datastore) get(dst DatastoreEntity) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	return d.dsClient.Get(ctx, dst.NameKey(), dst)
}

func (d *Datastore) getAll(q *datastore.Query, dst interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, err := d.dsClient.GetAll(ctx, q, dst)
	return err
}

func (d *Datastore) put(src DatastoreEntity) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, err := d.dsClient.Put(ctx, src.NameKey(), src)
	return err
}

func (d *Datastore) markProjectUpdatedAt(tx *datastore.Transaction, projectID uuid.UUID) error {
	var proj model.InternalProject

	key := model.ProjectNameKey(projectID)

	err := tx.Get(model.ProjectNameKey(projectID), &proj)
	if err != nil {
		return err
	}

	proj.UpdatedAt = time.Now()

	_, err = tx.Put(key, &proj)
	if err != nil {
		return err
	}

	return nil
}

// Users

func (d *Datastore) InsertUser(user *model.User) error {
	return d.put(user)
}

func (d *Datastore) GetUser(id uuid.UUID, user *model.User) error {
	user.ID = id
	return d.get(user)
}

// Projects

func (d *Datastore) CreateProject(
	proj *model.InternalProject,
	ttpls []*model.TransactionTemplate,
	stpls []*model.ScriptTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	entitiesToPut := []interface{}{proj}
	keys := []*datastore.Key{proj.NameKey()}

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		for _, ttpl := range ttpls {
			ttpl.Index = proj.TransactionTemplateCount
			proj.TransactionTemplateCount++
			entitiesToPut = append(entitiesToPut, ttpl)
			keys = append(keys, ttpl.NameKey())
		}

		for _, stpl := range stpls {
			stpl.Index = proj.ScriptTemplateCount
			proj.ScriptTemplateCount++
			entitiesToPut = append(entitiesToPut, stpl)
			keys = append(keys, stpl.NameKey())

		}

		proj.CreatedAt = time.Now()
		proj.UpdatedAt = proj.CreatedAt

		_, err := tx.PutMulti(keys, entitiesToPut)

		return err
	})

	return txErr
}

func (d *Datastore) UpdateProject(input model.UpdateProject, proj *model.InternalProject) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	proj.ID = input.ID

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		err := tx.Get(proj.NameKey(), proj)
		if err != nil {
			return err
		}

		if input.Title != nil {
			proj.Title = *input.Title
		}

		if input.Description != nil {
			proj.Description = *input.Description
		}

		if input.Readme != nil {
			proj.Readme = *input.Readme
		}

		if input.Persist != nil {
			proj.Persist = *input.Persist
		}

		proj.UpdatedAt = time.Now()

		_, err = tx.Put(proj.NameKey(), proj)
		return err
	})

	return txErr
}

func (d *Datastore) UpdateProjectOwner(id, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var proj model.InternalProject

		err := tx.Get(model.ProjectNameKey(id), &proj)
		if err != nil {
			return err
		}

		proj.UserID = userID

		_, err = tx.Put(proj.NameKey(), &proj)
		return err
	})

	return txErr
}

func (d *Datastore) UpdateProjectVersion(id uuid.UUID, version *semver.Version) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var proj model.InternalProject

		err := tx.Get(model.ProjectNameKey(id), &proj)
		if err != nil {
			return err
		}

		proj.Version = version

		_, err = tx.Put(proj.NameKey(), &proj)
		return err
	})

	return txErr
}

func (d *Datastore) ResetProjectState(proj *model.InternalProject) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	err := d.get(proj)
	if err != nil {
		return err
	}

	var txExes []*model.TransactionExecution
	err = d.GetTransactionExecutionsForProject(proj.ID, &txExes)
	if err != nil {
		return err
	}

	var scriptExes []*model.ScriptExecution
	err = d.GetScriptExecutionsForProject(proj.ID, &scriptExes)
	if err != nil {
		return err
	}

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		proj.TransactionCount = 0
		proj.TransactionExecutionCount = 0
		proj.UpdatedAt = time.Now()

		_, err = tx.Put(proj.NameKey(), proj)
		if err != nil {
			return err
		}

		// delete all transaction executions
		for _, txExe := range txExes {
			err = tx.Delete(txExe.NameKey())
			if err != nil {
				return err
			}
		}

		// delete all scripts executions
		for _, scriptExe := range scriptExes {
			err = tx.Delete(scriptExe.NameKey())
			if err != nil {
				return err
			}
		}

		return nil
	})

	if txErr != nil {
		return txErr
	}

	return nil
}

func (d *Datastore) GetProject(id uuid.UUID, proj *model.InternalProject) error {
	proj.ID = id
	return d.get(proj)
}

// Accounts

func (d *Datastore) InsertAccount(acc *model.InternalAccount) error {
	return d.put(acc)
}

func (d *Datastore) GetAccount(id model.ProjectChildID, acc *model.InternalAccount) error {
	acc.ProjectChildID = id
	return d.get(acc)
}

func (d *Datastore) UpdateAccount(input model.UpdateAccount, acc *model.InternalAccount) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	acc.ID = input.ID
	acc.ProjectID = input.ProjectID

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		err := tx.Get(acc.NameKey(), acc)
		if err != nil {
			return err
		}

		if input.DraftCode != nil {
			acc.DraftCode = *input.DraftCode
		}

		err = d.markProjectUpdatedAt(tx, acc.ProjectID)
		if err != nil {
			return err
		}

		_, err = tx.Put(acc.NameKey(), acc)
		return err
	})

	return txErr
}

func (d *Datastore) GetAccountsForProject(projectID uuid.UUID, accs *[]*model.InternalAccount) error {
	q := datastore.NewQuery("Account").Ancestor(model.ProjectNameKey(projectID)).Order("Index")
	return d.getAll(q, accs)
}

// GetAllProjects populates list of projects
func (d *Datastore) GetAllProjects(projects *[]*model.InternalProject) error {
	// TODO: Query to get all Projects or Project UUIDs in a list
	// If I can only query for all the UUIDs of projects then do that
	// And format the projects into the list as expected here
	q := datastore.NewQuery("Project")
	return d.getAll(q, projects)
}

func (d *Datastore) DeleteAccount(id model.ProjectChildID) error {
	acc := model.InternalAccount{ProjectChildID: id}

	_, txErr := d.dsClient.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		err := tx.Delete(acc.NameKey())
		if err != nil {
			return err
		}

		err = d.markProjectUpdatedAt(tx, id.ProjectID)
		if err != nil {
			return err
		}

		return nil
	})

	return txErr
}

// Transaction Templates

func (d *Datastore) InsertTransactionTemplate(tpl *model.TransactionTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		proj := &model.InternalProject{
			ID: tpl.ProjectID,
		}

		err := tx.Get(proj.NameKey(), proj)
		if err != nil {
			return err
		}

		tpl.Index = proj.TransactionTemplateCount

		proj.TransactionTemplateCount++

		proj.UpdatedAt = time.Now()

		_, err = tx.PutMulti(
			[]*datastore.Key{proj.NameKey(), tpl.NameKey()},
			[]interface{}{proj, tpl},
		)
		return err
	})

	return txErr

}
func (d *Datastore) UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	tpl.ID = input.ID
	tpl.ProjectID = input.ProjectID

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		err := tx.Get(tpl.NameKey(), tpl)
		if err != nil {
			return err
		}

		if input.Index != nil {
			tpl.Index = *input.Index
		}

		if input.Script != nil {
			tpl.Script = *input.Script
		}

		if input.Title != nil {
			tpl.Title = *input.Title
		}

		err = d.markProjectUpdatedAt(tx, input.ProjectID)
		if err != nil {
			return err
		}

		_, err = tx.Put(tpl.NameKey(), tpl)
		return err
	})

	return txErr
}

func (d *Datastore) GetTransactionTemplate(id model.ProjectChildID, tpl *model.TransactionTemplate) error {
	tpl.ProjectChildID = id
	return d.get(tpl)
}

func (d *Datastore) GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error {
	q := datastore.NewQuery("TransactionTemplate").Ancestor(model.ProjectNameKey(projectID)).Order("Index")
	return d.getAll(q, tpls)
}

func (d *Datastore) DeleteTransactionTemplate(id model.ProjectChildID) error {
	ttpl := model.TransactionTemplate{ProjectChildID: id}

	_, txErr := d.dsClient.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		err := tx.Delete(ttpl.NameKey())
		if err != nil {
			return err
		}

		err = d.markProjectUpdatedAt(tx, id.ProjectID)
		if err != nil {
			return err
		}

		return nil
	})

	return txErr
}

// Transaction Executions

func (d *Datastore) InsertTransactionExecution(exe *model.TransactionExecution) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		proj := &model.InternalProject{
			ID: exe.ProjectID,
		}

		err := tx.Get(proj.NameKey(), proj)
		if err != nil {
			return err
		}

		exe.Index = proj.TransactionExecutionCount

		proj.TransactionExecutionCount++
		proj.TransactionCount++
		proj.UpdatedAt = time.Now()

		_, err = tx.PutMulti(
			[]*datastore.Key{proj.NameKey(), exe.NameKey()},
			[]interface{}{proj, exe},
		)
		return err
	})

	return txErr

}

func (d *Datastore) GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error {
	q := datastore.NewQuery("TransactionExecution").Ancestor(model.ProjectNameKey(projectID)).Order("Index")
	return d.getAll(q, exes)
}

// Script Templates

func (d *Datastore) InsertScriptTemplate(tpl *model.ScriptTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		proj := &model.InternalProject{
			ID: tpl.ProjectID,
		}
		err := tx.Get(proj.NameKey(), proj)
		if err != nil {
			return err
		}
		tpl.Index = proj.ScriptTemplateCount

		proj.ScriptTemplateCount++

		proj.UpdatedAt = time.Now()

		_, err = tx.PutMulti(
			[]*datastore.Key{proj.NameKey(), tpl.NameKey()},
			[]interface{}{proj, tpl},
		)

		return err
	})

	return txErr
}

func (d *Datastore) UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	tpl.ID = input.ID
	tpl.ProjectID = input.ProjectID
	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		err := tx.Get(tpl.NameKey(), tpl)
		if err != nil {
			return err
		}

		if input.Index != nil {
			tpl.Index = *input.Index
		}

		if input.Script != nil {
			tpl.Script = *input.Script
		}

		if input.Title != nil {
			tpl.Title = *input.Title
		}

		err = d.markProjectUpdatedAt(tx, input.ProjectID)
		if err != nil {
			return err
		}

		_, err = tx.Put(tpl.NameKey(), tpl)
		return err
	})

	return txErr
}

func (d *Datastore) GetScriptTemplate(id model.ProjectChildID, tpl *model.ScriptTemplate) error {
	tpl.ProjectChildID = id
	return d.get(tpl)
}

func (d *Datastore) GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error {
	q := datastore.NewQuery("ScriptTemplate").Ancestor(model.ProjectNameKey(projectID)).Order("Index")
	return d.getAll(q, tpls)
}

func (d *Datastore) DeleteScriptTemplate(id model.ProjectChildID) error {
	stpl := model.ScriptTemplate{ProjectChildID: id}

	_, txErr := d.dsClient.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		err := tx.Delete(stpl.NameKey())
		if err != nil {
			return err
		}

		err = d.markProjectUpdatedAt(tx, id.ProjectID)
		if err != nil {
			return err
		}

		return nil
	})

	return txErr
}

// Script Executions

func (d *Datastore) InsertScriptExecution(exe *model.ScriptExecution) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err := tx.Put(exe.NameKey(), exe)
		if err != nil {
			return err
		}

		err = d.markProjectUpdatedAt(tx, exe.ProjectID)
		if err != nil {
			return err
		}

		return nil
	})

	return txErr
}

func (d *Datastore) GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error {
	q := datastore.NewQuery("ScriptExecution").Ancestor(model.ProjectNameKey(projectID)).Order("Index")
	return d.getAll(q, exes)
}