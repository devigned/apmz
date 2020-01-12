package time

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

// newUnixNanoCommand creates a new `apmz time unixnano` command
func newUnixNanoCommand(sl service.CommandServicer) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "unixnano",
		Short: "prints the current unix nano time",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			sl.GetPrinter().Printf("%d", time.Now().UnixNano())
			return nil
		}),
	}

	return cmd, nil
}
