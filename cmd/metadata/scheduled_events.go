package metadata

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/azmeta"
	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	ackArgs struct {
		EventIds []string
	}
)

// NewEventsCommandGroup will create a new command group for scheduled events functions
func NewEventsCommandGroup(sl service.CommandServicer) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:              "events",
		Short:            "Azure instance metadata service scheduled events related commands",
		TraverseChildren: true,
	}

	cmdFuncs := []func(locator service.CommandServicer) (*cobra.Command, error){
		NewScheduledEventsCommand,
		NewScheduledEventsAckCommand,
	}

	for _, f := range cmdFuncs {
		cmd, err := f(sl)
		if err != nil {
			return rootCmd, err
		}
		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

// NewScheduledEventsCommand creates a new `apmz metadata events get` command
func NewScheduledEventsCommand(sl service.CommandServicer) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "fetches the scheduled events for the machine",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			m, err := sl.GetMetadater()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to build Metadater service: %v", err)
				return err
			}

			se, err := m.GetScheduledEvents(ctx)
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to fetch scheduled events metadata: %v", err)
				return err
			}

			return sl.GetPrinter().Print(se)
		}),
	}

	return cmd, nil
}

// NewScheduledEventsAckCommand creates a new `apmz metadata events ack` command
func NewScheduledEventsAckCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs ackArgs
	cmd := &cobra.Command{
		Use:   "ack",
		Short: "acknowledging an outstanding event which indicates to Azure that it can shorten the minimum notification time",
		Long: "Acknowledging an event allows the event to proceed for all Resources in the event, not just the " +
			"virtual machine that acknowledges the event. You may therefore choose to elect a leader to coordinate " +
			"the acknowledgement, which may be as simple as the first machine in the Resources field.",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			m, err := sl.GetMetadater()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to build Metadater service: %v", err)
				return err
			}

			ackEvents := make([]azmeta.AckEvent, len(oArgs.EventIds))
			for i, id := range oArgs.EventIds {
				ackEvents[i].EventID = id
			}

			ack := azmeta.AckEvents{StartRequests:ackEvents}
			if err := m.AckScheduledEvents(ctx, ack); err != nil {
				sl.GetPrinter().ErrPrintf("unable to fetch scheduled events metadata: %v", err)
				return err
			}

			return nil
		}),
	}

	cmd.Flags().StringSliceVarP(&oArgs.EventIds, "event-ids", "e", nil, "Event IDs to acknowledge to Azure ( -e 'id1,id2,...')")
	err := cmd.MarkFlagRequired("event-ids")
	return cmd, err
}