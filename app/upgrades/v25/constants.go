package v25

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v25 upgrade.
const (
	UpgradeName = "v25"

	// USDC is used as the auction denom
	AuctionUSDCDenom = "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{auctiontypes.StoreKey},
		Deleted: []string{},
	},
}

// AuctionParams expected initial params for the block-sdk
var AuctionParams = auctiontypes.Params{
	MaxBundleSize:          3,
	ReserveFee:             sdk.NewCoin(AuctionUSDCDenom, sdk.NewInt(1000000)),
	MinBidIncrement:        sdk.NewCoin(AuctionUSDCDenom, sdk.NewInt(1000000)),
	EscrowAccountAddress:   auctiontypes.DefaultEscrowAccountAddress,
	FrontRunningProtection: true,
	ProposerFee:            osmomath.MustNewDecFromStr("0.05"),
}
