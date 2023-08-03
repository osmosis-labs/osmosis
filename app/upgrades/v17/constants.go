package v17

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v17/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v17 upgrade.
const UpgradeName = "v17"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}

const (
	QuoteAsset  = "uosmo"
	TickSpacing = 100
)

type AssetPair struct {
	BaseAsset         string
	SpreadFactor      sdk.Dec
	LinkedClassicPool uint64
	Superfluid        bool
}

var AssetPairs = []AssetPair{
	{LinkedClassicPool: 837},
	{
		SpreadFactor:      sdk.MustNewDecFromStr("0.0005"), // Normally 0.0002, but is not authorized
		LinkedClassicPool: 857,
	},
	{LinkedClassicPool: 712},
	{LinkedClassicPool: 773},
	{LinkedClassicPool: 9},
	{LinkedClassicPool: 3},
	{LinkedClassicPool: 812},
	{LinkedClassicPool: 584},
	{LinkedClassicPool: 604},
	{LinkedClassicPool: 497},
	{LinkedClassicPool: 806},
	{LinkedClassicPool: 907},
	{LinkedClassicPool: 1013},
	{LinkedClassicPool: 15},
	{LinkedClassicPool: 586},
	{LinkedClassicPool: 627},
	{LinkedClassicPool: 795},
	{LinkedClassicPool: 730},
	{LinkedClassicPool: 7},
	{LinkedClassicPool: 1039},
	{LinkedClassicPool: 5},
	{LinkedClassicPool: 573},
	{LinkedClassicPool: 641},
	{LinkedClassicPool: 605},
	{LinkedClassicPool: 971},
	{LinkedClassicPool: 625},
}

// AssetPairs contract: all AssetPairs being initialized in this upgrade handler all have the same quote asset (OSMO).
func InitializeAssetPairs(ctx sdk.Context, keepers *keepers.AppKeepers) []AssetPair {
	gammKeeper := keepers.GAMMKeeper
	superfluidKeeper := keepers.SuperfluidKeeper
	for i, assetPair := range AssetPairs {
		pool, err := gammKeeper.GetCFMMPool(ctx, assetPair.LinkedClassicPool)
		if err != nil {
			panic(err)
		}

		// Set the base asset as the non-osmo asset in the pool
		poolLiquidity := pool.GetTotalPoolLiquidity(ctx)
		for _, coin := range poolLiquidity {
			if coin.Denom != QuoteAsset {
				AssetPairs[i].BaseAsset = coin.Denom
				break
			}
		}

		// If the spread factor is not manually set above, set it to the the same value as the pool's spread factor.
		if assetPair.SpreadFactor.IsNil() {
			AssetPairs[i].SpreadFactor = pool.GetSpreadFactor(ctx)
		}

		// Check if the pool is superfluid.
		// If the pool is superfluid, set the superfluid flag to true.
		poolShareDenom := fmt.Sprintf("gamm/pool/%d", assetPair.LinkedClassicPool)
		_, err = superfluidKeeper.GetSuperfluidAsset(ctx, poolShareDenom)
		if err != nil {
			continue
		}
		AssetPairs[i].Superfluid = true
	}
	return AssetPairs
}

// The values below this comment are used strictly for testing.
// The above code pulls desired values directly from the pool.
// For E2E / gotests, the pools we need don't exist already, so we need to hardcode the values here.
// These values will be pulled directly from the existing pools in the upgrade handler.

