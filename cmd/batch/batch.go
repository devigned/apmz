package batch

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	batchArgs struct {
		FilePath string
	}
)

// NewBatchCommand creates a new `apmz batch` command
func NewBatchCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs batchArgs
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "upload a batch of telemetry to Application Insights",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			_, err := sl.GetAPMer()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to create App Insight client: %v", err)
				return err
			}

			return nil
		}),
	}

	f := cmd.Flags()
	f.StringVarP(&oArgs.FilePath, "file-path", "f", "", "file path to json events -- if not specified, then stdin will be assumed")
	return cmd, nil
}
