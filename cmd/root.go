package cmd

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/devigned/apmz-sdk/apmz"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/devigned/apmz/cmd/bash"
	"github.com/devigned/apmz/cmd/batch"
	"github.com/devigned/apmz/cmd/metric"
	timecmd "github.com/devigned/apmz/cmd/time"
	"github.com/devigned/apmz/cmd/trace"
	"github.com/devigned/apmz/pkg/format"
	"github.com/devigned/apmz/pkg/service"
)

func init() {
	_ = godotenv.Load() // load if possible
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})
}

// Execute kicks off the command line
func Execute() {
	cmd, err := newRootCommand()
	if err != nil {
		log.Fatalf("fatal error: commands failed to build! %v", err)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:              "apmz",
		Short:            "apmz provides a command line interface for the Azure Application Insights",
		TraverseChildren: true,
	}

	var apiKey string
	var cfgFile string
	var toOutput bool
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pub.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "App Insights API key")
	rootCmd.PersistentFlags().BoolVarP(&toOutput, "output", "o", false, "instead of sending directly to Application Insights, output event to stdout as json")

	var once sync.Once
	var apmer service.APMer
	printer := &format.StdPrinter{
		Format: format.JSONFormat,
	}
	registry := &service.Registry{
		APMerFactory: func() (service.APMer, error) {
			var err error
			once.Do(func() {
				if apiKey == "" && !toOutput {
					err = errors.New("must provide api-key")
					return
				}

				clientProxy := service.APMZProxy{
					TelemetryClient: apmz.NewTelemetryClient(apiKey),
				}
				if toOutput {
					clientProxy.Printer = printer
				}
				apmer = clientProxy
			})
			return apmer, err
		},
		PrinterFactory: func() format.Printer {
			return printer
		},
		APIKeyFactory: func() string {
			return apiKey
		},
	}

	cmdFuncs := []func(locator service.CommandServicer) (*cobra.Command, error){
		trace.NewTraceCommand,
		metric.NewMetricCommand,
		batch.NewBatchCommand,
		bash.NewBashCommand,
		timecmd.NewTimeCommandGroup,
		func(locator service.CommandServicer) (*cobra.Command, error) {
			return newVersionCommand(), nil
		},
	}

	for _, f := range cmdFuncs {
		cmd, err := f(registry)
		if err != nil {
			return rootCmd, err
		}
		rootCmd.AddCommand(cmd)
	}

	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if apmer == nil {
			return nil
		}

		select {
		case <-time.After(10 * time.Second):
			return errors.New("failed to flush events to Application Insights")
		case <-apmer.Channel().Close(2 * time.Second):
		}

		return nil
	}

	return rootCmd, nil
}
