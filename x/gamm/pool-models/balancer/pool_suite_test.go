package balancer_test

import (
	fmt "fmt"
	"math/rand"
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	v10 "github.com/osmosis-labs/osmosis/v7/app/upgrades/v10"
	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	// allowedErrRatio is the maximal multiplicative difference in either
	// direction (positive or negative) that we accept to tolerate in
	// unit tests for calcuating the number of shares to be returned by
	// joining a pool. The comparison is done between Wolfram estimates and our AMM logic.
	allowedErrRatio = "0.0000001"
	// doesNotExistDenom denom name assummed to be used in test cases where the provided
	// denom does not exist in pool
	doesNotExistDenom = "doesnotexist"
)

var (
	oneTrillion          = sdk.NewInt(1e12)
	defaultOsmoPoolAsset = balancer.PoolAsset{
		Token:  sdk.NewCoin("uosmo", oneTrillion),
		Weight: sdk.NewInt(100),
	}
	defaultAtomPoolAsset = balancer.PoolAsset{
		Token:  sdk.NewCoin("uatom", oneTrillion),
		Weight: sdk.NewInt(100),
	}
	oneTrillionEvenPoolAssets = []balancer.PoolAsset{
		defaultOsmoPoolAsset,
		defaultAtomPoolAsset,
	}
)

// calcJoinSharesTestCase defines a testcase for TestCalcSingleAssetJoin and
// TestCalcJoinPoolShares.
//
// CalcJoinPoolShares calls calcSingleAssetJoin. As a result, we can reuse
// the same test cases for unit testing both calcSingleAssetJoin and
//  CalcJoinPoolShares with only one tokensIn.
type calcJoinSharesTestCase struct {
	name         string
	swapFee      sdk.Dec
	poolAssets   []balancer.PoolAsset
	tokensIn     sdk.Coins
	expectShares sdk.Int
	expectLiq    sdk.Coins
	expectPanic  bool
	expErr       error
}

