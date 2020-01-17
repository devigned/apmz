package metadata_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/devigned/apmz/cmd/metadata"
	mocks "github.com/devigned/apmz/internal/test"
	"github.com/devigned/apmz/pkg/azmeta"
)

func TestNewEventsCommandGroup(t *testing.T) {
	root, err := metadata.NewEventsCommandGroup(nil)
	require.NoError(t, err)

	expected := []string{"get", "ack"}
	actual := make([]string, len(root.Commands()))
	for i, c := range root.Commands() {
		actual[i] = c.Name()
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestNewScheduledEventsCommand(t *testing.T) {
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
				assert.Equal(t, "get", cmd.Name())
			},
		},
		{
			name: "WithNoArgs",
			setup: func(t *testing.T) *mocks.ServiceMock {
				sl := new(mocks.ServiceMock)
				p := new(mocks.PrinterMock)
				m := new(mocks.MetadataMock)
				var se azmeta.ScheduledEvents
				bits, err := ioutil.ReadFile("./testdata/scheduled_events.json")
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(bits, &se))
				p.On("Print", &se).Return(nil)
				m.On("GetScheduledEvents", mock.Anything, mock.Anything).Return(&se, nil)
				sl.On("GetPrinter").Return(p)
				sl.On("GetMetadater").Return(m, nil)
				return sl
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
			cmd, err := metadata.NewScheduledEventsCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}

func TestNewScheduledEventsAckCommand(t *testing.T) {
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
				assert.Equal(t, "ack", cmd.Name())
				if f := cmd.Flags().Lookup("event-ids"); assert.NotNil(t, f) {
					assert.Equal(t, "e", f.Shorthand)
				}
			},
		},
		{
			name: "WithEventArgs",
			setup: func(t *testing.T) *mocks.ServiceMock {
				sl := new(mocks.ServiceMock)
				m := new(mocks.MetadataMock)
				ack := azmeta.AckEvents{
					StartRequests: []azmeta.AckEvent{
						{
							EventID: "123",
						},
						{
							EventID: "456",
						},
					},
				}
				m.On("AckScheduledEvents", mock.Anything, ack, mock.Anything).Return(nil)
				sl.On("GetMetadater").Return(m, nil)
				return sl
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				cmd.SetArgs([]string{"-e", "123,456"})
				assert.NoError(t, cmd.Execute())
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := metadata.NewScheduledEventsAckCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}