var (
	ION            = "uion"
	ISTIBCDenom    = "ibc/92BE0717F4678905E53F4E45B2DED18BC0CB97BF1F8B6A25AFEDF3D5A879B4D5"
	CMSTIBCDenom   = "ibc/23CA6C8D1AB2145DD13EB1E089A2E3F960DC298B468CCE034E19E5A78B61136E"
	WBTCIBCDenom   = "ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F"
	DOTIBCDenom    = "ibc/3FF92D26B407FD61AE95D975712A7C319CDE28DE4D80BDC9978D935932B991D7"
	CROIBCDenom    = "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1"
	AKTIBCDenom    = "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4"
	AXLIBCDenom    = "ibc/903A61A498756EA560B85A85132D3AEE21B5DEDD41213725D22ABF276EA6945E"
	SCRTIBCDenom   = "ibc/0954E1C28EB7AF5B72D24F3BC2B47BBB2FDF91BDDFD57B74B99E133AED40972A"
	STARSIBCDenom  = "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4"
	JUNOIBCDenom   = "ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED"
	STRDIBCDenom   = "ibc/A8CA5EE328FA10C9519DF6057DA1F69682D28F7D0F5CCC7ECB72E3DCA2D157A4"
	MARSIBCDenom   = "ibc/573FCD90FACEE750F55A8864EF7D38265F07E5A9273FA0E8DAFD39951332B580"
	XPRTIBCDenom   = "ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293"
	MEDIBCDenom    = "ibc/3BCCC93AD5DF58D11A6F8A05FA8BC801CBA0BA61A981F57E91B8B598BF8061CB"
	SOMMIBCDenom   = "ibc/9BBA9A1C257E971E38C1422780CE6F0B0686F0A3085E2D61118D904BFE0F5F5E"
	BLDIBCDenom    = "ibc/2DA9C149E9AD2BD27FEFA635458FB37093C256C1A940392634A16BEA45262604"
	KAVAIBCDenom   = "ibc/57AA1A70A4BC9769C525EBF6386F7A21536E04A79D62E1981EFCEF9428EBB205"
	IRISIBCDenom   = "ibc/7C4D60AA95E5A7558B0A364860979CA34B7FF8AAF255B87AF9E879374470CEC0"
	stIBCXDenom    = "factory/osmo1xqw2sl9zk8a6pch0csaw78n4swg5ws8t62wc5qta4gnjxfqg6v2qcs243k/stuibcx"
	DVPNIBCDenom   = "ibc/9712DBB13B9631EDFA9BF61B55F1B2D290B2ADB67E3A4EB3A875F3B6081B3B84"
	BTSGIBCDenom   = "ibc/4E5444C35610CC76FC94E7F7886B93121175C28262DDFDDE6F84E82BF2425452"
	UMEEIBCDenom   = "ibc/67795E528DF67C5606FC20F824EA39A6EF55BA133F4DC79C90A8C47A0901E17C"
	HUAHUAIBCDenom = "ibc/B9E0A1A524E98BB407D3CED8720EFEFD186002F90C1B1B7964811DD0CCC12228"
	NCTIBCDenom    = "ibc/A76EB6ECF4E3E2D4A23C526FD1B48FDD42F171B206C9D2758EF778A7826ADD68"
	GRAVIBCDenom   = "ibc/E97634A40119F1898989C2A23224ED83FDD0A57EA46B3A094E287288D1672B44"
)

var AssetPairsForTestsOnly = []AssetPair{
	{
		BaseAsset:         ISTIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 837,
		Superfluid:        true,
	},
	{
		BaseAsset:         CMSTIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.0005"), // Normally 0.0002, but is not authorized
		LinkedClassicPool: 857,
		Superfluid:        false,
	},
	{
		BaseAsset:         WBTCIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 712,
		Superfluid:        true,
	},
	{
		BaseAsset:         DOTIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 773,
		Superfluid:        true,
	},
	{
		BaseAsset:         CROIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 9,
		Superfluid:        true,
	},
	{
		BaseAsset:         AKTIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 3,
		Superfluid:        true,
	},
	{
		BaseAsset:         AXLIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 812,
		Superfluid:        true,
	},
	{
		BaseAsset:         SCRTIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 584,
		Superfluid:        true,
	},
	{
		BaseAsset:         STARSIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.003"),
		LinkedClassicPool: 604,
		Superfluid:        true,
	},
	{
		BaseAsset:         JUNOIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.003"),
		LinkedClassicPool: 497,
		Superfluid:        true,
	},
	{
		BaseAsset:         STRDIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 806,
		Superfluid:        true,
	},
	{
		BaseAsset:         MARSIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 907,
		Superfluid:        true,
	},
	{
		BaseAsset:         ION,
		SpreadFactor:      sdk.MustNewDecFromStr("0.005"),
		LinkedClassicPool: 1013,
		Superfluid:        true,
	},
	{
		BaseAsset:         XPRTIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 15,
		Superfluid:        true,
	},
	{
		BaseAsset:         MEDIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 586,
		Superfluid:        false,
	},
	{
		BaseAsset:         SOMMIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 627,
		Superfluid:        true,
	},
	{
		BaseAsset:         BLDIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 795,
		Superfluid:        true,
	},
	{
		BaseAsset:         KAVAIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 730,
		Superfluid:        true,
	},
	{
		BaseAsset:         IRISIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 7,
		Superfluid:        false,
	},
	{
		BaseAsset:         stIBCXDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.003"),
		LinkedClassicPool: 1039,
		Superfluid:        false,
	},
	{
		BaseAsset:         DVPNIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 5,
		Superfluid:        false,
	},
	{
		BaseAsset:         BTSGIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 573,
		Superfluid:        false,
	},
	{
		BaseAsset:         UMEEIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 641,
		Superfluid:        false,
	},
	{
		BaseAsset:         HUAHUAIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 605,
		Superfluid:        true,
	},
	{
		BaseAsset:         NCTIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 971,
		Superfluid:        false,
	},
	{
		BaseAsset:         GRAVIBCDenom,
		SpreadFactor:      sdk.MustNewDecFromStr("0.002"),
		LinkedClassicPool: 625,
		Superfluid:        false,
	},
}
