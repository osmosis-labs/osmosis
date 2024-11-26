package testutil

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/keepers"
	poolmanagercli "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/cli"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// commonArgs is args for CLI test commands.
var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10))).String()),
}

// MsgCreatePool broadcast a pool creation message.
func MsgCreatePool(
	t *testing.T,
	clientCtx client.Context,
	owner fmt.Stringer,
	tokenWeights string,
	initialDeposit string,
	spreadFactor string,
	exitFee string,
	futureGovernor string,
	extraArgs ...string,
) (testutil.BufferWriter, error) {
	t.Helper()
	args := []string{}

	jsonFile := testutil.WriteToNewTempFile(t,
		fmt.Sprintf(`
		{
		  "%s": "%s",
		  "%s": "%s",
		  "%s": "%s",
		  "%s": "%s",
		  "%s": "%s"
		}
		`, poolmanagercli.PoolFileWeights,
			tokenWeights,
			poolmanagercli.PoolFileInitialDeposit,
			initialDeposit,
			poolmanagercli.PoolFileSwapFee,
			spreadFactor,
			poolmanagercli.PoolFileExitFee,
			exitFee,
			poolmanagercli.PoolFileExitFee,
			exitFee,
		),
	)

	args = append(args,
		fmt.Sprintf("--%s=%s", poolmanagercli.FlagPoolFile, jsonFile.Name()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner.String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 400000),
	)

	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, poolmanagercli.NewCreatePoolCmd(), args)
}

// UpdateTxFeeDenom creates and modifies gamm genesis to pay fee with given denom.
func UpdateTxFeeDenom(cdc codec.Codec, denom string) map[string]json.RawMessage {
	// modification to pay fee with test bond denom "stake"
	genesisState := keepers.AppModuleBasics.DefaultGenesis(cdc)
	poolmanagerGen := poolmanagertypes.DefaultGenesis()
	poolmanagerGen.Params.PoolCreationFee = sdk.Coins{sdk.NewInt64Coin(denom, 1000000)}
	poolmanagerGenJson := cdc.MustMarshalJSON(poolmanagerGen)
	genesisState[poolmanagertypes.ModuleName] = poolmanagerGenJson
	return genesisState
}
