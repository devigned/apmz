package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmdNames(t *testing.T) {
	root, err := newRootCommand()
	require.NoError(t, err)

	expected := []string{"trace", "metric", "batch", "version", "bash", "time", "uuid", "metadata"}
	actual := make([]string, len(root.Commands()))
	for i, c := range root.Commands() {
		actual[i] = c.Name()
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestExecute(t *testing.T) {
	root, err := newRootCommand()
	require.NoError(t, err)
	assert.NoError(t, root.Execute())
}
