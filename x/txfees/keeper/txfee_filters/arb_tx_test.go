package txfee_filters_test

import (
	"encoding/json"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/app/apptesting"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v21/x/txfees/keeper/txfee_filters"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestTxFeeFilters(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Tests that the arb filter is enabled on the affiliate swap msg.
func (suite *KeeperTestSuite) TestIsArbTxLooseAuthz_AffiliateSwapMsg() {
	affiliateSwapMsg := &txfee_filters.AffiliateSwapMsg{
		Swap: txfee_filters.Swap{
			FeeCollector:  "osmo1dldrxz5p8uezxz3qstpv92de7wgfp7hvr72dcm",
			FeePercentage: sdk.ZeroDec(),
			Routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1221,
					TokenOutDenom: "uosmo",
				},
				{
					PoolId:        3,
					TokenOutDenom: "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4",
				},
				{
					PoolId:        4,
					TokenOutDenom: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB",
				},
				{
					PoolId:        1251,
					TokenOutDenom: "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",
				},
			},
			TokenOutMinAmount: sdk.NewCoin("ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4", sdk.NewInt(217084399)),
		},
	}

	affiliateSwapMsgBz, err := json.Marshal(affiliateSwapMsg)
	suite.Require().NoError(err)

	// https://celatone.osmosis.zone/osmosis-1/txs/315EB6284778EBB5BAC0F94CC740F5D7E35DDA5BBE4EC9EC79F012548589C6E5
	executeMsg := &wasmtypes.MsgExecuteContract{
		Contract: "osmo1etpha3a65tds0hmn3wfjeag6wgxgrkuwg2zh94cf5hapz7mz04dq6c25s5",
		Sender:   "osmo1dldrxz5p8uezxz3qstpv92de7wgfp7hvr72dcm",
		Funds:    sdk.NewCoins(sdk.NewCoin("ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4", sdk.NewInt(217084399))),
		Msg:      affiliateSwapMsgBz,
	}

	_, isArb := txfee_filters.IsArbTxLooseAuthz(executeMsg, executeMsg.Funds[0].Denom, map[types.LiquidityChangeType]bool{})
	suite.Require().True(isArb)
}

func (suite *KeeperTestSuite) TestIsArbTxLooseAuthz_OtherMsg() {
	otherMsg := []byte(`{"update_feed": {}}`)

	// https://celatone.osmosis.zone/osmosis-1/txs/315EB6284778EBB5BAC0F94CC740F5D7E35DDA5BBE4EC9EC79F012548589C6E5
	executeMsg := &wasmtypes.MsgExecuteContract{
		Contract: "osmo1etpha3a65tds0hmn3wfjeag6wgxgrkuwg2zh94cf5hapz7mz04dq6c25s5",
		Sender:   "osmo1dldrxz5p8uezxz3qstpv92de7wgfp7hvr72dcm",
		Funds:    sdk.NewCoins(sdk.NewCoin("ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4", sdk.NewInt(217084399))),
		Msg:      otherMsg,
	}

	_, isArb := txfee_filters.IsArbTxLooseAuthz(executeMsg, executeMsg.Funds[0].Denom, map[types.LiquidityChangeType]bool{})
	suite.Require().False(isArb)
}
