package stableswap

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	usdaDenom = "usda"
	usdbDenom = "usdb"
	usdcDenom = "usdc"
	usdtDenom = "usdt"
	ustDenom  = "ust"
)

func createTestPool(t *testing.T, poolLiquidity sdk.Coins, swapFee, exitFee sdk.Dec) types.PoolI {
	pool, err := NewStableswapPool(1, PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, poolLiquidity, "")

	require.NoError(t, err)

	return &pool
}

func createTestContext(t *testing.T) sdk.Context {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	ms := rootmulti.NewStore(db, logger)

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger)
}
