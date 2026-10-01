package v32

import (
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v32 upgrade.
const UpgradeName = "v32"

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
