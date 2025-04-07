package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	treasurytypes "github.com/osmosis-labs/osmosis/v27/x/treasury/types"
	"golang.org/x/exp/slices"
)

var _ bankkeeper.Keeper = (*CustomKeeper)(nil)

type CustomKeeper struct {
	*bankkeeper.BaseKeeper
	ak AccountKeeper
}

func NewCustomKeeper(baseBankKeeper *bankkeeper.BaseKeeper, ak AccountKeeper) CustomKeeper {
	return CustomKeeper{
		BaseKeeper: baseBankKeeper,
		ak:         ak,
	}
}

func (k *CustomKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Burner) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName))
	}
	index := slices.IndexFunc(amt, func(coin sdk.Coin) bool {
		return coin.Denom == appparams.BaseCoinUnit
	})
	// instead of burning, we would like to send the coins to the reserve.
	if index >= 0 {
		nativeCoin := amt[index]

		err := k.SendCoinsFromModuleToModule(ctx, moduleName, treasurytypes.ModuleName, sdk.NewCoins(nativeCoin))
		if err != nil {
			return fmt.Errorf("failed to send coins to the reserve on burn: %w", err)
		}

		// proceed with base logic but exclude native coin
		amt = slices.Delete(amt, index, index+1)
	}
	return k.BaseKeeper.BurnCoins(ctx, moduleName, amt)
}
