package v19

import (
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/types/migration"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v18 upgrade.
const UpgradeName = "v19"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}

var Records = []migration.BalancerToConcentratedPoolLink{
	// created at v16
	{BalancerPoolId: 674, ClPoolId: 1066}, // DAI
	// created at https://www.mintscan.io/osmosis/proposals/571
	{BalancerPoolId: 837, ClPoolId: 1088},  // IST
	{BalancerPoolId: 857, ClPoolId: 1089},  // CMST
	{BalancerPoolId: 712, ClPoolId: 1090},  // WBTC
	{BalancerPoolId: 773, ClPoolId: 1091},  // DOT
	{BalancerPoolId: 9, ClPoolId: 1092},    // CRO
	{BalancerPoolId: 3, ClPoolId: 1093},    // AKT
	{BalancerPoolId: 812, ClPoolId: 1094},  // AXL
	{BalancerPoolId: 584, ClPoolId: 1095},  // SCRT
	{BalancerPoolId: 604, ClPoolId: 1096},  // STARS
	{BalancerPoolId: 497, ClPoolId: 1097},  // JUNO
	{BalancerPoolId: 806, ClPoolId: 1098},  // STRD
	{BalancerPoolId: 907, ClPoolId: 1099},  // MARS
	{BalancerPoolId: 1013, ClPoolId: 1100}, // ION
	{BalancerPoolId: 15, ClPoolId: 1101},   // XPRT
	{BalancerPoolId: 586, ClPoolId: 1102},  // MED
	{BalancerPoolId: 627, ClPoolId: 1103},  // SOMM
	{BalancerPoolId: 795, ClPoolId: 1104},  // BLD
	{BalancerPoolId: 730, ClPoolId: 1105},  // KAVA
	{BalancerPoolId: 7, ClPoolId: 1106},    // IRIS
	{BalancerPoolId: 1039, ClPoolId: 1107}, // stIBCX
	{BalancerPoolId: 5, ClPoolId: 1108},    // DVPN
	{BalancerPoolId: 573, ClPoolId: 1109},  // BTSG
	{BalancerPoolId: 641, ClPoolId: 1110},  // UMEE
	{BalancerPoolId: 605, ClPoolId: 1111},  // HUAHUA
	{BalancerPoolId: 971, ClPoolId: 1112},  // NCT
	{BalancerPoolId: 625, ClPoolId: 1113},  // GRAV
	// created at https://www.mintscan.io/osmosis/proposals/597
	{BalancerPoolId: 678, ClPoolId: 1133}, // USDC
	{BalancerPoolId: 704, ClPoolId: 1134}, // WETH
	{BalancerPoolId: 1, ClPoolId: 1135},   // ATOM
	{BalancerPoolId: 803, ClPoolId: 1136}, // stATOM
}
