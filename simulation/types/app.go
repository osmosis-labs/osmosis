package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type App interface {
	GetBaseApp() *baseapp.BaseApp
	// The genesis state of the blockchain is represented here as a map of raw json
	// messages key'd by a identifier string.
	// The identifier is used to determine which module genesis information belongs
	// to so it may be appropriately routed during init chain.
	// Within this application default genesis information is retrieved from
	// the ModuleBasicManager which populates json from each BasicModule
	// object provided to it during init.
	// NewDefaultGenesisState() map[string]json.RawMessage
	AppCodec() codec.Codec
	GetAccountKeeper() AccountKeeper
	GetBankKeeper() BankKeeper
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetAllAccounts(ctx sdk.Context) []authtypes.AccountI
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// TODO: Revisit
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
