package metadata

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	attestArgs struct {
		Nonce string
	}
)

// NewMetadataCommandGroup will create a new command group for metadata functions
func NewMetadataCommandGroup(sl service.CommandServicer) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:              "metadata",
		Short:            "Azure instance metadata service related commands",
		TraverseChildren: true,
	}

	cmdFuncs := []func(locator service.CommandServicer) (*cobra.Command, error){
		newInstanceCommand,
		newAttestationCommand,
		NewEventsCommandGroup,
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

// newInstanceCommand creates a new `apmz metadata instance` command
func newInstanceCommand(sl service.CommandServicer) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "fetches the instance information via the instance metadata service",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			m, err := sl.GetMetadater()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to build Metadater service: %v", err)
			}

			instance, err := m.GetInstance(ctx)
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to fetch instance metadata: %v", err)
			}

			return sl.GetPrinter().Print(instance)
		}),
	}

	return cmd, nil
}

// newAttestationCommand creates a new `apmz metadata attest` command
func newAttestationCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs attestArgs
	cmd := &cobra.Command{
		Use:   "attest",
		Short: "requests attestation from the metadata instance service",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			m, err := sl.GetMetadater()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to build Metadater service: %v", err)
			}

			attest, err := m.GetAttestation(ctx, oArgs.Nonce)
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to fetch attest metadata: %v", err)
			}

			return sl.GetPrinter().Print(attest)
		}),
	}

	cmd.Flags().StringVarP(&oArgs.Nonce, "nonce", "n", "", "nonce is optional and must be digits with a max len of 10; eg '1234567890'")
	return cmd, nil
}
