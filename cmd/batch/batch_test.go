package batch

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	mocks "github.com/devigned/apmz/internal/test"
)

func TestNewBatchCommand(t *testing.T) {
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
				assert.Equal(t, "batch", cmd.Name())
				fp := cmd.Flags().Lookup("file-path")
				if assert.NotNil(t, fp) {
					assert.Equal(t, fp.Shorthand, "f")
				}
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := NewBatchCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}
