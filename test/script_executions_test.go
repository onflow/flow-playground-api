package test

import (
	client2 "github.com/dapperlabs/flow-playground-api/test/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/dapperlabs/flow-playground-api/model"
)

func TestScriptExecutions(t *testing.T) {
	t.Parallel()

	t.Run("valid, no return value", func(t *testing.T) {
		t.Parallel()

		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptExecutionResponse

		const script = "pub fun main() { }"

		err := c.Post(
			MutationCreateScriptExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)

		require.NoError(t, err)
		require.Empty(t, resp.CreateScriptExecution.Errors)
	})

	t.Run("invalid (parse error)", func(t *testing.T) {
		t.Parallel()

		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptExecutionResponse

		const script = "pub fun main() {"

		err := c.Post(
			MutationCreateScriptExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)

		require.NoError(t, err)
		assert.Equal(t, script, resp.CreateScriptExecution.Script)
		require.Equal(t,
			[]model.ProgramError{
				{
					Message: "expected token '}'",
					StartPosition: &model.ProgramPosition{
						Offset: 16,
						Line:   1,
						Column: 16,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 16,
						Line:   1,
						Column: 16,
					},
				},
			},
			resp.CreateScriptExecution.Errors,
		)
	})

	t.Run("invalid (semantic error)", func(t *testing.T) {
		t.Parallel()

		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptExecutionResponse

		const script = "pub fun main() { XYZ }"

		err := c.Post(
			MutationCreateScriptExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)

		require.NoError(t, err)
		assert.Equal(t, script, resp.CreateScriptExecution.Script)
		require.Equal(t,
			[]model.ProgramError{
				{
					Message: "cannot find variable in this scope: `XYZ`",
					StartPosition: &model.ProgramPosition{
						Offset: 17,
						Line:   1,
						Column: 17,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 19,
						Line:   1,
						Column: 19,
					},
				},
			},
			resp.CreateScriptExecution.Errors,
		)
	})

	t.Run("invalid (run-time error)", func(t *testing.T) {
		t.Parallel()

		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptExecutionResponse

		const script = "pub fun main() { panic(\"oh no\") }"

		err := c.Post(
			MutationCreateScriptExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)

		require.NoError(t, err)
		assert.Equal(t, script, resp.CreateScriptExecution.Script)
		require.Equal(t,
			[]model.ProgramError{
				{
					Message: "panic: oh no",
					StartPosition: &model.ProgramPosition{
						Offset: 17,
						Line:   1,
						Column: 17,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 30,
						Line:   1,
						Column: 30,
					},
				},
			},
			resp.CreateScriptExecution.Errors,
		)
	})

	t.Run("exceeding computation limit", func(t *testing.T) {
		t.Parallel()

		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptExecutionResponse

		const script = `
          pub fun main() {
              var i = 0
              while i < 1_000_000 {
                  i = i + 1
              }
          }
        `

		err := c.Post(
			MutationCreateScriptExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)

		require.NoError(t, err)
		assert.Equal(t, script, resp.CreateScriptExecution.Script)
		require.Equal(t,
			[]model.ProgramError{
				{
					Message: "[Error Code: 1110] computation exceeds limit (100000)",
					StartPosition: &model.ProgramPosition{
						Offset: 106,
						Line:   5,
						Column: 18,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 114,
						Line:   5,
						Column: 26,
					},
				},
			},
			resp.CreateScriptExecution.Errors,
		)
	})

	t.Run("return address", func(t *testing.T) {
		t.Parallel()

		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptExecutionResponse

		const script = "pub fun main(): Address { return 0x1 as Address }"

		err := c.Post(
			MutationCreateScriptExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.AddCookie(c.SessionCookie()),
		)

		require.NoError(t, err)
		assert.Equal(t, script, resp.CreateScriptExecution.Script)
		require.Empty(t, resp.CreateScriptExecution.Errors)
		assert.JSONEq(t,
			`{"type":"Address","value":"0x0000000000000001"}`,
			resp.CreateScriptExecution.Value,
		)
	})

	t.Run("argument", func(t *testing.T) {
		t.Parallel()

		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptExecutionResponse

		const script = "pub fun main(a: Int): Int { return a + 1 }"

		err := c.Post(
			MutationCreateScriptExecution,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("script", script),
			client2.Var("arguments", []string{
				`{"type":"Int","value":"2"}`,
			}),
			client2.AddCookie(c.SessionCookie()),
		)

		require.NoError(t, err)
		assert.Equal(t, script, resp.CreateScriptExecution.Script)
		require.Empty(t, resp.CreateScriptExecution.Errors)
		assert.JSONEq(t,
			`{"type":"Int","value":"3"}`,
			resp.CreateScriptExecution.Value,
		)
	})
}
