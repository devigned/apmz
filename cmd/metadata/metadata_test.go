package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetadataCommandGroup(t *testing.T) {
	root, err := NewMetadataCommandGroup(nil)
	require.NoError(t, err)

	expected := []string{"instance", "attest", "events", "token"}
	actual := make([]string, len(root.Commands()))
	for i, c := range root.Commands() {
		actual[i] = c.Name()
	}
	assert.ElementsMatch(t, expected, actual)
}