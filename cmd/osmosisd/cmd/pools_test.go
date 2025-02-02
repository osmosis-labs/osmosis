package cmd

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPoolsCmd(t *testing.T) {
	cmd := GetPoolsCmd()
	
	t.Run("command structure", func(t *testing.T) {
		assert.NotNil(t, cmd)
		assert.Equal(t, "pools", cmd.Use)
		assert.Equal(t, "Query liquidity pool information", cmd.Short)
		assert.Equal(t, cobra.MaximumNArgs(1), cmd.Args)
	})

	t.Run("query flags", func(t *testing.T) {
		flags := cmd.Flags()
		assert.True(t, flags.HasAvailableFlags())
		
		hasOutputFlag := false
		flags.VisitAll(func(flag *pflag.Flag) {
			if flag.Name == flags.FlagOutput {
				hasOutputFlag = true
			}
		})
		assert.True(t, hasOutputFlag)
	})

	t.Run("invalid pool id", func(t *testing.T) {
		cmd.SetArgs([]string{"invalid"})
		err := cmd.ExecuteContext(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pool-id invalid not a valid uint")
	})

	t.Run("valid pool id", func(t *testing.T) {
		cmd.SetArgs([]string{"1"})
		err := cmd.ExecuteContext(context.Background())
		// Note: This will fail without a running node, which is expected
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("no args - query all pools", func(t *testing.T) {
		cmd.SetArgs([]string{})
		err := cmd.ExecuteContext(context.Background())
		// Note: This will fail without a running node, which is expected
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})
} 
