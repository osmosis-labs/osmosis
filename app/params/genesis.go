package params

import (
	"fmt"
	"time"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
)

type GenesisParams struct {
	AirdropSupply sdk.Int

	ConsensusParams *tmproto.ConsensusParams

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

func MainnetGenesisParams() GenesisParams {
	genParams := GenesisParams{}

	genParams.AirdropSupply = sdk.NewIntWithDecimal(1, 14) // 10^15 uosmo, 10^8 (100 million) osmo
	genParams.ChainID = "osmosis-1"
	// genParams.GenesisTime = time.Now() // TODO: Finalize date

	genParams.NativeCoinMetadata = banktypes.Metadata{
		Description: fmt.Sprintf("The native token of Osmosis"),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BaseCoinUnit,
				Exponent: 0,
				Aliases:  nil,
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

	genParams.StakingParams = stakingtypes.DefaultParams()
	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	genParams.StakingParams.MaxValidators = 100
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadata.Base
	genParams.StakingParams.MinCommissionRate = sdk.MustNewDecFromStr("0.05")

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.EpochIdentifier = "daily"                                      // 1 week
	genParams.MintParams.GenesisEpochProvisions = sdk.NewDec(300_000_000).QuoInt64(365) // 300M / 365 = ~821917.8082191781
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadata.Base
	genParams.MintParams.ReductionFactor = sdk.NewDec(2).QuoInt64(3) // 2/3
	genParams.MintParams.ReductionPeriodInEpochs = 365               // 1 year (screw leap years)
	genParams.MintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          sdk.MustNewDecFromStr("0.25"), // 25%
		DeveloperRewards: sdk.MustNewDecFromStr("0.25"), // 25%
		PoolIncentives:   sdk.MustNewDecFromStr("0.5"),  // 50%  TODO: Reduce to 45% once Community Pool Allocation exists
	}
	genParams.MintParams.MintingRewardsDistributionStartEpoch = 1 // TODO: Finalize
	// genParams.MintParams.DeveloperRewardsReceiver

	genParams.DistributionParams = distributiontypes.DefaultParams()
	genParams.DistributionParams.BaseProposerReward = sdk.MustNewDecFromStr("0.01")
	genParams.DistributionParams.BonusProposerReward = sdk.MustNewDecFromStr("0.04")
	genParams.DistributionParams.CommunityTax = sdk.MustNewDecFromStr("0")
	genParams.DistributionParams.WithdrawAddrEnabled = true

	genParams.GovParams = govtypes.DefaultParams()
	genParams.GovParams.DepositParams.MaxDepositPeriod = time.Hour * 24 * 14 // 2 weeks
	genParams.GovParams.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadata.Base,
		genParams.AirdropSupply.QuoRaw(100_000), // 1000 OSMO
	))
	genParams.GovParams.TallyParams.Quorum = sdk.MustNewDecFromStr("0.25") // 25%
	genParams.GovParams.VotingParams.VotingPeriod = time.Hour * 96         // 5 days  TODO: Finalize

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadata.Base,
		genParams.AirdropSupply.QuoRaw(1_000), // 1/1,000 of airdrop supply  TODO: See how crisis invariant fee
	)

	genParams.SlashingParams = slashingtypes.DefaultParams()
	genParams.SlashingParams.SignedBlocksWindow = int64(30000)                       // 30000 blocks (~25 hr at 3 second blocks)
	genParams.SlashingParams.MinSignedPerWindow = sdk.MustNewDecFromStr("0.05")      // 5% minimum liveness
	genParams.SlashingParams.DowntimeJailDuration = time.Second                      // 1 second jail period
	genParams.SlashingParams.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.05") // 5% double sign slashing
	genParams.SlashingParams.SlashFractionDowntime = sdk.ZeroDec()                   // 0% liveness slashing

	genParams.Epochs = epochstypes.DefaultGenesis().Epochs
	for _, epoch := range genParams.Epochs {
		epoch.StartTime = genParams.GenesisTime
	}

	genParams.IncentivesParams = incentivestypes.DefaultParams()
	genParams.IncentivesParams.DistrEpochIdentifier = "daily"

	genParams.ClaimAirdropStartTime = genParams.GenesisTime
	genParams.ClaimDurationUntilDecay = time.Hour * 24 * 60 // 60 days = ~2 months
	genParams.ClaimDurationOfDecay = time.Hour * 24 * 120   // 120 days = ~4 months

	genParams.ConsensusParams = tmtypes.DefaultConsensusParams()
	genParams.ConsensusParams.Evidence.MaxAgeDuration = genParams.StakingParams.UnbondingTime
	genParams.ConsensusParams.Evidence.MaxAgeNumBlocks = int64(genParams.StakingParams.UnbondingTime.Seconds()) / 3
	genParams.ConsensusParams.Version.AppVersion = 1

	return genParams
}