// see calcJoinSharesTestCase struct definition.
var calcSingleAssetJoinTestCases = []calcJoinSharesTestCase{
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 2_499_999_968_750 = 1e20 * (( 1 + (50,000 / 1e12))^0.5 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100000000000000000000*%28%281+%2B+%2850000%2F1000000000000%29%29%5E0.5+-+1%29
		// 	Simplified:  P_issued = 2,499,999,968,750
		name:         "single tokensIn - equal weights with zero swap fee",
		swapFee:      sdk.MustNewDecFromStr("0"),
		poolAssets:   oneTrillionEvenPoolAssets,
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(2_499_999_968_750),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
		//
		// 2_487_500_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1e12))^0.5 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
		// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
		// 	Simplified:  P_issued = 2_487_500_000_000
		name:         "single tokensIn - equal weights with 0.01 swap fee",
		swapFee:      sdk.MustNewDecFromStr("0.01"),
		poolAssets:   oneTrillionEvenPoolAssets,
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(2_487_500_000_000),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
		//
		// 1_262_500_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.99) / 1e12))^0.5 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 0.5 (equal weights)
		// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=%28100+*+10%5E18+%29*+%28%28+1+%2B+%2850%2C000+*+%281+-+%281+-+0.5%29+*+0.99%29+%2F+1000000000000%29%29%5E0.5+-+1%29
		// 	Simplified:  P_issued = 1_262_500_000_000
		name:         "single tokensIn - equal weights with 0.99 swap fee",
		swapFee:      sdk.MustNewDecFromStr("0.99"),
		poolAssets:   oneTrillionEvenPoolAssets,
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(1_262_500_000_000),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
		//
		// 321_875_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.25) * 0.99) / 1e12))^0.25 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 0.25 (asset A, uosmo, has weight 1/4 of uatom)
		// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=%28100+*+10%5E18+%29*+%28%28+1+%2B+%2850%2C000+*+%281+-+%281+-+0.25%29+*+0.99%29+%2F+1000000000000%29%29%5E0.25+-+1%29
		// 	Simplified:  P_issued = 321_875_000_000
		name:    "single tokensIn - unequal weights with 0.99 swap fee",
		swapFee: sdk.MustNewDecFromStr("0.99"),
		poolAssets: []balancer.PoolAsset{
			defaultOsmoPoolAsset,
			{
				Token:  sdk.NewInt64Coin("uatom", 1e12),
				Weight: sdk.NewInt(300),
			},
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(321_875_000_000),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 4_159_722_200_000 = 1e20 * (( 1 + (50,000 / 1e12))^0.83 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 500 / (500 + 100) approx = 0.83
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28100+%2B+500%29%29%29+*+0%29%2F1000000000000%29%29%5E%28500+%2F+%28100+%2B+500%29%29+-+1%29
		// 	Simplified:  P_issued = 4_159_722_200_000
		name:    "single asset - token in weight is greater than the other token, with zero swap fee",
		swapFee: sdk.ZeroDec(),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 1e12),
				Weight: sdk.NewInt(500),
			},
			defaultAtomPoolAsset,
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(4_166_666_649_306),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 4_159_722_200_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.83) * 0.01) / 1e12))^0.83 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1e12
		//	W_t = normalized weight of deposited asset in pool = 500 / (500 + 100) approx = 0.83
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28100+%2B+500%29%29%29+*+0.01%29%2F1000000000000%29%29%5E%28500+%2F+%28100+%2B+500%29%29+-+1%29
		// 	Simplified:  P_issued = 4_159_722_200_000
		name:    "single asset - token in weight is greater than the other token, with non-zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0.01"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 1e12),
				Weight: sdk.NewInt(500),
			},
			defaultAtomPoolAsset,
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(4_159_722_200_000),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 833_333_315_972 = 1e20 * (( 1 + (50,000 / 1e12))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 200 / (200 + 1000) approx = 0.167
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28200+%2F+%28200+%2B+1000%29%29%29+*+0%29%2F1000000000000%29%29%5E%28200+%2F+%28200+%2B+1000%29%29+-+1%29
		// 	Simplified:  P_issued = 833_333_315_972
		name:    "single asset - token in weight is smaller than the other token, with zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 1e12),
				Weight: sdk.NewInt(200),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1e12),
				Weight: sdk.NewInt(1000),
			},
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(833_333_315_972),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 819_444_430_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.167) * 0.02) / 1e12))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 200 / (200 + 1000) approx = 0.167
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28200+%2F+%28200+%2B+1000%29%29%29+*+0.02%29%2F1000000000000%29%29%5E%28200+%2F+%28200+%2B+1000%29%29+-+1%29
		// 	Simplified:  P_issued = 819_444_430_000
		name:    "single asset - token in weight is smaller than the other token, with non-zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0.02"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 1e12),
				Weight: sdk.NewInt(200),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1e12),
				Weight: sdk.NewInt(1000),
			},
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(819_444_430_000),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 9_775_731_930_496_140_648 = 1e20 * (( 1 + (117552 / 156_736))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 117552
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 156_736
		//	W_t = normalized weight of deposited asset in pool = 200 / (200 + 1000) approx = 0.167
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%28117552*%281+-+%281-%28200+%2F+%28200+%2B+1000%29%29%29+*+0%29%2F156736%29%29%5E%28200+%2F+%28200+%2B+1000%29%29+-+1%29
		// 	Simplified:  P_issued = 9_775_731_930_496_140_648
		name:    "single asset - tokenIn is large relative to liquidity, token in weight is smaller than the other token, with zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 156_736),
				Weight: sdk.NewInt(200),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1e12),
				Weight: sdk.NewInt(1000),
			},
		},
		// 156_736 * 3 / 4 = 117552
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", (156_736*3)/4)),
		expectShares: sdk.NewIntFromUint64(9_775_731_930_496_140_648),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 9_644_655_900_000_000_000 = 1e20 * (( 1 + (117552 * (1 - (1 - 0.167) * 0.02) / 156_736))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 200 / (200 + 1000) approx = 0.167
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28200+%2F+%28200+%2B+1000%29%29%29+*+0.02%29%2F1000000000000%29%29%5E%28200+%2F+%28200+%2B+1000%29%29+-+1%29
		// 	Simplified:  P_issued = 9_644_655_900_000_000_000
		name:    "single asset - tokenIn is large relative to liquidity, token in weight is smaller than the other token, with non-zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0.02"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 156_736),
				Weight: sdk.NewInt(200),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1e12),
				Weight: sdk.NewInt(1000),
			},
		},
		// 156_736 / 4 * 3 = 117552
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 156_736/4*3)),
		expectShares: sdk.NewIntFromUint64(9_644_655_900_000_000_000),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 6_504_099_261_800_144_638 = 1e20 * (( 1 + (499_000 / 500_000))^0.09 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 1e20
		//	A_t = amount of deposited asset = 195920
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 156_736
		//	W_t = normalized weight of deposited asset in pool = 100 / (100 + 1000) approx = 0.09
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%28499999*%281+-+%281-%28100+%2F+%28100+%2B+1000%29%29%29+*+0%29%2F500000%29%29%5E%28100+%2F+%28100+%2B+1000%29%29+-+1%29
		// 	Simplified:  P_issued = 6_504_099_261_800_144_638
		name:    "single asset - (almost 1 == tokenIn / liquidity ratio), token in weight is smaller than the other token, with zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 500_000),
				Weight: sdk.NewInt(100),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1e12),
				Weight: sdk.NewInt(1000),
			},
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 499_999)),
		expectShares: sdk.NewIntFromUint64(6_504_099_261_800_144_638),
	},
	{
		// Currently, our Pow approximation function does not work correctly when one tries
		// to add liquidity that is larger than the existing liquidity.
		// The ratio of tokenIn / existing liquidity that is larger than or equal to 1 causes a panic.
		// This has been deemed as acceptable since it causes code complexity to fix
		// & only affects UX in an edge case (user has to split up single asset joins)
		name:    "single asset - (exactly 1 == tokenIn / liquidity ratio - failure), token in weight is smaller than the other token, with zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 500_000),
				Weight: sdk.NewInt(100),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1e12),
				Weight: sdk.NewInt(1000),
			},
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 500_000)),
		expectShares: sdk.NewIntFromUint64(6_504_099_261_800_144_638),
		expectPanic:  true,
	},
	{
		name:         "tokenIn asset does not exist in pool",
		swapFee:      sdk.MustNewDecFromStr("0"),
		poolAssets:   oneTrillionEvenPoolAssets,
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin(doesNotExistDenom, 50_000)),
		expectShares: sdk.ZeroInt(),
		expErr:       sdkerrors.Wrapf(types.ErrDenomNotFoundInPool, fmt.Sprintf(balancer.ErrMsgFormatNoPoolAssetFound, doesNotExistDenom)),
	},
	{
		// Pool liquidity is changed by 1e-12 / 2
		// P_issued = 1e20 * 1e-12 / 2 = 1e8 / 2 = 50_000_000
		name:    "minimum input single asset equal liquidity",
		swapFee: sdk.MustNewDecFromStr("0"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
				Weight: sdk.NewInt(100),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
				Weight: sdk.NewInt(100),
			},
		},
		tokensIn: sdk.NewCoins(
			sdk.NewInt64Coin("uosmo", 1),
		),
		expectShares: sdk.NewInt(50_000_000),
	},
	{
		// P_issued should be 1/10th that of the previous test
		// p_issued = 50_000_000 / 10 = 5_000_000
		name:    "minimum input single asset imbalanced liquidity",
		swapFee: sdk.MustNewDecFromStr("0"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 10_000_000_000_000),
				Weight: sdk.NewInt(100),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
				Weight: sdk.NewInt(100),
			},
		},
		tokensIn: sdk.NewCoins(
			sdk.NewInt64Coin("uosmo", 1),
		),
		expectShares: sdk.NewInt(5_000_000),
	},
}

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
	// be post-bug
	suite.Ctx = suite.Ctx.WithBlockHeight(v10.ForkHeight)
}

