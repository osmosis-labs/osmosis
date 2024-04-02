package v25

import (
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v25 upgrade.
const (
	UpgradeName  = "v25"
	AuctionDenom = "uosmo"
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
	ReserveFee:             sdk.NewCoin(AuctionDenom, sdk.NewInt(1000000)),
	MinBidIncrement:        sdk.NewCoin(AuctionDenom, sdk.NewInt(1000000)),
	EscrowAccountAddress:   auctiontypes.DefaultEscrowAccountAddress,
	FrontRunningProtection: true,
	ProposerFee:            auctiontypes.DefaultProposerFee,
}
