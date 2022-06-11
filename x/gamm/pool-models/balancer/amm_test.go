package balancer_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v9/osmoutils"
	"github.com/osmosis-labs/osmosis/v9/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v9/x/gamm/types"
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
	oneTrillion               = sdk.NewInt(1e12)
	oneTrillionEvenPoolAssets = []balancer.PoolAsset{
		{
			Token:  sdk.NewCoin("uosmo", oneTrillion),
			Weight: sdk.NewInt(100),
		},
		{
			Token:  sdk.NewCoin("uatom", oneTrillion),
			Weight: sdk.NewInt(100),
		},
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
		// 2_499_999_968_750 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.5 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
		// 2_487_500_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1_000_000_000_000))^0.5 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
		// 1_262_500_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.99) / 1_000_000_000_000))^0.5 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
		// 321_875_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.25) * 0.99) / 1_000_000_000_000))^0.25 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
			{
				Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
				Weight: sdk.NewInt(100),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
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
		// 4_159_722_200_000 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.83 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 500 / (500 + 100) approx = 0.83
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28100+%2B+500%29%29%29+*+0%29%2F1000000000000%29%29%5E%28500+%2F+%28100+%2B+500%29%29+-+1%29
		// 	Simplified:  P_issued = 4_159_722_200_000
		name:    "single asset - token in weight is greater than the other token, with zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
				Weight: sdk.NewInt(500),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
				Weight: sdk.NewInt(100),
			},
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(4_166_666_649_306),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 4_159_722_200_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.83) * 0.01) / 1_000_000_000_000))^0.83 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
		//	A_t = amount of deposited asset = 50,000
		//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
		//	W_t = normalized weight of deposited asset in pool = 500 / (500 + 100) approx = 0.83
		// Plugging all of this in, we get:
		// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28100+%2B+500%29%29%29+*+0.01%29%2F1000000000000%29%29%5E%28500+%2F+%28100+%2B+500%29%29+-+1%29
		// 	Simplified:  P_issued = 4_159_722_200_000
		name:    "single asset - token in weight is greater than the other token, with non-zero swap fee",
		swapFee: sdk.MustNewDecFromStr("0.01"),
		poolAssets: []balancer.PoolAsset{
			{
				Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
				Weight: sdk.NewInt(500),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
				Weight: sdk.NewInt(100),
			},
		},
		tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		expectShares: sdk.NewInt(4_159_722_200_000),
	},
	{
		// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
		// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
		//
		// 833_333_315_972 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
				Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
				Weight: sdk.NewInt(200),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
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
		// 819_444_430_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.167) * 0.02) / 1_000_000_000_000))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
				Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
				Weight: sdk.NewInt(200),
			},
			{
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
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
		// 9_775_731_930_496_140_648 = 100 * 10^18 * (( 1 + (117552 / 156_736))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
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
		// 9_644_655_900_000_000_000 = 100 * 10^18 * (( 1 + (117552 * (1 - (1 - 0.167) * 0.02) / 156_736))^0.167 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
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
		// 6_504_099_261_800_144_638 = 100 * 10^18 * (( 1 + (499_000 / 500_000))^0.09 - 1)
		//
		// where:
		// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
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
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
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
				Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
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

		poolId := suite.PrepareUni2PoolWithAssets(
			tc.baseDenomPoolInput,
			tc.quoteDenomPoolInput,
		)

		pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
		suite.Require().NoError(err, "test: %s", tc.name)
		balancerPool, isPool := pool.(*balancer.Pool)
		suite.Require().True(isPool, "test: %s", tc.name)

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
}