// This test sets up 2 asset pools, and then checks the spot price on them.
// It uses the pools spot price method, rather than the Gamm keepers spot price method.
func (suite *KeeperTestSuite) TestBalancerSpotPrice() {
	baseDenom := "uosmo"
	quoteDenom := "uion"

	tests := []struct {
		name                string
		baseDenomPoolInput  sdk.Coin
		quoteDenomPoolInput sdk.Coin
		expectError         bool
		expectedOutput      sdk.Dec
	}{
		{
			name:                "equal value",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 100),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("1"),
		},
		{
			name:                "1:2 ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 200),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.500000000000000000"),
		},
		{
			name:                "2:1 ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 200),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 100),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("2.000000000000000000"),
		},
		{
			name:                "rounding after sigfig ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 220),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 115),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("1.913043480000000000"), // ans is 1.913043478260869565, rounded is 1.91304348
		},
		{
			name:                "check number of sig figs",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 300),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.333333330000000000"),
		},
		{
			name:                "check number of sig figs high sizes",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 343569192534),
			quoteDenomPoolInput: sdk.NewCoin(quoteDenom, sdk.MustNewDecFromStr("186633424395479094888742").TruncateInt()),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.000000000001840877"),
		},
	}

	for _, tc := range tests {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			poolId := suite.PrepareUni2PoolWithAssets(
				tc.baseDenomPoolInput,
				tc.quoteDenomPoolInput,
			)

			pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().NoError(err, "test: %s", tc.name)
			balancerPool, isPool := pool.(*balancer.Pool)
			suite.Require().True(isPool, "test: %s", tc.name)

			sut := func() {
				spotPrice, err := balancerPool.SpotPrice(
					suite.Ctx,
					tc.baseDenomPoolInput.Denom,
					tc.quoteDenomPoolInput.Denom)

				if tc.expectError {
					suite.Require().Error(err, "test: %s", tc.name)
				} else {
					suite.Require().NoError(err, "test: %s", tc.name)
					suite.Require().True(spotPrice.Equal(tc.expectedOutput),
						"test: %s\nSpot price wrong, got %s, expected %s\n", tc.name,
						spotPrice, tc.expectedOutput)
				}
			}
			assertPoolStateNotModified(suite.T(), balancerPool, sut)
		})
	}
}

