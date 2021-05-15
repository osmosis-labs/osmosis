package params

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/c-osmosis/osmosis/x/epochs/types"
	incentivestypes "github.com/c-osmosis/osmosis/x/incentives/types"
	minttypes "github.com/c-osmosis/osmosis/x/mint/types"
)

type NetworkParams struct {
	TotalSupply        sdk.Int
	ChainID            string
	GenesisTime        time.Time
	NativeCoinMetadata banktypes.Metadata

	StakingParams      stakingtypes.Params
	MintParams         minttypes.Params
	DistributionParams distributiontypes.Params
	GovParams          govtypes.Params

	CrisisConstantFee sdk.Coin

	SlashingParams   slashingtypes.Params
	IncentivesParams incentivestypes.Params

	Epochs []epochstypes.EpochInfo

	ClaimAirdropStartTime   time.Time
	ClaimDurationUntilDecay time.Duration
	ClaimDurationOfDecay    time.Duration
}

func TestnetNetworkParams() NetworkParams {
	testnetNetworkParams := NetworkParams{}

	testnetNetworkParams.TotalSupply = sdk.NewIntWithDecimal(1, 9) // 1 Billion
	testnetNetworkParams.ChainID = "osmo-testnet-thanatos"
	testnetNetworkParams.GenesisTime = time.Now()
	testnetNetworkParams.NativeCoinMetadata = banktypes.Metadata{
		Description: fmt.Sprintf("The native token of Osmosis"),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BaseCoinUnit,
				Exponent: 0,
				Aliases: []string{
					fmt.Sprintf("u%s", HumanCoinUnit),
				},
			},
			{
				Denom:    HumanCoinUnit,
				Exponent: OsmoExponent,
				Aliases:  nil,
			},
		},
		Base:    BaseCoinUnit,
		Display: HumanCoinUnit,
	}

	testnetNetworkParams.StakingParams = stakingtypes.DefaultParams()
	testnetNetworkParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	testnetNetworkParams.StakingParams.MaxValidators = 100
	testnetNetworkParams.StakingParams.BondDenom = testnetNetworkParams.NativeCoinMetadata.Base

	testnetNetworkParams.MintParams = minttypes.DefaultParams()
	testnetNetworkParams.MintParams.EpochDuration = time.Hour * 24 * 7                                                      // 1 week
	testnetNetworkParams.MintParams.GenesisEpochProvisions = sdk.NewDecFromInt(testnetNetworkParams.TotalSupply.QuoRaw(10)) // 10% of final supply
	testnetNetworkParams.MintParams.MintDenom = testnetNetworkParams.NativeCoinMetadata.Base
	testnetNetworkParams.MintParams.ReductionFactor = sdk.NewDecWithPrec(5, 1) // 0.5
	testnetNetworkParams.MintParams.ReductionPeriodInEpochs = 52 * 3           // 3 years

	testnetNetworkParams.DistributionParams = distributiontypes.DefaultParams()
	testnetNetworkParams.DistributionParams.BaseProposerReward = sdk.MustNewDecFromStr("0.01")
	testnetNetworkParams.DistributionParams.BonusProposerReward = sdk.MustNewDecFromStr("0")
	testnetNetworkParams.DistributionParams.CommunityTax = sdk.MustNewDecFromStr("0")
	testnetNetworkParams.DistributionParams.WithdrawAddrEnabled = true

	testnetNetworkParams.GovParams = govtypes.DefaultParams()
	testnetNetworkParams.GovParams.DepositParams.MaxDepositPeriod = time.Hour * 24 * 7 // 1 week
	testnetNetworkParams.GovParams.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		testnetNetworkParams.NativeCoinMetadata.Base,
		testnetNetworkParams.TotalSupply.QuoRaw(1_000_000), // 1 millionth of supply
	))
	testnetNetworkParams.GovParams.TallyParams.Quorum = sdk.MustNewDecFromStr("0.25") // 25%
	testnetNetworkParams.GovParams.VotingParams.VotingPeriod = time.Hour * 24 * 7     // 1 week

	testnetNetworkParams.CrisisConstantFee = sdk.NewCoin(
		testnetNetworkParams.NativeCoinMetadata.Base,
		testnetNetworkParams.TotalSupply.QuoRaw(100_000), // 1/100,000 of total supply
	)

	testnetNetworkParams.SlashingParams = slashingtypes.DefaultParams()
	testnetNetworkParams.SlashingParams.SignedBlocksWindow = int64(10000)

	testnetNetworkParams.Epochs = epochstypes.DefaultGenesis().Epochs

	for _, epoch := range testnetNetworkParams.Epochs {
		epoch.StartTime = testnetNetworkParams.GenesisTime
	}

	testnetNetworkParams.IncentivesParams = incentivestypes.DefaultParams()
	testnetNetworkParams.IncentivesParams.DistrEpochIdentifier = "weekly"

	testnetNetworkParams.ClaimAirdropStartTime = testnetNetworkParams.GenesisTime
	testnetNetworkParams.ClaimDurationUntilDecay = time.Hour * 24 * 7 * 12 // 12 weeks
	testnetNetworkParams.ClaimDurationOfDecay = time.Hour * 24 * 7 * 12    // 12 weeks

	return testnetNetworkParams
}
