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

	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
)

type GenesisParams struct {
	AirdropSupply sdk.Int

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
