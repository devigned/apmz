package bash

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocks "github.com/devigned/apmz/internal/test"
)

func TestNewBashCommand(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T) *mocks.ServiceMock
		assertions func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "CommandConstruction",
			setup: func(t *testing.T) *mocks.ServiceMock {
				return nil
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.Equal(t, "bash", cmd.Name())
				assert.NotNil(t, cmd.Flags().Lookup("disabled"))
				assert.NotNil(t, cmd.Flags().Lookup("name"))
				assert.NotNil(t, cmd.Flags().Lookup("default-tags"))
			},
		},
		{
			name: "WithEnabled",
			setup: func(t *testing.T) *mocks.ServiceMock {
				s := serviceWithKey()
				p := new(mocks.PrinterMock)
				p.On("Printf", mock.MatchedBy(func(script string) bool {
					return strings.Contains(script, "__TMP_APMZ_BATCH_FILE=\"$(mktemp /tmp/apmz.XXXXXX)")
				}), mock.Anything)
				s.On("GetPrinter").Return(p)
				return s
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.Equal(t, "bash", cmd.Name())
				assert.NoError(t, cmd.Execute())
			},
		},
		{
			name: "WithDisabled",
			setup: func(t *testing.T) *mocks.ServiceMock {
				s := serviceWithKey()
				p := new(mocks.PrinterMock)
				p.On("Printf", mock.MatchedBy(func(script string) bool {
					return strings.Contains(script, "trace_err") && strings.Contains(script, "apmz")
				}), mock.Anything)
				s.On("GetPrinter").Return(p)
				return s
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.NoError(t, cmd.Execute())
			},
		},
		{
			name: "WithScriptDefaults",
			setup: func(t *testing.T) *mocks.ServiceMock {
				s := serviceWithKey()
				p := new(mocks.PrinterMock)
				p.On("Printf", mock.MatchedBy(func(script string) bool {
					return strings.Contains(script, `__SCRIPT_NAME="${__SCRIPT_NAME:-testcmd}"`) &&
						strings.Contains(script, `__DEFAULT_TAGS="${__DEFAULT_TAGS:-foo1=bar,bin=baz}"`)

				}), mock.Anything)
				s.On("GetPrinter").Return(p)
				return s
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.Equal(t, "bash", cmd.Name())
				cmd.SetArgs([]string{"--name", "testcmd", "--default-tags", "foo1=bar,bin=baz"})
				assert.NoError(t, cmd.Execute())
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := NewBashCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}

func serviceWithKey() *mocks.ServiceMock {
	s := new(mocks.ServiceMock)
	s.On("GetKey").Return("foo")
	return s
}
