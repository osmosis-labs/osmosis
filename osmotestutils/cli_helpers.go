package osmotestutils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultFeeString(cfg network.Config) string {
	feeCoins := sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, sdk.NewInt(10)))
	return fmt.Sprintf("--%s=%s", flags.FlagFees, feeCoins.String())
}