// TestCalculateAmountOutAndIn_InverseRelationship tests that the same amount of token is guaranteed upon
// sequential operation of CalcInAmtGivenOut and CalcOutAmtGivenIn.
func TestCalculateAmountOutAndIn_InverseRelationship(t *testing.T) {
	type testcase struct {
		denomOut         string
		initialPoolOut   int64
		initialWeightOut int64
		initialCalcOut   int64

		denomIn         string
		initialPoolIn   int64
		initialWeightIn int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000_000_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 200,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 50,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000,
			initialWeightOut: 200,
			initialCalcOut:   100000,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000,
			initialWeightIn: 50,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("tokenOutInitial: %d, tokenInInitial: %d, initialOut: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.initialPoolIn,
			tc.initialCalcOut,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				ctx := createTestContext(t)

				poolAssetOut := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut),
					Weight: sdk.NewInt(tc.initialWeightOut),
				}

				poolAssetIn := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn),
					Weight: sdk.NewInt(tc.initialWeightIn),
				}

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				require.NoError(t, err)

				pool := createTestPool(t, swapFeeDec, exitFeeDec, poolAssetOut, poolAssetIn)
				require.NotNil(t, pool)

				initialOut := sdk.NewInt64Coin(poolAssetOut.Token.Denom, tc.initialCalcOut)
				initialOutCoins := sdk.NewCoins(initialOut)

				actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, poolAssetIn.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(actualTokenIn), poolAssetOut.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				require.Equal(t, initialOut.Denom, inverseTokenOut.Denom)

				expected := initialOut.Amount.ToDec()
				actual := inverseTokenOut.Amount.ToDec()

				// allow a rounding error of up to 1 for this relation
				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, expected, actual, tol))
			})
		}
	}
}

func TestCalcSingleAssetInAndOut_InverseRelationship(t *testing.T) {
	type testcase struct {
		initialPoolOut   int64
		initialPoolIn    int64
		initialWeightOut int64
		tokenOut         int64
		initialWeightIn  int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 50,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         50,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  50,
		},
		{
			initialPoolOut:   1_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   2_351_333,
			tokenOut:         7,
			initialWeightOut: 148,
			initialWeightIn:  57,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         25,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         26,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("initialPoolOut: %d, initialCalcOut: %d, initialWeightOut: %d, initialWeightIn: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.tokenOut,
			tc.initialWeightOut,
			tc.initialWeightIn,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				initialPoolBalanceOut := sdk.NewInt(tc.initialPoolOut)

				initialWeightOut := sdk.NewInt(tc.initialWeightOut)
				initialWeightIn := sdk.NewInt(tc.initialWeightIn)

				initialTotalShares := types.InitPoolSharesSupply.ToDec()
				initialCalcTokenOut := sdk.NewInt(tc.tokenOut)

				actualSharesOut := balancer.CalcPoolSharesOutGivenSingleAssetIn(
					initialPoolBalanceOut.ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares,
					initialCalcTokenOut.ToDec(),
					swapFeeDec,
				)

				inverseCalcTokenOut := balancer.CalcSingleAssetInGivenPoolSharesOut(
					initialPoolBalanceOut.Add(initialCalcTokenOut).ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares.Add(actualSharesOut),
					actualSharesOut,
					swapFeeDec,
				)

				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, initialCalcTokenOut.ToDec(), inverseCalcTokenOut, tol))
			})
		}
	}
}

