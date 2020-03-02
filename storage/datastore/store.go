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
	proj.ID = input.ID
	err := d.get(proj)
	if err != nil {
		return err
	}
	if input.Persist != nil {
		proj.Persist = *input.Persist
	}
	return d.put(proj)
}

func (d *Datastore) GetProject(id uuid.UUID, proj *model.InternalProject) error {
	proj.ID = id
	return d.get(proj)
}

func (d *Datastore) InsertAccount(acc *model.Account) error {
	return d.put(acc)
}

// Accounts

func (d *Datastore) GetAccount(id uuid.UUID, acc *model.Account) error {
	acc.ID = id
	return d.get(acc)
}
func (d *Datastore) UpdateAccount(input model.UpdateAccount, acc *model.Account) error {
	acc.ID = input.ID
	err := d.get(acc)
	if err != nil {
		return err
	}
	if input.DraftCode != nil {
		acc.DraftCode = *input.DraftCode
	}

	if input.DeployedCode != nil {
		acc.DeployedCode = *input.DeployedCode
	}
	return d.put(acc)
}
func (d *Datastore) GetAccountsForProject(projectID uuid.UUID, accs *[]*model.Account) error {
	q := datastore.NewQuery("Account").Filter("ProjectID=", projectID).Order("Index")
	return d.getAll(q, accs)
}
func (d *Datastore) DeleteAccount(id uuid.UUID) error {
	return d.delete(&model.Account{ID: id})
}

// Transaction Templates

func (d *Datastore) InsertTransactionTemplate(tpl *model.TransactionTemplate) error {
	return d.put(tpl)
}
func (d *Datastore) UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error {
	tpl.ID = input.ID
	err := d.get(tpl)
	if err != nil {
		return err
	}
	if input.Index != nil {
		tpl.Index = *input.Index
	}

	if input.Script != nil {
		tpl.Script = *input.Script
	}

	return d.put(tpl)
}
func (d *Datastore) GetTransactionTemplate(id uuid.UUID, tpl *model.TransactionTemplate) error {
	tpl.ID = id
	return d.get(tpl)
}
func (d *Datastore) GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error {
	q := datastore.NewQuery("TransactionTemplate").Filter("ProjectID=", projectID).Order("Index")
	return d.getAll(q, tpls)
}
func (d *Datastore) DeleteTransactionTemplate(id uuid.UUID) error {
	return d.delete(&model.TransactionTemplate{ID: id})
}

// Transaction Executions

func (d *Datastore) InsertTransactionExecution(exe *model.TransactionExecution, delta state.Delta) error {
	exes := []*model.TransactionExecution{}
	err := d.GetTransactionExecutionsForProject(exe.ProjectID, &exes)
	if err != nil {
		return err
	}
	index := len(exes)
	exe.Index = index

	err = d.put(exe)
	if err != nil {
		return err
	}

	return d.InsertRegisterDelta(exe.ProjectID, delta)
}
func (d *Datastore) GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error {
	q := datastore.NewQuery("TransactionExecution").Filter("ProjectID=", projectID).Order("Index")
	return d.getAll(q, exes)
}

// Script Templates

func (d *Datastore) InsertScriptTemplate(tpl *model.ScriptTemplate) error {
	return d.put(tpl)
}
func (d *Datastore) UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error {
	tpl.ID = input.ID
	err := d.get(tpl)
	if err != nil {
		return err
	}

	if input.Index != nil {
		tpl.Index = *input.Index
	}

	if input.Script != nil {
		tpl.Script = *input.Script
	}
	return d.put(tpl)
}
func (d *Datastore) GetScriptTemplate(id uuid.UUID, tpl *model.ScriptTemplate) error {
	tpl.ID = id
	return d.get(tpl)
}
func (d *Datastore) GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error {
	q := datastore.NewQuery("ScriptTemplate").Filter("ProjectID=", projectID).Order("Index")
	return d.getAll(q, tpls)
}
func (d *Datastore) DeleteScriptTemplate(id uuid.UUID) error {
	return d.delete(&model.ScriptTemplate{ID: id})
}

// Script Executions

func (d *Datastore) InsertScriptExecution(exe *model.ScriptExecution) error {
	return d.put(exe)
}
func (d *Datastore) GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error {
	q := datastore.NewQuery("ScriptExecution").Filter("ProjectID=", projectID).Order("Index")
	return d.getAll(q, exes)
}

// Register Deltas

func (d *Datastore) InsertRegisterDelta(projectID uuid.UUID, delta state.Delta) error {
	proj := &model.InternalProject{
		ID: projectID,
	}
	err := d.get(proj)
	if err != nil {
		return err
	}
	index := proj.TransactionCount + 1

	regDelta := model.RegisterDelta{
		ProjectID: projectID,
		Index:     index,
		Delta:     delta,
	}
	err = d.put(&regDelta)
	if err != nil {
		return err
	}

	proj.TransactionCount = index
	return d.put(proj)
}
func (d *Datastore) GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]state.Delta) error {
	reg := []model.RegisterDelta{}
	q := datastore.NewQuery("RegisterDelta").Filter("ProjectID=", projectID).Order("Index")
	err := d.getAll(q, &reg)
	if err != nil {
		return err
	}
	for _, d := range reg {
		*deltas = append(*deltas, d.Delta)
	}
	return nil
}