func (suite *KeeperTestSuite) TestCalcJoinPoolShares() {
	// We append shared calcSingleAssetJoinTestCases with multi-asset and edge
	// test cases.
	//
	// See calcJoinSharesTestCase struct definition for explanation why the
	// sharing is needed.
	testCases := []calcJoinSharesTestCase{
		{
			name:       "swap equal weights with zero swap fee",
			swapFee:    sdk.MustNewDecFromStr("0"),
			poolAssets: oneTrillionEvenPoolAssets,
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 25_000),
			),
			// Raises liquidity perfectly by 25_000 / 1_000_000_000_000.
			// Initial number of pool shares = 100 * 10**18 = 10**20
			// Expected increase = liquidity_increase_ratio * initial number of pool shares = (25_000 / 1e12) * 10**20 = 2500000000000.0 = 2.5 * 10**12
			expectShares: sdk.NewInt(2.5e12),
		},
		{
			name:       "swap equal weights with 0.001 swap fee",
			swapFee:    sdk.MustNewDecFromStr("0.001"),
			poolAssets: oneTrillionEvenPoolAssets,
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 25_000),
			),
			expectShares: sdk.NewInt(2500000000000),
		},
		{
			// For uosmos and uatom
			// join pool is first done to the extent where the ratio can be preserved, which is 25,000 uosmo and 25,000 uatom
			// then we perfrom single asset deposit for the remaining 25,000 uatom with the equation below
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			// 1_249_999_960_937 = (1e20 + 2.5e12) * (( 1 + (25000 * 1 / 1000000025000))^0.5 - 1) (without fee)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 25,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,025,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full Solution without fees: https://www.wolframalpha.com/input?i=%28100+*+10%5E18+%2B+2.5e12+%29*+%28%28+1%2B+++++%2825000+*+%281%29+%2F+1000000025000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_500_000_000_000 + 1_249_999_960_937

			name:       "Multi-tokens In: unequal amounts, equal weights with 0 swap fee",
			swapFee:    sdk.ZeroDec(),
			poolAssets: oneTrillionEvenPoolAssets,
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 50_000),
			),

			expectShares: sdk.NewInt(2.5e12 + 1249999992187),
		},
		{
			// For uosmos and uatom
			// join pool is first done to the extent where the ratio can be preserved, which is 25,000 uosmo and 25,000 uatom
			// then we perfrom single asset deposit for the remaining 25,000 uatom with the equation below
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			// 1_243_750_000_000 = (1e20 + 2.5e12)*  (( 1 + (25000 * (1 - (1 - 0.5) * 0.01) / 1000000025000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 25,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,025,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution with fees: https://www.wolframalpha.com/input?i=%28100+*10%5E18%2B2.5e12%29*%28%281%2B+++%2825000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_500_000_000_000 + 1_243_750_000_000

			name:       "Multi-tokens In: unequal amounts, equal weights with 0.01 swap fee",
			swapFee:    sdk.MustNewDecFromStr("0.01"),
			poolAssets: oneTrillionEvenPoolAssets,
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 50_000),
			),

			expectShares: sdk.NewInt(2.5e12 + 1243750000000),
		},
		{
			// join pool is first done to the extent where the ratio can be preserved, which is 25,000 uosmo and 12,500 uatom.
			// the minimal total share resulted here would be 1,250,000,000,000 =  2500 / 2,000,000,000,000 * 100,000,000,000,000,000,000
			// then we perfrom single asset deposit for the remaining 37,500 uatom with the equation below
			//
			// For uatom:
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			// 609,374,990,000 = (1e20 + 1,250,000,000,000) *  (( 1 + (37,500 * (1 - (1 - 1/6) * 0.03) / 10,000,00,025,000))^1/6 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20 + 1_250_000_000_000 (from first join pool)
			//	A_t = amount of deposited asset = 37,500
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,025,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution with fees: https://www.wolframalpha.com/input?i=%28100+*10%5E18+%2B+1250000000000%29*%28%281%2B++++%2837500*%281+-+%281-1%2F6%29+*+0.03%29%2F1000000012500%29%29%5E%281%2F6%29+-+1%29
			// 	Simplified:  P_issued = 1,250,000,000,000 + 609,374,990,000
			name:    "Multi-tokens In: unequal amounts, with unequal weights with 0.03 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.03"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 2_000_000_000_000),
					Weight: sdk.NewInt(500),
				},
				defaultAtomPoolAsset,
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 50_000),
			),
			expectShares: sdk.NewInt(1250000000000 + 609374990000),
		},
		{
			// This test doubles the liquidity in a fresh pool, so it should generate the base number of LP shares for pool creation as new shares
			// This is set to 1e20 (or 100 * 10^18) for Osmosis, so we should expect:
			// P_issued = 1e20
			name:    "minimum input with two assets and minimum liquidity",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 1),
				sdk.NewInt64Coin("uatom", 1),
			),
			expectShares: sdk.NewInt(1e18).Mul(sdk.NewInt(100)),
		},
		{
			// Pool liquidity is changed by 1e-12
			// P_issued = 1e20 * 1e-12 = 1e8
			name:    "minimum input two assets equal liquidity",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 1),
				sdk.NewInt64Coin("uatom", 1),
			),
			expectShares: sdk.NewInt(100_000_000),
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.T().Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.ZeroDec(), tc.poolAssets...)

			// system under test
			sut := func() {
				shares, liquidity, err := pool.CalcJoinPoolShares(suite.Ctx, tc.tokensIn, tc.swapFee)
				if tc.expErr != nil {
					require.Error(t, err)
					require.ErrorAs(t, tc.expErr, &err)
					require.Equal(t, sdk.ZeroInt(), shares)
					require.Equal(t, sdk.NewCoins(), liquidity)
				} else {
					require.NoError(t, err)
					assertExpectedSharesErrRatio(t, tc.expectShares, shares)
					assertExpectedLiquidity(t, tc.tokensIn, liquidity)
				}
			}

			balancerPool, ok := pool.(*balancer.Pool)
			require.True(t, ok)

			assertPoolStateNotModified(t, balancerPool, func() {
				osmoutils.ConditionalPanic(t, tc.expectPanic, sut)
			})
		})
	}
}

