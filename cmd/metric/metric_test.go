package metric

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	mocks "github.com/devigned/apmz/internal/test"
)

func TestNewMetricCommand(t *testing.T) {
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
				assert.Equal(t, "metric", cmd.Name())
				name := cmd.Flags().Lookup("name")
				if assert.NotNil(t, name) {
					assert.Equal(t, name.Shorthand, "n")
				}
				tags := cmd.Flags().Lookup("tags")
				if assert.NotNil(t, tags) {
					assert.Equal(t, tags.Shorthand, "t")
				}
				value := cmd.Flags().Lookup("value")
				if assert.NotNil(t, value) {
					assert.Equal(t, value.Shorthand, "v")
				}
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := NewMetricCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}
