package bash_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bashtest "github.com/devigned/apmz/internal/test/bash"
	"github.com/devigned/apmz/pkg/service"
)

type (
	testScriptInput struct {
		Args   string
		BinDir string
		Script string
	}
)

func TestNewBashCommandEnabled(t *testing.T) {
	cases := []struct {
		name       string
		env        []string
		args       []string
		script     string
		assertions func(t *testing.T, stdout, stderr, eventFilePath string)
	}{
		{
			name: "RunWithDefaultSettings",
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				warning := `Warning: apmz event collection is enabled, but --api-key is not specified. You must override
the __APP_INSIGHTS_KEY env var or events will not be set to Application Insights on script exit.
`
				assert.Equal(t, warning, stderr)
				assert.Equal(t, "", stdout)
				assert.Error(t, err, "should not find file because the script should have cleaned it up")
			},
		},
		{
			name: "RunWithPreserveTmpFile",
			env:  []string{"__PRESERVE_TMP_FILE=true"},
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				assert.NoError(t, err)
			},
		},
		{
			name: "ShouldHaveTwoEventsByDefaultInEventFile",
			env:  []string{"__PRESERVE_TMP_FILE=true", "__DEFAULT_TAGS=foo=bar"},
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				require.NoError(t, err)
				lines := readEventFile(t, eventFilePath)
				assert.Equal(t, 2, len(lines))
			},
		},
		{
			name:   "ShouldHaveThreeEventsWhenInvokingAnErrorTrace",
			env:    []string{"__PRESERVE_TMP_FILE=true", "__DEFAULT_TAGS=foo=bar"},
			script: `trace_err "error_event"`,
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				require.NoError(t, err)
				lines := readEventFile(t, eventFilePath)
				assert.Equal(t, 3, len(lines))
			},
		},
		{
			name:   "ShouldHaveThreeEventsWhenInvokingAnInfoTrace",
			env:    []string{"__PRESERVE_TMP_FILE=true", "__DEFAULT_TAGS=foo=bar"},
			script: `trace_info "info_event"`,
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				require.NoError(t, err)
				lines := readEventFile(t, eventFilePath)
				assert.Equal(t, 3, len(lines))
			},
		},
		{
			name:   "ShouldHaveThreeEventsWhenInvokingAMetricEvent",
			env:    []string{"__PRESERVE_TMP_FILE=true", "__DEFAULT_TAGS=foo=bar"},
			script: `time_metric "time_sleep" sleep 1`,
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				require.NoError(t, err)
				lines := readEventFile(t, eventFilePath)
				assert.Equal(t, 3, len(lines))
			},
		},
		{
			name: "WithKeyAsArgs",
			env:  []string{"__PRESERVE_TMP_FILE=true"},
			args: []string{"--api-key", "foo"},
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				require.NoError(t, err)
				lines := readEventFile(t, eventFilePath)
				assert.Equal(t, 2, len(lines))
				assert.Empty(t, stderr)
			},
		},
		{
			name: "WithNameAsArgs",
			env:  []string{"__PRESERVE_TMP_FILE=true"},
			args: []string{"-n", "helloworld"},
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				require.NoError(t, err)
				lines := readEventFile(t, eventFilePath)
				assert.Equal(t, 2, len(lines))
				assert.Contains(t, lines[0], `"Message":"helloworld-exit"`)
				assert.Contains(t, lines[0], `"code":"0"`)
				assert.Contains(t, lines[1], "helloworld-duration")
			},
		},
		{
			name: "WithDefaultTagsAsArgs",
			env:  []string{"__PRESERVE_TMP_FILE=true"},
			args: []string{"-t", "foo=bar,fast=slow"},
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				require.NoError(t, err)
				lines := readEventFile(t, eventFilePath)
				events := eventsFromLines(t, lines)
				assert.Equal(t, 2, len(events))

				assert.Equal(t, events[0].Type, "TraceTelemetry")
				props0 := events[0].Item.GetProperties()
				assert.Equal(t, props0["code"], "0")
				assert.Equal(t, props0["fast"], "slow")
				assert.Equal(t, props0["foo"], "bar")
				assert.NotEmpty(t, props0["correlation_id"])

				assert.Equal(t, events[1].Type, "MetricTelemetry")
				props1 := events[1].Item.GetProperties()
				assert.Equal(t, props1["fast"], "slow")
				assert.Equal(t, props1["foo"], "bar")
				assert.NotEmpty(t, props1["correlation_id"])

				assert.Equal(t, props0["correlation_id"], props1["correlation_id"])
			},
		},
		{
			name:   "HasSessionIDSet",
			script: "echo $__SCRIPT_SESSION_ID",
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				stdout = strings.TrimSuffix(stdout, "\n")
				_, err := uuid.Parse(stdout)
				assert.NoError(t, err, stdout)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			scriptFileName, eventFileName, del := generateTmpFiles(t)
			defer del()

			writeTestScript(t, scriptFileName, testScriptInput{
				Args:   strings.Join(c.args, " "),
				Script: c.script,
				BinDir: "../../bin",
			})

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, scriptFileName)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Env = append(c.env, fmt.Sprintf("__TMP_APMZ_BATCH_FILE=%s", eventFileName))
			err := cmd.Run()
			outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
			if err != nil {
				require.NoError(t, err, errStr)
			}
			c.assertions(t, outStr, errStr, eventFileName)
		})
	}
}

