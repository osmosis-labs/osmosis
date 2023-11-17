package routertesting_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/routertesting"
)

func TestReadPoolsFileFromState(t *testing.T) {
	pools, err := routertesting.ReadPools()
	require.NoError(t, err)

	require.NotEmpty(t, pools)
	require.Greater(t, len(pools), 500)
}
