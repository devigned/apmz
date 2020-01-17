package trace

import (
	"context"

	"github.com/devigned/apmz-sdk/apmz"
	"github.com/devigned/apmz-sdk/apmz/contracts"
	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	traceArgs struct {
		Name  string
		Level int
		Tags  map[string]string
		Out   bool
	}
)

// NewTraceCommand creates a new `apmz trace` command
func NewTraceCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs traceArgs
	cmd := &cobra.Command{
		Use:   "trace",
		Short: "send a trace event (traces) to Application Insights",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			trace := apmz.NewTraceTelemetry(oArgs.Name, contracts.SeverityLevel(oArgs.Level))
			for k, v := range oArgs.Tags {
				trace.Properties[k] = v
			}

			apmer, err := sl.GetAPMer()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to create App Insight client: %v\n", err)
				return err
			}

			apmer.Track(trace)
			return nil
		}),
	}

	f := cmd.Flags()
	f.IntVarP(&oArgs.Level, "level", "l", 0, "severity level for the event")
	f.StringVarP(&oArgs.Name, "name", "n", "", "trace event name")
	f.StringToStringVarP(&oArgs.Tags, "tags", "t", map[string]string{}, "custom tags to be applied to the trace formatted as key=value")
	err := cmd.MarkFlagRequired("name")
	return cmd, err
}
