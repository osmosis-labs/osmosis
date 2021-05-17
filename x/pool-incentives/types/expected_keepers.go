package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
	incentivestypes "github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin

	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
}

type GAMMKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (gammtypes.PoolI, error)
}

type IncentivesKeeper interface {
	CreatePot(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
	GetPotByID(ctx sdk.Context, potID uint64) (*incentivestypes.Pot, error)

	AddToPotRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, potID uint64) error
}

type DistrKeeper interface {
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
}
