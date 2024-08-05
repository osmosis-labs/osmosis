package cmd

import (
	"encoding/json"
	"fmt"
	txfeestypes "github.com/osmosis-labs/osmosis/v23/x/txfees/types"
	"time"

	"github.com/spf13/cobra"

	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appParams "github.com/osmosis-labs/osmosis/v23/app/params"

	incentivestypes "github.com/osmosis-labs/osmosis/v23/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v23/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v23/x/pool-incentives/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// PrepareGenesisCmd returns prepare-genesis cobra Command.
//

func PrepareGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis",
		Short: "Prepare a genesis file with initial setup",
		Long: `Prepare a genesis file with initial setup.
Examples include:
	- Setting module initial params
	- Setting denom metadata
Example:
	symphonyd prepare-genesis mainnet symphony-1
	- Check input genesis:
		file is at ~/.symphonyd/config/genesis.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.Codec
			cdc := depCdc
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

			// get genesis params
			chainID := args[1]

			// run Prepare Genesis
			appState, genDoc, err = PrepareGenesis(clientCtx, appState, genDoc, genesisParams, chainID)
			if err != nil {
				return err
			}

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

func PrepareGenesis(clientCtx client.Context, appState map[string]json.RawMessage, genDoc *tmtypes.GenesisDoc, genesisParams GenesisParams, chainID string) (map[string]json.RawMessage, *tmtypes.GenesisDoc, error) {
	depCdc := clientCtx.Codec
	cdc := depCdc

	// chain params genesis
	genDoc.ChainID = chainID
	genDoc.GenesisTime = genesisParams.GenesisTime

	genDoc.ConsensusParams = genesisParams.ConsensusParams

	// ---
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
	govGenState := govtypesv1.DefaultGenesisState()
	govGenState.Params = &genesisParams.GovParams
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
	incentivesGenState := incentivestypes.GetGenesisStateFromAppState(depCdc, appState)
	incentivesGenState.Params = genesisParams.IncentivesGenesis.Params
	incentivesGenState.LockableDurations = genesisParams.IncentivesGenesis.LockableDurations
	incentivesGenState.Gauges = genesisParams.IncentivesGenesis.Gauges
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

	// poolincentives module genesis
	poolincentivesGenState := &genesisParams.PoolIncentivesGenesis
	poolincentivesGenStateBz, err := cdc.MarshalJSON(poolincentivesGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal poolincentives genesis state: %w", err)
	}
	appState[poolincentivestypes.ModuleName] = poolincentivesGenStateBz

	// txtypes module genesis
	txfeesGenState := &genesisParams.TxFees
	txfeesGenStateBz, err := cdc.MarshalJSON(txfeesGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal txfees genesis state: %w", err)
	}
	appState[txfeestypes.ModuleName] = txfeesGenStateBz

	// return appState and genDoc
	return appState, genDoc, nil
}

type GenesisParams struct {
	//AirdropSupply osmomath.Int

	//StrategicReserveAccounts []banktypes.Balance

	ConsensusParams *tmtypes.ConsensusParams

	GenesisTime         time.Time
	NativeCoinMetadatas []banktypes.Metadata

	StakingParams      stakingtypes.Params
	MintParams         minttypes.Params
	DistributionParams distributiontypes.Params
	GovParams          govtypesv1.Params

	CrisisConstantFee sdk.Coin

	SlashingParams    slashingtypes.Params
	IncentivesGenesis incentivestypes.GenesisState

	PoolIncentivesGenesis poolincentivestypes.GenesisState

	Epochs []epochstypes.EpochInfo

	TxFees txfeestypes.GenesisState
}

func MainnetGenesisParams() GenesisParams {
	genParams := GenesisParams{}

	//genParams.AirdropSupply = osmomath.NewIntWithDecimal(5, 13)           // 5*10^13 note, 5*10^7 (50 million) melody
	genParams.GenesisTime = time.Now()

	genParams.NativeCoinMetadatas = []banktypes.Metadata{
		{
			Description: "The native token of Symphony",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    appParams.BaseCoinUnit,
					Exponent: 0,
					Aliases:  nil,
				},
				{
					Denom:    appParams.HumanCoinUnit,
					Exponent: appParams.MelodyExponent,
					Aliases:  nil,
				},
			},
			Base:    appParams.BaseCoinUnit,
			Display: appParams.HumanCoinUnit,
		},
	}

	genParams.StakingParams = stakingtypes.DefaultParams()
	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	genParams.StakingParams.MaxValidators = 100
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.StakingParams.MinCommissionRate = osmomath.MustNewDecFromStr("0.05")

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.EpochIdentifier = "week"
	genParams.MintParams.GenesisEpochProvisions = osmomath.NewDec(150250000000)
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.MintParams.ReductionFactor = osmomath.NewDecWithPrec(5, 1) // 0.5
	genParams.MintParams.ReductionPeriodInEpochs = 208                   // 4 years
	genParams.MintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          osmomath.NewDecWithPrec(7, 1),  // 0.7
		PoolIncentives:   osmomath.NewDec(0),             // 0.0
		DeveloperRewards: osmomath.NewDecWithPrec(25, 2), // 0.25
		CommunityPool:    osmomath.NewDecWithPrec(5, 2),  // 0.05
	}
	genParams.MintParams.MintingRewardsDistributionStartEpoch = 1
	genParams.MintParams.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{
		{
			Address: "symphony1g6pxgl8g0rnk7q86j9zh7yxsqdsn7jvdmc8fkx",
			Weight:  osmomath.MustNewDecFromStr("1.0"),
		},
	}

	genParams.DistributionParams = distributiontypes.DefaultParams()
	genParams.DistributionParams.CommunityTax = osmomath.MustNewDecFromStr("0")
	genParams.DistributionParams.WithdrawAddrEnabled = true

	genParams.GovParams = govtypesv1.DefaultParams()
	*genParams.GovParams.MaxDepositPeriod = time.Hour * 24 * 14 // 2 weeks
	genParams.GovParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(2_500_000_000),
	))
	genParams.GovParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(5_000_000_000),
	))
	genParams.GovParams.Quorum = "0.2"                     // 20%
	*genParams.GovParams.VotingPeriod = time.Hour * 24 * 3 // 3 days

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(500_000_000_000),
	)

	genParams.SlashingParams = slashingtypes.DefaultParams()
	genParams.SlashingParams.SignedBlocksWindow = int64(30000)                            // 30000 blocks (~41 hr at 5 second blocks)
	genParams.SlashingParams.MinSignedPerWindow = osmomath.MustNewDecFromStr("0.05")      // 5% minimum liveness
	genParams.SlashingParams.DowntimeJailDuration = time.Minute                           // 1 minute jail period
	genParams.SlashingParams.SlashFractionDoubleSign = osmomath.MustNewDecFromStr("0.05") // 5% double sign slashing
	genParams.SlashingParams.SlashFractionDowntime = osmomath.ZeroDec()                   // 0% liveness slashing

	genParams.Epochs = epochstypes.DefaultGenesis().Epochs
	for _, epoch := range genParams.Epochs {
		epoch.StartTime = genParams.GenesisTime
	}

	genParams.IncentivesGenesis = *incentivestypes.DefaultGenesis()
	genParams.IncentivesGenesis.Params.DistrEpochIdentifier = "day"
	genParams.IncentivesGenesis.LockableDurations = []time.Duration{
		time.Hour * 24,      // 1 day
		time.Hour * 24 * 7,  // 7 day
		time.Hour * 24 * 14, // 14 days
	}

	genParams.ConsensusParams = tmtypes.DefaultConsensusParams()
	genParams.ConsensusParams.Block.MaxBytes = 5 * 1024 * 1024
	genParams.ConsensusParams.Block.MaxGas = 6_000_000
	genParams.ConsensusParams.Evidence.MaxAgeDuration = genParams.StakingParams.UnbondingTime
	genParams.ConsensusParams.Evidence.MaxAgeNumBlocks = int64(genParams.StakingParams.UnbondingTime.Seconds()) / 3
	genParams.ConsensusParams.Version.App = 1

	genParams.PoolIncentivesGenesis = *poolincentivestypes.DefaultGenesisState()
	genParams.PoolIncentivesGenesis.Params.MintedDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.PoolIncentivesGenesis.LockableDurations = genParams.IncentivesGenesis.LockableDurations
	genParams.PoolIncentivesGenesis.DistrInfo = &poolincentivestypes.DistrInfo{
		TotalWeight: osmomath.NewInt(1000),
		Records: []poolincentivestypes.DistrRecord{
			{
				GaugeId: 0,
				Weight:  osmomath.NewInt(1000),
			},
		},
	}

	genParams.TxFees = *txfeestypes.DefaultGenesis()
	genParams.TxFees.Basedenom = genParams.NativeCoinMetadatas[0].Base

	return genParams
}

func TestnetGenesisParams() GenesisParams {
	genParams := MainnetGenesisParams()

	genParams.GenesisTime = time.Now()

	genParams.Epochs = append(genParams.Epochs, epochstypes.EpochInfo{
		Identifier:            "15min",
		StartTime:             time.Time{},
		Duration:              15 * time.Minute,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	})

	for _, epoch := range genParams.Epochs {
		epoch.StartTime = genParams.GenesisTime
	}

	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks

	genParams.MintParams.EpochIdentifier = "15min"     // 15min
	genParams.MintParams.ReductionPeriodInEpochs = 192 // 2 days

	genParams.GovParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(1000000), // 1 OSMO
	))
	genParams.GovParams.Quorum = "0.0000000001"           // 0.00000001%
	*genParams.GovParams.VotingPeriod = time.Second * 300 // 300 seconds

	genParams.IncentivesGenesis = *incentivestypes.DefaultGenesis()
	genParams.IncentivesGenesis.Params.DistrEpochIdentifier = "15min"
	genParams.IncentivesGenesis.LockableDurations = []time.Duration{
		time.Minute * 30, // 30 min
		time.Hour * 1,    // 1 hour
		time.Hour * 2,    // 2 hours
	}

	genParams.PoolIncentivesGenesis.LockableDurations = genParams.IncentivesGenesis.LockableDurations

	return genParams
}