func TestNewBashCommandDisabled(t *testing.T) {
	cases := []struct {
		name       string
		env        []string
		args       []string
		script     string
		assertions func(t *testing.T, stdout, stderr, eventFilePath string)
	}{
		{
			name: "RunWithDefaultSettings",
			args: []string{"-d"},
			script: `
trace_info "foo"
trace_err "bar"
time_metric "some_metric" echo "me"
`,
			assertions: func(t *testing.T, stdout, stderr, eventFilePath string) {
				_, err := os.Stat(eventFilePath)
				assert.Error(t, err, "file should not be made since apmz is disabled")
				assert.Equal(t, "me\n", stdout)
				assert.Empty(t, stderr)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			scriptFileName, eventFileName, del := generateTmpFiles(t)
			defer del()

			writeTestScript(t, scriptFileName, testScriptInput{
				Args:   strings.Join(c.args, " "),
				Script: c.script,
				BinDir: "../../bin",
			})

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, scriptFileName)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Env = append(c.env, fmt.Sprintf("__TMP_APMZ_BATCH_FILE=%s", eventFileName))
			err := cmd.Run()
			outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
			if err != nil {
				require.NoError(t, err, errStr)
			}
			c.assertions(t, outStr, errStr, eventFileName)
		})
	}
}

func generateTmpFiles(t *testing.T) (testScript, events string, del func()) {
	eventFile, err := ioutil.TempFile("", "apmz_events.*.json")
	require.NoError(t, err)
	require.NoError(t, eventFile.Close())
	_ = os.Remove(eventFile.Name()) // this will be made in the script, but we need a unique name

	scriptFile, err := ioutil.TempFile("", "apmz_test_script.*.sh")
	require.NoError(t, err)
	require.NoError(t, scriptFile.Close())

	require.NoError(t, os.Chmod(scriptFile.Name(), 0700))

	return scriptFile.Name(), eventFile.Name(), func() {
		_ = os.Remove(scriptFile.Name())
	}
}

func writeTestScript(t *testing.T, scriptFileName string, tmplArgs testScriptInput) {
	bits, err := bashtest.Asset("cmd/bash/testdata/base_script.gosh")
	require.NoError(t, err)

	tmpl, err := template.New("script").Parse(string(bits))
	require.NoError(t, err)

	b := bytes.Buffer{}
	require.NoError(t, tmpl.Execute(&b, tmplArgs))

	require.NoError(t, ioutil.WriteFile(scriptFileName, b.Bytes(), 0700))
}

func readEventFile(t *testing.T, eventFileName string) []string {
	bits, err := ioutil.ReadFile(eventFileName)
	require.NoError(t, err)

	str := strings.TrimSuffix(string(bits), "\n")
	return strings.Split(str, "\n")
}

func eventsFromLines(t *testing.T, lines []string) []service.Event {
	events := make([]service.Event, len(lines))
	for i, line := range lines {
		var event service.Event
		require.NoError(t, json.Unmarshal([]byte(line), &event))
		events[i] = event
	}
	return events
}
