package metric

import (
	"context"

	"github.com/devigned/apmz-sdk/apmz"
	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	metricArgs struct {
		Name  string
		Value float64
		Tags  map[string]string
	}
)

// NewMetricCommand creates a new `apmz batch` command
func NewMetricCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs metricArgs
	cmd := &cobra.Command{
		Use:   "metric",
		Short: "send a metric (customMetrics) to Application Insights",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			apmer, err := sl.GetAPMer()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to create App Insight client: %v", err)
				return err
			}

			metric := apmz.NewMetricTelemetry(oArgs.Name, oArgs.Value)
			for k, v := range oArgs.Tags {
				metric.Properties[k] = v
			}

			apmer.Track(metric)
			return nil
		}),
	}

	f := cmd.Flags()
	f.Float64VarP(&oArgs.Value, "value", "v", 0, "value of the metric as a float64")
	f.StringVarP(&oArgs.Name, "name", "n", "", "trace event name")
	f.StringToStringVarP(&oArgs.Tags, "tags", "t", map[string]string{}, "custom tags to be applied to the trace formatted as key=value")
	err := cmd.MarkFlagRequired("name")
	return cmd, err
}
