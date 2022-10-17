// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/Masterminds/semver"
	"github.com/google/uuid"
)

type ContractDeployment struct {
	ID      uuid.UUID       `json:"id"`
	Script  string          `json:"script"`
	Address Address         `json:"address"`
	Errors  []*ProgramError `json:"errors"`
	Events  []*Event        `json:"events"`
	Logs    []string        `json:"logs"`
}

type ContractTemplate struct {
	ID     uuid.UUID `json:"id"`
	Index  int       `json:"index"`
	Title  string    `json:"title"`
	Script string    `json:"script"`
}

type Event struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

type NewContractDeployment struct {
	ProjectID uuid.UUID `json:"projectId"`
	Script    *string   `json:"script"`
	Address   Address   `json:"address"`
}

type NewContractTemplate struct {
	ProjectID uuid.UUID `json:"projectId"`
	Title     string    `json:"title"`
	Script    string    `json:"script"`
}

type NewFile struct {
	ProjectID uuid.UUID `json:"projectId"`
	Title     string    `json:"title"`
	Script    string    `json:"script"`
}

type NewProject struct {
	ParentID             *uuid.UUID                       `json:"parentId"`
	Title                string                           `json:"title"`
	Description          string                           `json:"description"`
	Readme               string                           `json:"readme"`
	Seed                 int                              `json:"seed"`
	Accounts             []string                         `json:"accounts"`
	TransactionTemplates []*NewProjectTransactionTemplate `json:"transactionTemplates"`
	ScriptTemplates      []*NewProjectScriptTemplate      `json:"scriptTemplates"`
	ContractTemplates    []*NewProjectContractTemplate    `json:"contractTemplates"`
}

type NewProjectContractTemplate struct {
	Title  string `json:"title"`
	Script string `json:"script"`
}

type NewProjectFile struct {
	Title  string `json:"title"`
	Script string `json:"script"`
}

type NewProjectScriptTemplate struct {
	Title  string `json:"title"`
	Script string `json:"script"`
}

type NewProjectTransactionTemplate struct {
	Title  string `json:"title"`
	Script string `json:"script"`
}

type NewScriptExecution struct {
	ProjectID uuid.UUID `json:"projectId"`
	Script    string    `json:"script"`
	Arguments []string  `json:"arguments"`
}

type NewScriptTemplate struct {
	ProjectID uuid.UUID `json:"projectId"`
	Title     string    `json:"title"`
	Script    string    `json:"script"`
}

type NewTransactionExecution struct {
	ProjectID uuid.UUID `json:"projectId"`
	Script    string    `json:"script"`
	Signers   []Address `json:"signers"`
	Arguments []string  `json:"arguments"`
}

type NewTransactionTemplate struct {
	ProjectID uuid.UUID `json:"projectId"`
	Title     string    `json:"title"`
	Script    string    `json:"script"`
}

type PlaygroundInfo struct {
	APIVersion     semver.Version `json:"apiVersion"`
	CadenceVersion semver.Version `json:"cadenceVersion"`
}

type ProgramError struct {
	Message       string           `json:"message"`
	StartPosition *ProgramPosition `json:"startPosition"`
	EndPosition   *ProgramPosition `json:"endPosition"`
}

type ProgramPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type UpdateContractTemplate struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	ProjectID uuid.UUID `json:"projectId"`
	Index     *int      `json:"index"`
	Script    *string   `json:"script"`
}

type UpdateFile struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	ProjectID uuid.UUID `json:"projectId"`
	Index     *int      `json:"index"`
	Script    *string   `json:"script"`
}

type UpdateProject struct {
	ID          uuid.UUID `json:"id"`
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Readme      *string   `json:"readme"`
	Persist     *bool     `json:"persist"`
}

type UpdateScriptTemplate struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	ProjectID uuid.UUID `json:"projectId"`
	Index     *int      `json:"index"`
	Script    *string   `json:"script"`
}

type UpdateTransactionTemplate struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	ProjectID uuid.UUID `json:"projectId"`
	Index     *int      `json:"index"`
	Script    *string   `json:"script"`
}
