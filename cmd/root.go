package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/devigned/apmz-sdk/apmz"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/devigned/apmz/cmd/bash"
	"github.com/devigned/apmz/cmd/batch"
	"github.com/devigned/apmz/cmd/metadata"
	"github.com/devigned/apmz/cmd/metric"
	timecmd "github.com/devigned/apmz/cmd/time"
	"github.com/devigned/apmz/cmd/trace"
	"github.com/devigned/apmz/cmd/uuid"
	"github.com/devigned/apmz/pkg/azmeta"
	"github.com/devigned/apmz/pkg/format"
	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
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

	var apiKeys []string
	var toOutput bool
	rootCmd.PersistentFlags().StringSliceVar(&apiKeys, "api-keys", nil, "comma separated keys for the Application Insights accounts to send to; eg 'key1,key2,key3'")
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
				if apiKeys == nil && !toOutput {
					err = errors.New("must provide api-keys")
					return
				}

				clients := make([]apmz.TelemetryClient, len(apiKeys))
				for i, key := range apiKeys {
					clients[i] = apmz.NewTelemetryClient(key)
				}

				clientProxy := service.APMZProxy{
					Clients: clients,
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
		APIKeysFactory: func() []string {
			return apiKeys
		},
		MetadataFactory: func() (service.Metadater, error) {
			return azmeta.New()
		},
	}

	cmdFuncs := []func(locator service.CommandServicer) (*cobra.Command, error){
		trace.NewTraceCommand,
		metric.NewMetricCommand,
		batch.NewBatchCommand,
		bash.NewBashCommand,
		timecmd.NewTimeCommandGroup,
		uuid.NewUUIDCommand,
		metadata.NewMetadataCommandGroup,
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

	rootCmd.PersistentPostRunE = xcobra.PostRunWithCtxE(func(ctx context.Context, cmd *cobra.Command, args []string) error {
		if apmer == nil {
			return nil
		}

		apmer.Close(ctx)
		return nil
	})

	return rootCmd, nil
}
