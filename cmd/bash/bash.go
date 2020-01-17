package bash

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/data"
	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	bashFlags struct {
		Disable     bool
		ScriptName  string
		DefaultTags map[string]string
	}
)

// NewBashCommand creates a new `apmz bash` command
func NewBashCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs bashFlags
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "prints a bash script to source which provides functionality for common tracing and metrics operations",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			script, err := data.Asset("data/enabled_bash.gosh")
			if err != nil {
				return err
			}
			if oArgs.Disable {
				script, err = data.Asset("data/disabled_bash.gosh")
				if err != nil {
					return err
				}
			} else {
				// enabled, so we need to have the AppInsightsKey set
				if sl.GetKeys() == nil || len(sl.GetKeys()) == 0 {
					warning := "Warning: apmz event collection is enabled, but --api-keys is not specified. You must override the __APP_INSIGHTS_KEY env var or events will not be set to Application Insights on script exit.\n"
					sl.GetPrinter().ErrPrintf(warning)
				}
			}

			var kvs []string
			for k, v := range oArgs.DefaultTags {
				kvs = append(kvs, fmt.Sprintf("%s=%s", k, v))
			}

			tags := strings.Join(kvs, ",")
			input := struct {
				ScriptName      string
				DefaultTags     string
				AppInsightsKeys string
			}{
				ScriptName:  oArgs.ScriptName,
				DefaultTags: tags,
			}

			if sl.GetKeys() != nil {
				input.AppInsightsKeys = strings.Join(sl.GetKeys(), ",")
			}

			tmpl, err := template.New("script").Parse(string(script))
			if err != nil {
				sl.GetPrinter().ErrPrintf("template would not parse: %v", err)
				return err
			}

			b := strings.Builder{}
			if err := tmpl.Execute(&b, input); err != nil {
				sl.GetPrinter().ErrPrintf("template would not execute: %v", err)
				return err
			}

			sl.GetPrinter().Printf(b.String())
			return nil
		}),
	}

	cmd.Flags().BoolVarP(&oArgs.Disable, "disabled", "d", false, "disable event collection; if disabled, then all script functions are defined, but do not collect events.")
	cmd.Flags().StringVarP(&oArgs.ScriptName, "name", "n", "script", "name of script for use in script start and exit events")
	cmd.Flags().StringToStringVarP(&oArgs.DefaultTags, "default-tags", "t", map[string]string{}, "default tags for all events and metrics formatted as key=value")
	return cmd, nil
}
