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

package e2eTest

import (
	client2 "github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransactionExecutions(t *testing.T) {
	addr1 := "0000000000000001"

	t.Run("Create execution for non-existent project", func(t *testing.T) {
		c := newClient()

		badID := uuid.New().String()

		var resp CreateTransactionExecutionResponse

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", badID),
			client2.Var("script", "transaction { execute { log(\"Hello, World!\") } }"),
		)

		assert.Error(t, err)
	})

	t.Run("Create execution without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionExecutionResponse

		const script = "transaction { execute { log(\"Hello, World!\") } }"

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
		)

		assert.Error(t, err)
	})

	t.Run("Create execution", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionExecutionResponse

		const script = "transaction { execute { log(\"Hello, World!\") } }"

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Empty(t, resp.CreateTransactionExecution.Errors)
		assert.Contains(t, resp.CreateTransactionExecution.Logs, "\"Hello, World!\"")
		assert.Equal(t, script, resp.CreateTransactionExecution.Script)
	})

	t.Run("Multiple executions", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { prepare(signer: AuthAccount) { AuthAccount(payer: signer) } }"

		err := c.Post(
			MutationCreateTransactionExecution,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.Var("signers", []string{addr1}),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Empty(t, respA.CreateTransactionExecution.Errors)
		require.Len(t, respA.CreateTransactionExecution.Events, 6)

		eventA := respA.CreateTransactionExecution.Events[5]

		// first account should have address 0x0a
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.JSONEq(t,
			`{"type":"Address","value":"0x000000000000000a"}`,
			eventA.Values[0],
		)

		var respB CreateTransactionExecutionResponse

		err = c.Post(
			MutationCreateTransactionExecution,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.Var("signers", []string{addr1}),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Empty(t, respB.CreateTransactionExecution.Errors)
		require.Len(t, respB.CreateTransactionExecution.Events, 6)

		eventB := respB.CreateTransactionExecution.Events[5]

		// second account should have address 0x07
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.JSONEq(t,
			`{"type":"Address","value":"0x000000000000000b"}`,
			eventB.Values[0],
		)
	})

	t.Run("Multiple executions with reset", func(t *testing.T) {
		// manually construct resolver
		c := newClient()
		project := createProject(t, c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { prepare(signer: AuthAccount) { AuthAccount(payer: signer) } }"

		err := c.Post(
			MutationCreateTransactionExecution,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.Var("signers", []string{addr1}),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Empty(t, respA.CreateTransactionExecution.Errors)
		require.Len(t, respA.CreateTransactionExecution.Events, 6)

		eventA := respA.CreateTransactionExecution.Events[5]

		// first account should have address 0x0a
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.JSONEq(t,
			`{"type":"Address","value":"0x000000000000000a"}`,
			eventA.Values[0],
		)

		_, err = c.projects.Reset(uuid.MustParse(project.ID))
		require.NoError(t, err)

		var respB CreateTransactionExecutionResponse

		err = c.Post(
			MutationCreateTransactionExecution,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.Var("signers", []string{addr1}),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Len(t, respB.CreateTransactionExecution.Events, 6)

		eventB := respB.CreateTransactionExecution.Events[5]

		// second account should have address 0x0a again due to reset
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.JSONEq(t,
			`{"type":"Address","value":"0x000000000000000a"}`,
			eventB.Values[0],
		)
	})

	t.Run("invalid (parse error)", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionExecutionResponse

		const script = `
          transaction(a: Int) {
        `

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Equal(t,
			[]model.ProgramError{
				{
					Message: "unexpected token: EOF",
					StartPosition: &model.ProgramPosition{
						Offset: 41,
						Line:   3,
						Column: 8,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 41,
						Line:   3,
						Column: 8,
					},
				},
			},
			resp.CreateTransactionExecution.Errors,
		)
		require.Empty(t, resp.CreateTransactionExecution.Logs)
	})

	t.Run("invalid (semantic error)", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionExecutionResponse

		const script = `
          transaction { execute { XYZ } }
        `

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Equal(t,
			[]model.ProgramError{
				{
					Message: "cannot find variable in this scope: `XYZ`",
					StartPosition: &model.ProgramPosition{
						Offset: 35,
						Line:   2,
						Column: 34,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 37,
						Line:   2,
						Column: 36,
					},
				},
			},
			resp.CreateTransactionExecution.Errors,
		)
		require.Empty(t, resp.CreateTransactionExecution.Logs)
	})

	t.Run("invalid (run-time error)", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionExecutionResponse

		const script = `
          transaction { execute { panic("oh no") } }
        `

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Equal(t,
			[]model.ProgramError{
				{
					Message: "panic: oh no",
					StartPosition: &model.ProgramPosition{
						Offset: 35,
						Line:   2,
						Column: 34,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 48,
						Line:   2,
						Column: 47,
					},
				},
			},
			resp.CreateTransactionExecution.Errors,
		)
		require.Empty(t, resp.CreateTransactionExecution.Logs)
	})

	t.Run("exceeding computation limit", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionExecutionResponse

		const script = `
          transaction {
              execute {
                  var i = 0
                  while i < 1_000_000 {
                      i = i + 1
                  }
              }
          }
        `

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, script, resp.CreateTransactionExecution.Script)
		require.Equal(t,
			"[Error Code: 1110] computation exceeds limit (100000)",
			resp.CreateTransactionExecution.Errors[0].Message,
		)
	})

	t.Run("argument", func(t *testing.T) {
		t.Parallel()
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionExecutionResponse

		const script = `
          transaction(a: Int) {
              execute {
                  log(a)
              }
          }
        `

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.Var("arguments", []string{
				`{"type": "Int", "value": "42"}`,
			}),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Empty(t, resp.CreateTransactionExecution.Errors)
		require.Equal(t, resp.CreateTransactionExecution.Logs, []string{"42"})
	})
}
