package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

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
			apmzer, err := sl.GetAPMer()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to create App Insight client: %v", err)
				return err
			}

			reader := io.Reader(os.Stdin)
			if oArgs.FilePath != "" {
				bits, err := ioutil.ReadFile(oArgs.FilePath)
				if err != nil {
					sl.GetPrinter().ErrPrintf("unable to read file: %v", err)
					return err
				}
				reader = bytes.NewReader(bits)
			}

			eventsBits, err := ioutil.ReadAll(reader)
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to read: %v", err)
				return err
			}

			lines := strings.Split(string(eventsBits), "\n")
			sent := 0
			for _, l := range lines {
				var evt service.Event

				if strings.TrimSpace(l) == "" {
					continue
				}

				if err := json.Unmarshal([]byte(l), &evt); err != nil {
					sl.GetPrinter().ErrPrintf("unable to unmarshal events: %v -- \n%v", err, l)
					return err
				}

				apmzer.Track(evt.Item)
				sent++
			}

			return sl.GetPrinter().Print(struct{ Result string }{Result: fmt.Sprintf("sent %d events", sent)})
		}),
	}

	f := cmd.Flags()
	f.StringVarP(&oArgs.FilePath, "file-path", "f", "", "file path to json events -- if not specified, then stdin will be assumed")
	return cmd, nil
}