func TestCalcJoinPoolShares(t *testing.T) {
	oneTrillionEvenPoolAssets := []balancer.PoolAsset{
		{
			Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
			Weight: sdk.NewInt(100),
		},
		{
			Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
			Weight: sdk.NewInt(100),
		},
	}
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
			// Expected increase = liquidity_increase_ratio * initial number of pool shares = (25_000 / 1_000_000_000_000) * 10**20 = 2500000000000.0 = 2.5 * 10**12
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
			// 1_249_999_960_937 = 100 * 10^18 * (( 1 + (25000 * 1 / 1000000025000))^0.5 - 1) (without fee)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 25,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,025,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full Solution without fees: https://www.wolframalpha.com/input?i=100+*+10%5E18+*+%28%28+1+%2B+++%2825000+*+%281%29+%2F+1000000025000%29%29%5E0.5+-+1%29
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
			// 1_243_750_000_000 = 100 * 10^18 *  (( 1 + (25000 * (1 - (1 - 0.5) * 0.01) / 1000000025000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 25,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,025,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution with fees: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2825000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_500_000_000_000 + 1_243_750_000_000

			name:       "Multi-tokens In: unequal amounts, equal weights with 0.01 swap fee",
			swapFee:    sdk.MustNewDecFromStr("0.01"),
			poolAssets: oneTrillionEvenPoolAssets,
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 50_000),
			),
			expectShares: sdk.NewInt(2.5e12 + 1243749900000),
		},
		{
			// join pool is first done to the extent where the ratio can be preserved, which is 25,000 uosmo and 12,500 uatom.
			// the minimal total share resulted here would be 1,250,000,000,000 =  2500 / 2,000,000,000,000 * 100,000,000,000,000,000,000
			// then we perfrom single asset deposit for the remaining 37,500 uatom with the equation below
			//
			// For uatom:
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			// 609,374,990,000 = (100 * 10^18 + 1,250,000,000,000) *  (( 1 + (37,500 * (1 - (1 - 1/6) * 0.03) / 10,000,00,025,000))^1/6 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 + 1_250_000_000_000 (from first join pool)
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
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 50_000),
			),
			expectShares: sdk.NewInt(1250000000000 + 609374990000),
		},
	}
	testCases = append(testCases, calcSingleAssetJoinTestCases...)

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

			// system under test
			sut := func() {
				shares, liquidity, err := pool.CalcJoinPoolShares(sdk.Context{}, tc.tokensIn, tc.swapFee)
				if tc.expErr != nil {
					require.Error(t, err)
					require.ErrorAs(t, tc.expErr, &err)
					require.Equal(t, sdk.ZeroInt(), shares)
					require.Equal(t, sdk.NewCoins(), liquidity)
				} else {
					require.NoError(t, err)
					assertExpectedSharesErrRatio(t, tc.expectShares, shares)
					assertExpectedLiquidity(t, tc.expectLiq, tc.tokensIn, liquidity)
				}
			}

			if tc.expectPanic {
				require.Panics(t, sut)
			} else {
				require.NotPanics(t, sut)
			}
		})
	}
}

