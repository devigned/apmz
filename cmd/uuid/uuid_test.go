package uuid_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	gid "github.com/google/uuid"

	"github.com/devigned/apmz/cmd/uuid"
	mocks "github.com/devigned/apmz/internal/test"
)

func TestUUIDCommand(t *testing.T) {
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
				assert.Equal(t, "uuid", cmd.Name())
			},
		},
		{
			name: "PrintsNewUUID",
			setup: func(t *testing.T) *mocks.ServiceMock {
				s := new(mocks.ServiceMock)
				p := new(mocks.PrinterMock)
				p.On("Printf", "%s", mock.MatchedBy(func(output []interface{}) bool {
					require.Len(t, output, 1)
					firstArg, ok := output[0].(string)
					if !ok {
						return false
					}

					_, err := gid.Parse(firstArg)
					if err != nil {
						return false
					}

					return true
				}))
				s.On("GetPrinter").Return(p)
				return s
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.NoError(t, cmd.Execute())
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := uuid.NewUUIDCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}
