package bash_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bash_test "github.com/devigned/apmz/internal/test/bash"
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
				fmt.Println(lines)
				assert.Equal(t, 3, len(lines))
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
			wd, err := os.Getwd()
			require.NoError(t, err)
			fmt.Println(wd)

			cmd.Env = append(c.env, fmt.Sprintf("__TMP_APMZ_BATCH_FILE=%s", eventFileName))
			err = cmd.Run()
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

	scriptFile, err := ioutil.TempFile("", "apmz_test_script.*.sh")
	require.NoError(t, err)

	require.NoError(t, os.Chmod(scriptFile.Name(), 0700))

	return scriptFile.Name(), eventFile.Name(), func() {
		_ = os.Remove(eventFile.Name())
		_ = os.Remove(scriptFile.Name())
	}
}

func writeTestScript(t *testing.T, scriptFileName string, tmplArgs testScriptInput) {
	bits, err := bash_test.Asset("cmd/bash/testdata/base_script.gosh")
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