// TestUpdateIntermediaryPoolAssetsLiquidity tests if `updateIntermediaryPoolAssetsLiquidity` returns poolAssetsByDenom map
// with the updated liquidity given by the parameter
func TestUpdateIntermediaryPoolAssetsLiquidity(t *testing.T) {
	testCases := []struct {
		name string

		// returns newLiquidity, originalPoolAssetsByDenom, expectedPoolAssetsByDenom
		setup func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset)

		err error
	}{
		{
			name: "regular case with multiple pool assets and a subset of newLiquidity to update",

			setup: func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset) {
				const (
					uosmoValueOriginal = 1_000_000_000_000
					atomValueOriginal  = 123
					ionValueOriginal   = 657

					// Weight does not affect calculations so it is shared
					weight = 100
				)

				newLiquidity := sdk.NewCoins(
					sdk.NewInt64Coin("uosmo", 1_000),
					sdk.NewInt64Coin("atom", 2_000),
					sdk.NewInt64Coin("ion", 3_000))

				originalPoolAssetsByDenom := map[string]balancer.PoolAsset{
					"uosmo": {
						Token:  sdk.NewInt64Coin("uosmo", uosmoValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"atom": {
						Token:  sdk.NewInt64Coin("atom", atomValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"ion": {
						Token:  sdk.NewInt64Coin("ion", ionValueOriginal),
						Weight: sdk.NewInt(weight),
					},
				}

				expectedPoolAssetsByDenom := map[string]balancer.PoolAsset{}
				for k, v := range originalPoolAssetsByDenom {
					expectedValue := balancer.PoolAsset{Token: v.Token, Weight: v.Weight}
					expectedValue.Token.Amount = expectedValue.Token.Amount.Add(newLiquidity.AmountOf(k))
					expectedPoolAssetsByDenom[k] = expectedValue
				}

				return newLiquidity, originalPoolAssetsByDenom, expectedPoolAssetsByDenom
			},
		},
		{
			name: "new liquidity has no coins",

			setup: func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset) {
				const (
					uosmoValueOriginal = 1_000_000_000_000
					atomValueOriginal  = 123
					ionValueOriginal   = 657

					// Weight does not affect calculations so it is shared
					weight = 100
				)

				newLiquidity := sdk.NewCoins()

				originalPoolAssetsByDenom := map[string]balancer.PoolAsset{
					"uosmo": {
						Token:  sdk.NewInt64Coin("uosmo", uosmoValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"atom": {
						Token:  sdk.NewInt64Coin("atom", atomValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"ion": {
						Token:  sdk.NewInt64Coin("ion", ionValueOriginal),
						Weight: sdk.NewInt(weight),
					},
				}

				return newLiquidity, originalPoolAssetsByDenom, originalPoolAssetsByDenom
			},
		},
		{
			name: "newLiquidity has a coin that poolAssets don't",

			setup: func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset) {
				const (
					uosmoValueOriginal = 1_000_000_000_000

					// Weight does not affect calculations so it is shared
					weight = 100
				)

				newLiquidity := sdk.NewCoins(
					sdk.NewInt64Coin("juno", 1_000))

				originalPoolAssetsByDenom := map[string]balancer.PoolAsset{
					"uosmo": {
						Token:  sdk.NewInt64Coin("uosmo", uosmoValueOriginal),
						Weight: sdk.NewInt(weight),
					},
				}

				return newLiquidity, originalPoolAssetsByDenom, originalPoolAssetsByDenom
			},

			err: fmt.Errorf(balancer.ErrMsgFormatFailedInterimLiquidityUpdate, "juno"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newLiquidity, originalPoolAssetsByDenom, expectedPoolAssetsByDenom := tc.setup()

			err := balancer.UpdateIntermediaryPoolAssetsLiquidity(newLiquidity, originalPoolAssetsByDenom)

			require.Equal(t, tc.err, err)

			if tc.err != nil {
				return
			}

			require.Equal(t, expectedPoolAssetsByDenom, originalPoolAssetsByDenom)
		})
	}
}

func TestCalcSingleAssetJoin(t *testing.T) {
	for _, tc := range calcSingleAssetJoinTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

			balancerPool, ok := pool.(*balancer.Pool)
			require.True(t, ok)

			tokenIn := tc.tokensIn[0]

			poolAssetInDenom := tokenIn.Denom
			// when testing a case with tokenIn that does not exist in pool, we just want
			// to provide any pool asset.
			if tc.expErr != nil && errors.Is(tc.expErr, types.ErrDenomNotFoundInPool) {
				poolAssetInDenom = tc.poolAssets[0].Token.Denom
			}

			// find pool asset in pool
			// must be in pool since weights get scaled in Balancer pool
			// constructor
			poolAssetIn, err := balancerPool.GetPoolAsset(poolAssetInDenom)
			require.NoError(t, err)

			// system under test
			sut := func() {
				shares, err := balancerPool.CalcSingleAssetJoin(tokenIn, tc.swapFee, poolAssetIn, pool.GetTotalShares())

				if tc.expErr != nil {
					require.Error(t, err)
					require.ErrorAs(t, tc.expErr, &err)
					require.Equal(t, sdk.ZeroInt(), shares)
					return
				}

				require.NoError(t, err)
				assertExpectedSharesErrRatio(t, tc.expectShares, shares)
			}

			if tc.expectPanic {
				require.Panics(t, sut)
			} else {
				require.NotPanics(t, sut)
			}
		})
	}
}

func TestCalcJoinSingleAssetTokensIn(t *testing.T) {
	testCases := []struct {
		name           string
		swapFee        sdk.Dec
		poolAssets     []balancer.PoolAsset
		tokensIn       sdk.Coins
		expectShares   sdk.Int
		expectLiqudity sdk.Coins
		expErr         error
	}{
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 2_499_999_968_750 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100000000000000000000*%28%281+%2B+%2850000%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2,499,999,968,750
			name:         "one token in - equal weights with zero swap fee",
			swapFee:      sdk.MustNewDecFromStr("0"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(2_499_999_968_750),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 2_499_999_968_750 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100000000000000000000*%28%281+%2B+%2850000%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2,499,999,968,750
			name:         "two tokens in - equal weights with zero swap fee",
			swapFee:      sdk.MustNewDecFromStr("0"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 50_000)),
			expectShares: sdk.NewInt(2_499_999_968_750 * 2),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_487_500_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1_000_000_000_000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_487_500_000_000
			name:         "one token in - equal weights with swap fee of 0.01",
			swapFee:      sdk.MustNewDecFromStr("0.01"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(2_487_500_000_000),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_487_500_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1_000_000_000_000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_487_500_000_000
			name:         "two tokens in - equal weights with swap fee of 0.01",
			swapFee:      sdk.MustNewDecFromStr("0.01"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 50_000)),
			expectShares: sdk.NewInt(2_487_500_000_000 * 2),
		},
		{
			// For uosmo:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_072_912_400_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.83) * 0.03) / 2_000_000_000))^0.83 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 2_000_000_000
			//	W_t = normalized weight of deposited asset in pool = 500 / 500 + 100 = 0.83
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F2000000000%29%29%5E%28500+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 2_072_912_400_000_000
			//
			//
			// For uatom:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 1_624_999_900_000 = 100 * 10^18 * (( 1 + (100_000 * (1 - (1 - 0.167) * 0.03) / 1_000_000_000_000))^0.167 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 100 / 500 + 100 = 0.167
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%28100000*%281+-+%281-%28100+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F1000000000000%29%29%5E%28100+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 1_624_999_900_000
			name:    "two varying tokens in, varying weights, with swap fee of 0.03",
			swapFee: sdk.MustNewDecFromStr("0.03"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 2_000_000_000),
					Weight: sdk.NewInt(500),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 100_000)),
			expectShares: sdk.NewInt(2_072_912_400_000_000 + 1_624_999_900_000),
		},
		{
			name:    "no tokens in",
			swapFee: sdk.MustNewDecFromStr("0.03"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 2_000_000_000),
					Weight: sdk.NewInt(500),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(),
			expectShares: sdk.NewInt(0),
		},
		{
			name:       "one of the tokensIn asset does not exist in pool",
			swapFee:    sdk.MustNewDecFromStr("0"),
			poolAssets: oneTrillionEvenPoolAssets,
			// Second tokenIn does not exist.
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin(doesNotExistDenom, 50_000)),
			expectShares: sdk.ZeroInt(),
			expErr:       fmt.Errorf(balancer.ErrMsgFormatNoPoolAssetFound, doesNotExistDenom),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

			balancerPool, ok := pool.(*balancer.Pool)
			require.True(t, ok)

			poolAssetsByDenom, err := balancer.GetPoolAssetsByDenom(balancerPool.GetAllPoolAssets())
			require.NoError(t, err)

			// estimate expected liquidity
			expectedNewLiquidity := sdk.NewCoins()
			for _, tokenIn := range tc.tokensIn {
				expectedNewLiquidity = expectedNewLiquidity.Add(tokenIn)
			}

			totalNumShares, totalNewLiquidity, err := balancerPool.CalcJoinSingleAssetTokensIn(tc.tokensIn, pool.GetTotalShares(), poolAssetsByDenom, tc.swapFee)

			if tc.expErr != nil {
				require.Error(t, err)
				require.ErrorAs(t, tc.expErr, &err)
				require.Equal(t, sdk.ZeroInt(), totalNumShares)
				require.Equal(t, sdk.Coins{}, totalNewLiquidity)
				return
			}

			require.NoError(t, err)

			require.Equal(t, expectedNewLiquidity, totalNewLiquidity)

			if tc.expectShares.Int64() == 0 {
				require.Equal(t, tc.expectShares, totalNumShares)
				return
			}

			assertExpectedSharesErrRatio(t, tc.expectShares, totalNumShares)
		})
	}
}

