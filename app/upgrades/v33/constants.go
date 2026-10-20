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
// Every user account's spendable native WBTC is deposited into the allBTC alloy transmuter on its
// behalf (join_pool), which mints the same normalized amount of allBTC back to the account. Module
// accounts (including GAMM and CL pools) and wasm contracts (including cosmwasm pools) are skipped,
// since they track their own balances.
const (
	// MainnetChainID is the only chain the WBTC migration runs on.
	MainnetChainID = "osmosis-1"

	// WBTCDenom is native WBTC, issued on Osmosis through tokenfactory.
	WBTCDenom = "factory/osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743/wbtc"

	// AllBTCTransmuterAddress is the alloy transmuter contract that mints allBTC.
	AllBTCTransmuterAddress = "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3"

	// AllBTCDenom is the alloyed denom minted by AllBTCTransmuterAddress.
	AllBTCDenom = "factory/" + AllBTCTransmuterAddress + "/alloyed/allBTC"

	// MaxWBTCHoldersToMigrate bounds the migration loop. If the chain has more WBTC holders than this,
	// the whole migration is skipped. Confirm against the mainnet DenomOwners count before release.
	MaxWBTCHoldersToMigrate = 100_000
)
