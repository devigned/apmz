package time

import (
	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/service"
)

// NewTimeCommandGroup will create a new command group for time functions
func NewTimeCommandGroup(sl service.CommandServicer) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:              "time",
		Short:            "time related commands",
		TraverseChildren: true,
	}

	cmdFuncs := []func(locator service.CommandServicer) (*cobra.Command, error){
		newUnixNanoCommand,
		newDiffUnixNanoCommand,
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
