package controller

import (
	"github.com/google/uuid"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/compute"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Scripts struct {
	store    storage.Store
	computer *compute.Computer
}

func NewScripts(
	store storage.Store,
	computer *compute.Computer,
) *Scripts {
	return &Scripts{
		store:    store,
		computer: computer,
	}
}

func (s *Scripts) CreateExecution(proj *model.InternalProject, script string) (*model.ScriptExecution, error) {
	if len(script) == 0 {
		return nil, errors.New("cannot execute empty script")
	}

	result, err := s.computer.ExecuteScript(
		proj.ID,
		proj.TransactionCount,
		func() ([]*model.RegisterDelta, error) {
			var deltas []*model.RegisterDelta
			err := s.store.GetRegisterDeltasForProject(proj.ID, &deltas)
			if err != nil {
				return nil, err
			}

			return deltas, nil
		},
		script,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute script")
	}

	exe := model.ScriptExecution{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: proj.ID,
		},
		Script: script,
		Logs:   result.Logs,
	}

	if result.Err != nil {
		runtimeErr := result.Err.Error()
		exe.Error = &runtimeErr
	} else {
		enc, err := jsoncdc.Encode(result.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode to JSON-CDC")
		}

		exe.Value = string(enc)
	}

	err = s.store.InsertScriptExecution(&exe)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert script execution record")
	}

	return &exe, nil
}
