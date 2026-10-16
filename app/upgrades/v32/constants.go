package v32

import (
	"github.com/osmosis-labs/osmosis/v32/app/upgrades"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v32 upgrade.
const UpgradeName = "v32"

// LiquiditySubDAOAddress is the DAO DAO core contract of the liquidity sub-DAO.
// Proposal 1044 allBTC and non-base-denom protorev balances are sent here so the
// sub-DAO can swap them to OSMO and burn it.
// https://daodao.zone/dao/osmo1rvq5cq2j35k7sqqz49e5e8zezl45fcywcawazh46qnc0g96d0d6sasqsgc
const LiquiditySubDAOAddress = "osmo1rvq5cq2j35k7sqqz49e5e8zezl45fcywcawazh46qnc0g96d0d6sasqsgc"

// MainnetChainID is the only chain the proposal 1044 seizure runs on. Other chains,
// including the e2e and testnet chains, do not hold the exploiter's allBTC.
const MainnetChainID = "osmosis-1"

// Proposal 1044 seizes the allBTC frozen on the exploiter by v31.1.0 and sends
// it to the Liquidity subDAO. The amount is the base-unit balance queried on
// 2026-09-24 (22.65060846 allBTC). The proposal text rounds this to 22.650608.
// The account's only other coin is 703055696 uosmo, which this upgrade leaves
// in place.
const (
	ExploiterAddress              = "osmo1wq76r2mhqsa9yaygghuwyq4wy6dcsgf8vtzltn"
	AllBTCDenom                   = "factory/osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3/alloyed/allBTC"
	ExpectedExploiterAllBTCAmount = "2265060846"
)

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