func TestnetGenesisParams() GenesisParams {
	testnetGenesisParams := GenesisParams{}

	testnetGenesisParams.AirdropSupply = sdk.NewIntWithDecimal(1, 15) // 10^15 ions, 10^9 (1 billion) osmo
	testnetGenesisParams.ChainID = "osmo-testnet-thanatos"
	testnetGenesisParams.GenesisTime = time.Now()
	testnetGenesisParams.NativeCoinMetadata = banktypes.Metadata{
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

	testnetGenesisParams.StakingParams = stakingtypes.DefaultParams()
	testnetGenesisParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	testnetGenesisParams.StakingParams.MaxValidators = 100
	testnetGenesisParams.StakingParams.BondDenom = testnetGenesisParams.NativeCoinMetadata.Base
	testnetGenesisParams.StakingParams.MinCommissionRate = sdk.MustNewDecFromStr("0.05")

	testnetGenesisParams.MintParams = minttypes.DefaultParams()
	testnetGenesisParams.MintParams.EpochIdentifier = "weekly"                                                                // 1 week
	testnetGenesisParams.MintParams.GenesisEpochProvisions = sdk.NewDecFromInt(testnetGenesisParams.AirdropSupply.QuoRaw(10)) // 10% of airdrop supply
	testnetGenesisParams.MintParams.MintDenom = testnetGenesisParams.NativeCoinMetadata.Base
	testnetGenesisParams.MintParams.ReductionFactor = sdk.NewDecWithPrec(5, 1) // 0.5
	testnetGenesisParams.MintParams.ReductionPeriodInEpochs = 52 * 3           // 3 years

	testnetGenesisParams.DistributionParams = distributiontypes.DefaultParams()
	testnetGenesisParams.DistributionParams.BaseProposerReward = sdk.MustNewDecFromStr("0.01")
	testnetGenesisParams.DistributionParams.BonusProposerReward = sdk.MustNewDecFromStr("0")
	testnetGenesisParams.DistributionParams.CommunityTax = sdk.MustNewDecFromStr("0")
	testnetGenesisParams.DistributionParams.WithdrawAddrEnabled = true

	testnetGenesisParams.GovParams = govtypes.DefaultParams()
	testnetGenesisParams.GovParams.DepositParams.MaxDepositPeriod = time.Hour * 24 * 7 // 1 week
	testnetGenesisParams.GovParams.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		testnetGenesisParams.NativeCoinMetadata.Base,
		testnetGenesisParams.AirdropSupply.QuoRaw(1_000_000), // 1 millionth of airdrop supply
	))
	testnetGenesisParams.GovParams.TallyParams.Quorum = sdk.MustNewDecFromStr("0.25") // 25%
	testnetGenesisParams.GovParams.VotingParams.VotingPeriod = time.Hour * 6          // 6 hours

	testnetGenesisParams.CrisisConstantFee = sdk.NewCoin(
		testnetGenesisParams.NativeCoinMetadata.Base,
		testnetGenesisParams.AirdropSupply.QuoRaw(100_000), // 1/100,000 of airdrop supply
	)

	testnetGenesisParams.SlashingParams = slashingtypes.DefaultParams()
	testnetGenesisParams.SlashingParams.SignedBlocksWindow = int64(10000)

	testnetGenesisParams.Epochs = epochstypes.DefaultGenesis().Epochs
	for _, epoch := range testnetGenesisParams.Epochs {
		epoch.StartTime = testnetGenesisParams.GenesisTime
	}

	testnetGenesisParams.IncentivesParams = incentivestypes.DefaultParams()
	testnetGenesisParams.IncentivesParams.DistrEpochIdentifier = "daily"

	testnetGenesisParams.ClaimAirdropStartTime = testnetGenesisParams.GenesisTime
	testnetGenesisParams.ClaimDurationUntilDecay = time.Hour            // 1 hour
	testnetGenesisParams.ClaimDurationOfDecay = time.Hour * 24 * 7 * 12 // 12 weeks

	return testnetGenesisParams
}
