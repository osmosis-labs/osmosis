package testutil

import (
	"fmt"

	claimcli "github.com/c-osmosis/osmosis/x/claim/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// commonArgs is args for CLI test commands
var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
}

// MsgClaim creates a claim message
func MsgClaim(clientCtx client.Context, sender fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {

	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, sender.String()),
	}

	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, claimcli.NewCmdClaim(), args)
}
