package model

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type RegisterDelta struct {
	ProjectID uuid.UUID
	Index     int
	Delta     delta.Delta
}

func (r *RegisterDelta) NameKey() *datastore.Key {
	return datastore.NameKey("RegisterDelta", fmt.Sprintf("%s-%d", r.ProjectID.String(), r.Index), ProjectNameKey(r.ProjectID))
}

func (r *RegisterDelta) Load(ps []datastore.Property) error {
	tmp := struct {
		ProjectID         string
		Index             int
		Delta             []byte
		IsAccountCreation bool // IsAccountCreation field kept for backwards compatibility
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := r.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	r.Index = tmp.Index

	var delta delta.Delta

	decoder := gob.NewDecoder(bytes.NewReader(tmp.Delta))
	err := decoder.Decode(&delta)
	if err != nil {
		return errors.Wrap(err, "failed to decode Delta")
	}

	r.Delta = delta

	return nil
}

func (r *RegisterDelta) Save() ([]datastore.Property, error) {
	w := new(bytes.Buffer)

	encoder := gob.NewEncoder(w)
	err := encoder.Encode(&r.Delta)
	if err != nil {
		return nil, err
	}

	delta := w.Bytes()

	return []datastore.Property{
		{
			Name:  "ProjectID",
			Value: r.ProjectID.String(),
		},
		{
			Name:  "Index",
			Value: r.Index,
		},
		{
			Name:    "Delta",
			Value:   delta,
			NoIndex: true,
		},
	}, nil
}