func TestRandomizedJoinPoolExitPoolInvariants(t *testing.T) {
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

	now := time.Now().Unix()
	rng := rand.NewSource(now)
	t.Logf("Using random source of %d\n", now)

	// generate test case with randomized initial assets and join/exit ratio
	newCase := func() (tc *testCase) {
		tc = new(testCase)
		tc.initialTokensDenomIn = rng.Int63() % 1_000_000
		tc.initialTokensDenomOut = rng.Int63() % 1_000_000

		// 1%~100% of initial assets
		tc.percentRatio = rng.Int63()%100 + 1

		return tc
	}

	swapFeeDec, err := sdk.NewDecFromStr("0")
	require.NoError(t, err)

	exitFeeDec, err := sdk.NewDecFromStr("0")
	require.NoError(t, err)

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

		pool = createTestPool(t, swapFeeDec, exitFeeDec, poolAssetOut, poolAssetIn).(*balancer.Pool)
		require.NotNil(t, pool)

		return pool
	}

	// joins with predetermined ratio
	joinPool := func(pool types.PoolI, tc *testCase) {
		tokensIn := sdk.Coins{
			sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn*tc.percentRatio/100),
			sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut*tc.percentRatio/100),
		}
		numShares, err := pool.JoinPool(sdk.Context{}, tokensIn, swapFeeDec)
		require.NoError(t, err)
		tc.numShares = numShares
	}

	// exits for same amount of shares minted
	exitPool := func(pool types.PoolI, tc *testCase) {
		_, err := pool.ExitPool(sdk.Context{}, tc.numShares, exitFeeDec)
		require.NoError(t, err)
	}

	invariantJoinExitInversePreserve := func(
		beforeCoins, afterCoins sdk.Coins,
		beforeShares, afterShares sdk.Int,
	) {
		// test token amount has been preserved
		require.True(t,
			!beforeCoins.IsAnyGT(afterCoins),
			"Coins has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeCoins, afterCoins,
		)
		// test share amount has been preserved
		require.True(t,
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

	for i := 0; i < 1000; i++ {
		testPoolInvariants()
	}
}

// TestGetPoolAssetsByDenom tests if `GetPoolAssetsByDenom` succesfully creates a map of denom to pool asset
// given pool asset as parameter
func TestGetPoolAssetsByDenom(t *testing.T) {
	testCases := []struct {
		name                      string
		poolAssets                []balancer.PoolAsset
		expectedPoolAssetsByDenom map[string]balancer.PoolAsset

		err error
	}{
		{
			name:                      "zero pool assets",
			poolAssets:                []balancer.PoolAsset{},
			expectedPoolAssetsByDenom: make(map[string]balancer.PoolAsset),
		},
		{
			name: "one pool asset",
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			expectedPoolAssetsByDenom: map[string]balancer.PoolAsset{
				"uosmo": {
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
		},
		{
			name: "two pool assets",
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("atom", 123),
					Weight: sdk.NewInt(400),
				},
			},
			expectedPoolAssetsByDenom: map[string]balancer.PoolAsset{
				"uosmo": {
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				"atom": {
					Token:  sdk.NewInt64Coin("atom", 123),
					Weight: sdk.NewInt(400),
				},
			},
		},
		{
			name: "duplicate pool assets",
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uosmo", 123),
					Weight: sdk.NewInt(400),
				},
			},
			err: fmt.Errorf(balancer.ErrMsgFormatRepeatingPoolAssetsNotAllowed, "uosmo"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPoolAssetsByDenom, err := balancer.GetPoolAssetsByDenom(tc.poolAssets)

			require.Equal(t, tc.err, err)

			if tc.err != nil {
				return
			}

			require.Equal(t, tc.expectedPoolAssetsByDenom, actualPoolAssetsByDenom)
		})
	}
}

func assertExpectedSharesErrRatio(t *testing.T, expectedShares, actualShares sdk.Int) {
	allowedErrRatioDec, err := sdk.NewDecFromStr(allowedErrRatio)
	require.NoError(t, err)

	errTolerance := osmoutils.ErrTolerance{
		MultiplicativeTolerance: allowedErrRatioDec,
	}

	require.Equal(
		t,
		0,
		errTolerance.Compare(expectedShares, actualShares),
		fmt.Sprintf("expectedShares: %s, actualShares: %s", expectedShares.String(), actualShares.String()))
}

func assertExpectedLiquidity(t *testing.T, expectLiq, tokensJoined, liquidity sdk.Coins) {
	if len(expectLiq) != 0 {
		require.Equal(t, expectLiq, liquidity)
	} else {
		require.Equal(t, tokensJoined, liquidity)
	}
}
