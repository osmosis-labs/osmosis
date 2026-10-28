package v33

import (
	"github.com/osmosis-labs/osmosis/v32/app/upgrades"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v33 upgrade.
const UpgradeName = "v33"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}

// HUAHUA recovery, approved by governance in proposal 609
// (https://www.mintscan.io/osmosis/proposals/609).
//
// In January 2022 six external gauges were created to incentivize pools 605 (HUAHUA/OSMO) and
// 606 (HUAHUA/ATOM), but with the lock denoms "GAMM605" / "GAMM606" instead of
// "gamm/pool/605" / "gamm/pool/606". No lock can ever match these denoms, so the gauges never
// distributed and 6,000,000,000 HUAHUA remain held by the incentives module account.
const (
	// HuahuaDenom is uhuahua over transfer/channel-113.
	HuahuaDenom = "ibc/B9E0A1A524E98BB407D3CED8720EFEFD186002F90C1B1B7964811DD0CCC12228"

	// HuahuaRecoveryAddress is the Chihuahua ecosystem fund address on Osmosis.
	HuahuaRecoveryAddress = "osmo14fketv99hlrlk80mkggw643spsj3yyf7t2pjhr"
)

// HuahuaStuckGauges maps each stuck gauge ID to the (incorrect) lock denom it distributes to.
var HuahuaStuckGauges = []struct {
	GaugeID uint64
	Denom   string
}{
	{1954, "GAMM605"},
	{1955, "GAMM605"},
	{1956, "GAMM605"},
	{1957, "GAMM606"},
	{1958, "GAMM606"},
	{1959, "GAMM606"},
}

// Native WBTC to allBTC migration, approved by governance in proposal TODO(prop): fill in the
// proposal number and link before merging.
//
// The spendable native WBTC of every account that has a public key, apart from the WBTC issuer's
// merchants and custodian, is deposited into the allBTC alloy transmuter on its behalf (join_pool),
// which mints the same amount of allBTC back to the account. Accounts without a public key are
// skipped: they are protocol accounts (module accounts, pools, contracts, IBC escrows, CL spread
// reward and incentive addresses, ICA host accounts) whose WBTC balance is owed to someone else.
const (
	// MainnetChainID is the only chain the WBTC migration runs on.
	MainnetChainID = "osmosis-1"

	// WBTCDenom is native WBTC, issued on Osmosis through tokenfactory.
	WBTCDenom = "factory/" + WBTCIssuerAddress + "/wbtc"

	// WBTCIssuerAddress is the WBTC controller contract. Its merchants and custodian hold WBTC that
	// backs mints and burns, so they are not migrated.
	WBTCIssuerAddress = "osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743"

	// AllBTCTransmuterAddress is the alloy transmuter contract that mints allBTC.
	AllBTCTransmuterAddress = "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3"

	// AllBTCDenom is the alloyed denom minted by AllBTCTransmuterAddress.
	AllBTCDenom = "factory/" + AllBTCTransmuterAddress + "/alloyed/allBTC"

	// MaxWBTCHoldersToMigrate bounds the migration loop. If the chain has more WBTC holders than this,
	// the whole migration is skipped. osmosis-1 had 1,940 holders at height 73270661.
	MaxWBTCHoldersToMigrate = 5_000

	// maxWBTCIssuerMembers bounds the merchant list read from the WBTC issuer.
	maxWBTCIssuerMembers = 100
)

// Noble USDC to allUSDC migration, approved by governance in proposal TODO(prop): fill in the
// proposal number and link before merging.
//
// The spendable Noble USDC of every account that has a public key, apart from the accounts in
// nobleUSDCExcludedBech32, is deposited into the allUSDC alloy transmuter on its behalf (join_pool),
// which mints the same amount of allUSDC back to the account. Accounts without a public key are
// skipped, same as the WBTC migration. Noble USDC is an IBC voucher: there is no tokenfactory
// issuer, so there are no merchants or a custodian to exclude.
const (
	// NobleUSDCDenom is Noble USDC on Osmosis, an IBC voucher. It has no tokenfactory issuer.
	NobleUSDCDenom = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"

	// AllUSDCTransmuterAddress is the alloy transmuter contract that mints allUSDC (code 996,
	// cosmwasm pool 3497).
	AllUSDCTransmuterAddress = "osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0"

	// AllUSDCDenom is the alloyed denom minted by AllUSDCTransmuterAddress.
	AllUSDCDenom = "factory/" + AllUSDCTransmuterAddress + "/alloyed/allUSDC"

	// MaxNobleUSDCHoldersToMigrate bounds the migration loop. If the chain has more Noble USDC
	// holders than this, the whole migration is skipped. osmosis-1 listed 31,516 indexed holders
	// in October 2026. Revalidate against the final cmd/alloy-holder-scan before merge: indexed
	// plus unindexed holders must stay under this cap.
	MaxNobleUSDCHoldersToMigrate = 40_000
)

// nobleUSDCExcludedBech32 is the explicit list of accounts whose Noble USDC is not migrated.
// TODO(ops): exchanges, market makers, and bridge-solver wallets must be decided before merge.
var nobleUSDCExcludedBech32 = []string{}
