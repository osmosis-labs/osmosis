package simtypes

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	tokenfactorytypes "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/types"
)

type App interface {
	GetBaseApp() *baseapp.BaseApp
	AppCodec() codec.Codec
	GetAccountKeeper() AccountKeeper
	GetBankKeeper() BankKeeper
	GetTokenFactoryKeeper() TokenFactoryKeeper
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	// GetAllAccounts(ctx sdk.Context) []authtypes.AccountI
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// TODO: Revisit
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

type TokenFactoryKeeper interface {
	GetParams(ctx sdk.Context) (params tokenfactorytypes.Params)
}
