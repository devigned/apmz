package time

import (
	"context"
	"math"
	"time"

	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	diffArgs struct {
		A      int64
		B      int64
		Format string
	}
)

// newDiffUnixNanoCommand creates a new `apmz time diff` command
func newDiffUnixNanoCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs diffArgs
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "absolute difference between to unixnano times",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			a := time.Unix(0, oArgs.A)
			b := time.Unix(0, oArgs.B)
			elapsed := a.Sub(b)
			switch oArgs.Format {
			case "nano":
				sl.GetPrinter().Printf("%d", int64(math.Abs(float64(elapsed.Nanoseconds()))))
			case "micro":
				sl.GetPrinter().Printf("%d", int64(math.Abs(float64(elapsed.Microseconds()))))
			case "ms":
				sl.GetPrinter().Printf("%d", int64(math.Abs(float64(elapsed.Milliseconds()))))
			case "sec":
				sl.GetPrinter().Printf("%f", math.Abs(elapsed.Seconds()))
			default:
				sl.GetPrinter().ErrPrintf("unknown time resolution %q", oArgs.Format)
			}
			return nil
		}),
	}

	f := cmd.Flags()
	f.Int64VarP(&oArgs.A, "first", "a", 0, "first time in unixnano format")
	f.Int64VarP(&oArgs.B, "second", "b", 0, "second time in unixnano format")
	f.StringVarP(&oArgs.Format, "resolution", "r", "sec", "time resolution [nano, micro, ms, sec]")

	return cmd, nil
}
