// TODO Lots of places that could use transactions in this file
package datastore

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-go/engine/execution/execution/state"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

// Config is the configuration required to connect to Datastore.
type Config struct {
	DatastoreProjectID string
	DatastoreTimeout   time.Duration
}

const (
	defaultTimeout = time.Second * 5
)

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
) (storage.Store, error) {
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

func (d *Datastore) delete(src DatastoreEntity) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	return d.dsClient.Delete(ctx, src.NameKey())
}

// Projects

func (d *Datastore) InsertProject(proj *model.InternalProject) error {
	return d.put(proj)
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
		if input.Persist != nil {
			proj.Persist = *input.Persist
		}
		_, err = tx.Put(proj.NameKey(), proj)
		return err
	})

	return txErr
}

func (d *Datastore) GetProject(id uuid.UUID, proj *model.InternalProject) error {
	proj.ID = id
	return d.get(proj)
}

func (d *Datastore) InsertAccount(acc *model.Account) error {
	return d.put(acc)
}

// Accounts

func (d *Datastore) GetAccount(id model.ProjectChildID, acc *model.Account) error {
	acc.ProjectChildID = id
	return d.get(acc)
}
func (d *Datastore) UpdateAccount(input model.UpdateAccount, acc *model.Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	acc.ID = input.ID
	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		err := tx.Get(acc.NameKey(), acc)
		if err != nil {
			return err
		}
		if input.DraftCode != nil {
			acc.DraftCode = *input.DraftCode
		}

		if input.DeployedCode != nil {
			acc.DeployedCode = *input.DeployedCode
		}
		_, err = tx.Put(acc.NameKey(), acc)
		return err
	})

	return txErr
}
func (d *Datastore) GetAccountsForProject(projectID uuid.UUID, accs *[]*model.Account) error {
	q := datastore.NewQuery("Account").Filter("ProjectID=", projectID.String()).Order("Index")
	return d.getAll(q, accs)
}
func (d *Datastore) DeleteAccount(id model.ProjectChildID) error {
	return d.delete(&model.Account{ProjectChildID: id})
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
	q := datastore.NewQuery("TransactionTemplate").Filter("ProjectID=", projectID.String()).Order("Index")
	return d.getAll(q, tpls)
}
func (d *Datastore) DeleteTransactionTemplate(id model.ProjectChildID) error {
	return d.delete(&model.TransactionTemplate{ProjectChildID: id})
}

// Transaction Executions

func (d *Datastore) InsertTransactionExecution(exe *model.TransactionExecution, delta state.Delta) error {
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

		regDelta := &model.RegisterDelta{
			ProjectID: proj.ID,
			Index:     proj.TransactionCount,
			Delta:     delta,
		}

		proj.TransactionExecutionCount++
		proj.TransactionCount++

		_, err = tx.PutMulti(
			[]*datastore.Key{proj.NameKey(), exe.NameKey(), regDelta.NameKey()},
			[]interface{}{proj, exe, regDelta},
		)
		return err
	})

	return txErr

}
func (d *Datastore) GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error {
	q := datastore.NewQuery("TransactionExecution").Filter("ProjectID=", projectID.String()).Order("Index")
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
	q := datastore.NewQuery("ScriptTemplate").Filter("ProjectID=", projectID.String()).Order("Index")
	return d.getAll(q, tpls)
}
func (d *Datastore) DeleteScriptTemplate(id model.ProjectChildID) error {
	return d.delete(&model.ScriptTemplate{ProjectChildID: id})
}

// Script Executions

func (d *Datastore) InsertScriptExecution(exe *model.ScriptExecution) error {
	return d.put(exe)
}
func (d *Datastore) GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error {
	q := datastore.NewQuery("ScriptExecution").Filter("ProjectID=", projectID.String()).Order("Index")
	return d.getAll(q, exes)
}

// Register Deltas

func (d *Datastore) InsertRegisterDelta(projectID uuid.UUID, delta state.Delta) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.conf.DatastoreTimeout)
	defer cancel()

	proj := &model.InternalProject{
		ID: projectID,
	}

	_, txErr := d.dsClient.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		err := tx.Get(proj.NameKey(), proj)
		if err != nil {
			return err
		}

		regDelta := &model.RegisterDelta{
			ProjectID: projectID,
			Index:     proj.TransactionCount,
			Delta:     delta,
		}
		proj.TransactionCount++

		_, err = tx.PutMulti(
			[]*datastore.Key{proj.NameKey(), regDelta.NameKey()},
			[]interface{}{proj, regDelta},
		)
		return err
	})

	return txErr
}
func (d *Datastore) GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]state.Delta) error {
	reg := []model.RegisterDelta{}
	q := datastore.NewQuery("RegisterDelta").Filter("ProjectID=", projectID.String()).Order("Index")
	err := d.getAll(q, &reg)
	if err != nil {
		return err
	}
	for _, d := range reg {
		*deltas = append(*deltas, d.Delta)
	}
	return nil
}
