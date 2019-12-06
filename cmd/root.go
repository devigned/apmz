package cmd

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/devigned/apmz-sdk/apmz"

	"github.com/devigned/apmz/cmd/metric"
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pub.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "App Insights API key")

	var once sync.Once
	var apmer service.APMer
	registry := &service.Registry{
		APMerFactory: func() (service.APMer, error) {
			once.Do(func() {
				apmer = apmz.NewTelemetryClient(apiKey)
			})
			return apmer, nil
		},
		PrinterFactory: func() format.Printer {
			return &format.StdPrinter{
				Format: format.JSONFormat,
			}
		},
	}

	cmdFuncs := []func(locator service.CommandServicer) (*cobra.Command, error){
		trace.NewTraceCommand,
		metric.NewMetricCommand,
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
		case <-time.After(3 * time.Second):
			return errors.New("failed to flush events to Application Insights")
		case <-apmer.Channel().Close(2 * time.Second):
		}

		return nil
	}

	return rootCmd, nil
}
