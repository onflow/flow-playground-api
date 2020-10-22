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

func (s *Scripts) CreateTemplate(projectID uuid.UUID, input model.NewScriptTemplate) (*model.ScriptTemplate, error) {
	tpl := model.ScriptTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: projectID,
		},
		Title:  input.Title,
		Script: input.Script,
	}

	err := s.store.InsertScriptTemplate(&tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store script template")
	}

	return &tpl, nil
}

func (s *Scripts) UpdateTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error {
	err := s.store.UpdateScriptTemplate(input, tpl)
	if err != nil {
		return errors.Wrap(err, "failed to update script template")
	}

	return nil
}

func (s *Scripts) DeleteTemplate(scriptID, projectID uuid.UUID) error {
	err := s.store.DeleteScriptTemplate(model.NewProjectChildID(scriptID, projectID))
	if err != nil {
		return errors.Wrap(err, "failed to delete script template")
	}

	return nil
}

func (s *Scripts) CreateExecution(
	proj *model.InternalProject,
	script string,
	arguments []string,
) (
	*model.ScriptExecution,
	error,
) {

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
		arguments,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute script")
	}

	exe := model.ScriptExecution{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: proj.ID,
		},
		Script:    script,
		Arguments: arguments,
		Logs:      result.Logs,
	}

	if result.Err != nil {
		exe.Errors = compute.ExtractProgramErrors(result.Err)
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
