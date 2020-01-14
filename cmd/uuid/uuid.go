package uuid

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/google/uuid"

	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

// NewUUIDCommand will create a new `apmz uuid` command
func NewUUIDCommand(sl service.CommandServicer) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "uuid",
		Short: "generate a new uuid",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			id := uuid.New()
			sl.GetPrinter().Printf("%s", id.String())
			return nil
		}),
	}

	return cmd, nil
}
