package v25

import (
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"

	store "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	smartaccounttypes "github.com/osmosis-labs/osmosis/v31/x/smart-account/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v25 upgrade.
const (
	UpgradeName = "v25"

	// Noble USDC is used as the auction denom
	AuctionUSDCDenom = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"

	NewMaxAgeNumBlocks = int64(1_000_000)      // 1.5s blocks * 1_000_000 = 1.5M seconds > 2 weeks
	NewMaxAgeDuration  = time.Second * 1209600 // 2 weeks

	// MaximumUnauthenticatedGas for smart account transactions to verify the fee payer
	MaximumUnauthenticatedGas = uint64(120_000)

	// IsSmartAccountActive is used for the smart account circuit breaker, smartaccounts are deactivated for v25
	IsSmartAccountActive = false

	// CircuitBreakerController is a DAODAO address, used only to deactivate the smart account module
	// https://daodao.zone/dao/osmo1wn58hxkv0869ua7qmz3gvek3sz773l89a778fjqvenl6anwuhgnq6ks7kl/home
	CircuitBreakerController = "osmo1wn58hxkv0869ua7qmz3gvek3sz773l89a778fjqvenl6anwuhgnq6ks7kl"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			auctiontypes.StoreKey,
			smartaccounttypes.StoreKey,
		},
		Deleted: []string{},
	},
}

// AuctionParams expected initial params for the block-sdk
var AuctionParams = auctiontypes.Params{
	MaxBundleSize:          5,
	ReserveFee:             sdk.NewCoin(AuctionUSDCDenom, osmomath.NewInt(1000000)),
	MinBidIncrement:        sdk.NewCoin(AuctionUSDCDenom, osmomath.NewInt(1000000)),
	EscrowAccountAddress:   auctiontypes.DefaultEscrowAccountAddress,
	FrontRunningProtection: true,
	ProposerFee:            osmomath.MustNewDecFromStr("0.05"),
}
