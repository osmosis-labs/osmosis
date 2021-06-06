package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	appParams "github.com/osmosis-labs/osmosis/app/params"

	claimtypes "github.com/osmosis-labs/osmosis/x/claim/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
)

func PrepareGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis",
		Short: "Prepare a genesis file with initial setup",
		Long: `Prepare a genesis file with initial setup.
Examples include:
	- Setting module initial params
	- Setting denom metadata
Example:
	osmosisd prepare-genesis mainnet
	- Check input genesis:
		file is at ~/.gaiad/config/genesis.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			// read genesis file
			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// get genesis params
			var genesisParams GenesisParams
			network := args[0]
			if network == "testnet" {
				genesisParams = TestnetGenesisParams()
			} else if network == "mainnet" {
				genesisParams = MainnetGenesisParams()
			} else {
				return fmt.Errorf("please choose 'mainnet' or 'testnet'")
			}

			// run Prepare Genesis
			appState, genDoc, err = PrepareGenesis(clientCtx, appState, genDoc, genesisParams)

			// validate genesis state
			if err = mbm.ValidateGenesis(cdc, clientCtx.TxConfig, appState); err != nil {
				return fmt.Errorf("error validating genesis file: %s", err.Error())
			}

			// save genesis
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			err = genutil.ExportGenesisFile(genDoc, genFile)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func PrepareGenesis(clientCtx client.Context, appState map[string]json.RawMessage, genDoc *tmtypes.GenesisDoc, genesisParams GenesisParams) (map[string]json.RawMessage, *tmtypes.GenesisDoc, error) {
	depCdc := clientCtx.JSONMarshaler
	cdc := depCdc.(codec.Marshaler)

	// chain params genesis
	genDoc.GenesisTime = genesisParams.GenesisTime
	genDoc.ChainID = genesisParams.ChainID

	// bank module genesis
	bankGenState := banktypes.GetGenesisStateFromAppState(depCdc, appState)
	bankGenState.DenomMetadata = []banktypes.Metadata{
		genesisParams.NativeCoinMetadata,
	}
	bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	appState[banktypes.ModuleName] = bankGenStateBz

	// staking module genesis
	stakingGenState := stakingtypes.GetGenesisStateFromAppState(depCdc, appState)
	stakingGenState.Params = genesisParams.StakingParams
	stakingGenStateBz, err := cdc.MarshalJSON(stakingGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	}
	appState[stakingtypes.ModuleName] = stakingGenStateBz

	// mint module genesis
	mintGenState := minttypes.DefaultGenesisState()
	mintGenState.Params = genesisParams.MintParams
	mintGenStateBz, err := cdc.MarshalJSON(mintGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal mint genesis state: %w", err)
	}
	appState[minttypes.ModuleName] = mintGenStateBz

	// distribution module genesis
	distributionGenState := distributiontypes.DefaultGenesisState()
	distributionGenState.Params = genesisParams.DistributionParams
	// TODO Set initial community pool
	// distributionGenState.FeePool.CommunityPool = sdk.NewDecCoins()
	distributionGenStateBz, err := cdc.MarshalJSON(distributionGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal distribution genesis state: %w", err)
	}
	appState[distributiontypes.ModuleName] = distributionGenStateBz

	// gov module genesis
	govGenState := govtypes.DefaultGenesisState()
	govGenState.DepositParams = genesisParams.GovParams.DepositParams
	govGenState.TallyParams = genesisParams.GovParams.TallyParams
	govGenState.VotingParams = genesisParams.GovParams.VotingParams
	govGenStateBz, err := cdc.MarshalJSON(govGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal gov genesis state: %w", err)
	}
	appState[govtypes.ModuleName] = govGenStateBz

	// crisis module genesis
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = genesisParams.CrisisConstantFee
	// TODO Set initial community pool
	// distributionGenState.FeePool.CommunityPool = sdk.NewDecCoins()
	crisisGenStateBz, err := cdc.MarshalJSON(crisisGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal crisis genesis state: %w", err)
	}
	appState[crisistypes.ModuleName] = crisisGenStateBz

	// slashing module genesis
	slashingGenState := slashingtypes.DefaultGenesisState()
	slashingGenState.Params = genesisParams.SlashingParams
	slashingGenStateBz, err := cdc.MarshalJSON(slashingGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal slashing genesis state: %w", err)
	}
	appState[slashingtypes.ModuleName] = slashingGenStateBz

	// incentives module genesis
	incentivesGenState := incentivestypes.DefaultGenesis()
	incentivesGenState.Params = genesisParams.IncentivesParams
	incentivesGenStateBz, err := cdc.MarshalJSON(incentivesGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal incentives genesis state: %w", err)
	}
	appState[incentivestypes.ModuleName] = incentivesGenStateBz

	// epochs module genesis
	epochsGenState := epochstypes.DefaultGenesis()
	epochsGenState.Epochs = genesisParams.Epochs
	epochsGenStateBz, err := cdc.MarshalJSON(epochsGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal epochs genesis state: %w", err)
	}
	appState[epochstypes.ModuleName] = epochsGenStateBz

	// claim module genesis
	claimGenState := claimtypes.GetGenesisStateFromAppState(depCdc, appState)
	claimGenState.Params = genesisParams.ClaimParams
	claimGenStateBz, err := cdc.MarshalJSON(claimGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal claim genesis state: %w", err)
	}
	appState[claimtypes.ModuleName] = claimGenStateBz

	// poolincentives module genesis
	poolincentivesGenState := &genesisParams.PoolIncentivesGenesis
	poolincentivesGenStateBz, err := cdc.MarshalJSON(poolincentivesGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal poolincentives genesis state: %w", err)
	}
	appState[poolincentivestypes.ModuleName] = poolincentivesGenStateBz

	// return appState and genDoc
	return appState, genDoc, nil
}

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

	PoolIncentivesGenesis poolincentivestypes.GenesisState

	Epochs []epochstypes.EpochInfo

	ClaimParams claimtypes.Params
}

func MainnetGenesisParams() GenesisParams {
	genParams := GenesisParams{}

	genParams.AirdropSupply = sdk.NewIntWithDecimal(5, 13) // 5*10^13 uosmo, 5*10^7 (50 million) osmo
	genParams.ChainID = "osmosis-1"
	genParams.GenesisTime = time.Date(2021, 6, 16, 17, 0, 0, 0, time.UTC) // Jun 16, 2021 - 17:00 UTC

	genParams.NativeCoinMetadata = banktypes.Metadata{
		Description: fmt.Sprintf("The native token of Osmosis"),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    appParams.BaseCoinUnit,
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    appParams.HumanCoinUnit,
				Exponent: appParams.OsmoExponent,
				Aliases:  nil,
			},
		},
		Base:    appParams.BaseCoinUnit,
		Display: appParams.HumanCoinUnit,
	}

	genParams.StakingParams = stakingtypes.DefaultParams()
	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	genParams.StakingParams.MaxValidators = 100
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadata.Base
	genParams.StakingParams.MinCommissionRate = sdk.MustNewDecFromStr("0.05")

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.EpochIdentifier = "day"                                        // 1 week
	genParams.MintParams.GenesisEpochProvisions = sdk.NewDec(300_000_000).QuoInt64(365) // 300M / 365 = ~821917.8082191781
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadata.Base
	genParams.MintParams.ReductionFactor = sdk.NewDec(2).QuoInt64(3) // 2/3
	genParams.MintParams.ReductionPeriodInEpochs = 365               // 1 year (screw leap years)
	genParams.MintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          sdk.MustNewDecFromStr("0.25"), // 25%
		DeveloperRewards: sdk.MustNewDecFromStr("0.25"), // 25%
		PoolIncentives:   sdk.MustNewDecFromStr("0.45"), // 45%
		CommunityPool:    sdk.MustNewDecFromStr("0.05"), // 5%
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
	genParams.GovParams.VotingParams.VotingPeriod = time.Hour * 72         // 3 days

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadata.Base,
		genParams.AirdropSupply.QuoRaw(1_000), // 1/1,000 of airdrop supply  TODO: See how crisis invariant fee
	)

	genParams.SlashingParams = slashingtypes.DefaultParams()
	genParams.SlashingParams.SignedBlocksWindow = int64(30000)                       // 30000 blocks (~25 hr at 3 second blocks)
	genParams.SlashingParams.MinSignedPerWindow = sdk.MustNewDecFromStr("0.05")      // 5% minimum liveness
	genParams.SlashingParams.DowntimeJailDuration = time.Minute                      // 1 minute jail period
	genParams.SlashingParams.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.05") // 5% double sign slashing
	genParams.SlashingParams.SlashFractionDowntime = sdk.ZeroDec()                   // 0% liveness slashing

	genParams.Epochs = epochstypes.DefaultGenesis().Epochs
	for _, epoch := range genParams.Epochs {
		epoch.StartTime = genParams.GenesisTime
	}

	genParams.IncentivesParams = incentivestypes.DefaultParams()
	genParams.IncentivesParams.DistrEpochIdentifier = "day"

	genParams.ClaimParams = claimtypes.Params{
		AirdropStartTime:   genParams.GenesisTime,
		DurationUntilDecay: time.Hour * 24 * 60,  // 60 days = ~2 months
		DurationOfDecay:    time.Hour * 24 * 120, // 120 days = ~4 months
	}

	genParams.ConsensusParams = tmtypes.DefaultConsensusParams()
	genParams.ConsensusParams.Evidence.MaxAgeDuration = genParams.StakingParams.UnbondingTime
	genParams.ConsensusParams.Evidence.MaxAgeNumBlocks = int64(genParams.StakingParams.UnbondingTime.Seconds()) / 3
	genParams.ConsensusParams.Version.AppVersion = 1

	genParams.PoolIncentivesGenesis = *poolincentivestypes.DefaultGenesisState()
	genParams.PoolIncentivesGenesis.Params.MintedDenom = genParams.NativeCoinMetadata.Base
	genParams.PoolIncentivesGenesis.LockableDurations = []time.Duration{
		time.Hour * 24,      // 1 day
		time.Hour * 24 * 7,  // 7 day
		time.Hour * 24 * 14, // 14 days
	}
	genParams.PoolIncentivesGenesis.DistrInfo = &poolincentivestypes.DistrInfo{
		TotalWeight: sdk.NewInt(1),
		Records: []poolincentivestypes.DistrRecord{
			{
				GaugeId: 0,
				Weight:  sdk.NewInt(1),
			},
		},
	}

	return genParams
}

func TestnetGenesisParams() GenesisParams {
	genParams := GenesisParams{}

	genParams.AirdropSupply = sdk.NewIntWithDecimal(5, 13) // 5*10^13 uosmo, 5*10^7 (50 million) osmo
	genParams.ChainID = "osmo-testnet-1"
	genParams.GenesisTime = time.Now()

	genParams.NativeCoinMetadata = banktypes.Metadata{
		Description: fmt.Sprintf("The native token of Osmosis"),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    appParams.BaseCoinUnit,
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    appParams.HumanCoinUnit,
				Exponent: appParams.OsmoExponent,
				Aliases:  nil,
			},
		},
		Base:    appParams.BaseCoinUnit,
		Display: appParams.HumanCoinUnit,
	}

	genParams.StakingParams = stakingtypes.DefaultParams()
	genParams.StakingParams.UnbondingTime = time.Hour * 24 // 1 day
	genParams.StakingParams.MaxValidators = 100
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadata.Base
	genParams.StakingParams.MinCommissionRate = sdk.MustNewDecFromStr("0.05")

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.EpochIdentifier = "hour"                                       // 1 hour
	genParams.MintParams.GenesisEpochProvisions = sdk.NewDec(300_000_000).QuoInt64(365) // 300M / 365 = ~821917.8082191781
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadata.Base
	genParams.MintParams.ReductionFactor = sdk.NewDec(2).QuoInt64(3) // 2/3
	genParams.MintParams.ReductionPeriodInEpochs = 6                 // 6 hours
	genParams.MintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          sdk.MustNewDecFromStr("0.25"), // 25%
		DeveloperRewards: sdk.MustNewDecFromStr("0.25"), // 25%
		PoolIncentives:   sdk.MustNewDecFromStr("0.45"), // 45%
		CommunityPool:    sdk.MustNewDecFromStr("0.05"), // 5%
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
		sdk.NewInt(1000000), // 1 OSMO
	))
	genParams.GovParams.TallyParams.Quorum = sdk.MustNewDecFromStr("0.0000001") // 25%
	genParams.GovParams.VotingParams.VotingPeriod = time.Hour * 3               // 3 hours

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadata.Base,
		genParams.AirdropSupply.QuoRaw(1_000), // 1/1,000 of airdrop supply  TODO: See how crisis invariant fee
	)

	genParams.SlashingParams = slashingtypes.DefaultParams()
	genParams.SlashingParams.SignedBlocksWindow = int64(30000)                       // 30000 blocks (~25 hr at 3 second blocks)
	genParams.SlashingParams.MinSignedPerWindow = sdk.MustNewDecFromStr("0.05")      // 5% minimum liveness
	genParams.SlashingParams.DowntimeJailDuration = time.Minute                      // 1 minute jail period
	genParams.SlashingParams.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.05") // 5% double sign slashing
	genParams.SlashingParams.SlashFractionDowntime = sdk.ZeroDec()                   // 0% liveness slashing

	genParams.Epochs = epochstypes.DefaultGenesis().Epochs

	genParams.Epochs = append(genParams.Epochs, epochstypes.EpochInfo{
		Identifier:            "hour",
		StartTime:             time.Time{},
		Duration:              time.Hour,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
		CurrentEpochEnded:     true,
	})

	for _, epoch := range genParams.Epochs {
		epoch.StartTime = genParams.GenesisTime
	}

	genParams.IncentivesParams = incentivestypes.DefaultParams()
	genParams.IncentivesParams.DistrEpochIdentifier = "hour"

	genParams.ClaimParams = claimtypes.Params{
		AirdropStartTime:   genParams.GenesisTime,
		DurationUntilDecay: time.Second * 10, // 60 days = ~2 months
		DurationOfDecay:    time.Second * 30, // 120 days = ~4 months
	}

	genParams.ConsensusParams = tmtypes.DefaultConsensusParams()
	genParams.ConsensusParams.Evidence.MaxAgeDuration = genParams.StakingParams.UnbondingTime
	genParams.ConsensusParams.Evidence.MaxAgeNumBlocks = int64(genParams.StakingParams.UnbondingTime.Seconds()) / 3
	genParams.ConsensusParams.Version.AppVersion = 1

	genParams.PoolIncentivesGenesis = *poolincentivestypes.DefaultGenesisState()
	genParams.PoolIncentivesGenesis.Params.MintedDenom = genParams.NativeCoinMetadata.Base
	genParams.PoolIncentivesGenesis.LockableDurations = []time.Duration{
		time.Hour * 1, // 1 hour
		time.Hour * 3, // 3 hours
		time.Hour * 7, // 7 hours
	}
	genParams.PoolIncentivesGenesis.DistrInfo = &poolincentivestypes.DistrInfo{
		TotalWeight: sdk.NewInt(1),
		Records: []poolincentivestypes.DistrRecord{
			{
				GaugeId: 0,
				Weight:  sdk.NewInt(1),
			},
		},
	}

	return genParams
}
