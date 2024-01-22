package osmoassert

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var diffTypesErrorMessage = "cannot compare variables of different types"

// nolint: thelper
func failNowIfNot(t *testing.T, ok bool) {
	if !ok {
		require.FailNow(t, diffTypesErrorMessage)
	}
}
