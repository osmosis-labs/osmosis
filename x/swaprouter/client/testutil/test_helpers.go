package testutil

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/osmosis-labs/osmosis/v12/app"
	swaproutercli "github.com/osmosis-labs/osmosis/v12/x/swaprouter/client/cli"
	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"

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
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
}

// MsgCreatePool broadcast a pool creation message.
func MsgCreatePool(
	t *testing.T,
	clientCtx client.Context,
	owner fmt.Stringer,
	tokenWeights string,
	initialDeposit string,
	swapFee string,
	exitFee string,
	futureGovernor string,
	extraArgs ...string,
) (testutil.BufferWriter, error) {
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
		`, swaproutercli.PoolFileWeights,
			tokenWeights,
			swaproutercli.PoolFileInitialDeposit,
			initialDeposit,
			swaproutercli.PoolFileSwapFee,
			swapFee,
			swaproutercli.PoolFileExitFee,
			exitFee,
			swaproutercli.PoolFileExitFee,
			exitFee,
		),
	)

	args = append(args,
		fmt.Sprintf("--%s=%s", swaproutercli.FlagPoolFile, jsonFile.Name()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner.String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 400000),
	)

	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, swaproutercli.NewCreatePoolCmd(), args)
}

// UpdateTxFeeDenom creates and modifies gamm genesis to pay fee with given denom.
func UpdateTxFeeDenom(cdc codec.Codec, denom string) map[string]json.RawMessage {
	// modification to pay fee with test bond denom "stake"
	genesisState := app.ModuleBasics.DefaultGenesis(cdc)
	swaprouterGen := swaproutertypes.DefaultGenesis()
	swaprouterGen.Params.PoolCreationFee = sdk.Coins{sdk.NewInt64Coin(denom, 1000000)}
	swaprouterGenJson := cdc.MustMarshalJSON(swaprouterGen)
	genesisState[swaproutertypes.ModuleName] = swaprouterGenJson
	return genesisState
}