// Tests selecting a random amount of coins to LP, and then that ExitPool(JoinPool(tokens))
// preserves the pools number of LP shares, and returns fewer coins to the acter than they started with.
func (suite *KeeperTestSuite) TestRandomizedJoinPoolExitPoolInvariants() {
	type testCase struct {
		initialTokensDenomIn  int64
		initialTokensDenomOut int64

		percentRatio int64

		numShares sdk.Int
	}

	const (
		denomOut = "denomOut"
		denomIn  = "denomIn"
	)

	now := int64(time.Now().Unix())
	rng := rand.NewSource(now)
	suite.T().Logf("Using random source of %d\n", now)

	// generate test case with randomized initial assets and join/exit ratio
	newCase := func() (tc *testCase) {
		tc = new(testCase)
		tc.initialTokensDenomIn = rng.Int63() % (1 << 62)
		tc.initialTokensDenomOut = rng.Int63() % (1 << 62)

		// 1%~100% of initial assets
		tc.percentRatio = rng.Int63()%100 + 1

		return tc
	}

	swapFeeDec := sdk.ZeroDec()
	exitFeeDec := sdk.ZeroDec()

	// create pool with randomized initial token amounts
	// and randomized ratio of join/exit
	createPool := func(tc *testCase) (pool *balancer.Pool) {
		poolAssetOut := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut),
			Weight: sdk.NewInt(5),
		}

		poolAssetIn := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn),
			Weight: sdk.NewInt(5),
		}

		pool = createTestPool(suite.T(), swapFeeDec, exitFeeDec, poolAssetOut, poolAssetIn).(*balancer.Pool)
		suite.Require().NotNil(pool)

		return pool
	}

	// joins with predetermined ratio
	joinPool := func(pool types.PoolI, tc *testCase) {
		tokensIn := sdk.Coins{
			sdk.NewCoin(denomIn, sdk.NewInt(tc.initialTokensDenomIn).MulRaw(tc.percentRatio).QuoRaw(100)),
			sdk.NewCoin(denomOut, sdk.NewInt(tc.initialTokensDenomOut).MulRaw(tc.percentRatio).QuoRaw(100)),
		}
		numShares, err := pool.JoinPool(suite.Ctx, tokensIn, swapFeeDec)
		suite.Require().NoError(err)
		tc.numShares = numShares
	}

	// exits for same amount of shares minted
	exitPool := func(pool types.PoolI, tc *testCase) {
		_, err := pool.ExitPool(suite.Ctx, tc.numShares, exitFeeDec)
		suite.Require().NoError(err)
	}

	invariantJoinExitInversePreserve := func(
		beforeCoins, afterCoins sdk.Coins,
		beforeShares, afterShares sdk.Int,
	) {
		// test token amount has been preserved
		suite.Require().True(
			!beforeCoins.IsAnyGT(afterCoins),
			"Coins has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeCoins, afterCoins,
		)
		// test share amount has been preserved
		suite.Require().True(
			beforeShares.Equal(afterShares),
			"Shares has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeShares, afterShares,
		)
	}

	testPoolInvariants := func() {
		tc := newCase()
		pool := createPool(tc)
		originalCoins, originalShares := pool.GetTotalPoolLiquidity(sdk.Context{}), pool.GetTotalShares()
		joinPool(pool, tc)
		exitPool(pool, tc)
		invariantJoinExitInversePreserve(
			originalCoins, pool.GetTotalPoolLiquidity(sdk.Context{}),
			originalShares, pool.GetTotalShares(),
		)
	}

	for i := 0; i < 50000; i++ {
		testPoolInvariants()
	}
}
