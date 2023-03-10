package initialization

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
	tmjson "github.com/tendermint/tendermint/libs/json"

	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v15/x/mint/types"
	poolitypes "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	twaptypes "github.com/osmosis-labs/osmosis/v15/x/twap/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	types1 "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/osmosis-labs/osmosis/v15/tests/e2e/util"
)

// NodeConfig is a confiuration for the node supplied from the test runner
// to initialization scripts. It should be backwards compatible with earlier
// versions. If this struct is updated, the change must be backported to earlier
// branches that might be used for upgrade testing.
type NodeConfig struct {
	Name               string // name of the config that will also be assigned to Docke container.
	Pruning            string // default, nothing, everything, or custom
	PruningKeepRecent  string // keep all of the last N states (only used with custom pruning)
	PruningInterval    string // delete old states from every Nth block (only used with custom pruning)
	SnapshotInterval   uint64 // statesync snapshot every Nth block (0 to disable)
	SnapshotKeepRecent uint32 // number of recent snapshots to keep and serve (0 to keep all)
	IsValidator        bool   // flag indicating whether a node should be a validator
}

const (
	// common
	OsmoDenom           = "uosmo"
	IonDenom            = "uion"
	StakeDenom          = "stake"
	AtomDenom           = "uatom"
	OsmoIBCDenom        = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	StakeIBCDenom       = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B7787"
	E2EFeeToken         = "e2e-default-feetoken"
	UstIBCDenom         = "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC"
	LuncIBCDenom        = "ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0"
	MinGasPrice         = "0.000"
	IbcSendAmount       = 3300000000
	ValidatorWalletName = "val"
	// chainA
	ChainAID      = "osmo-test-a"
	OsmoBalanceA  = 20000000000000
	IonBalanceA   = 100000000000
	StakeBalanceA = 110000000000
	StakeAmountA  = 100000000000
	UstBalanceA   = 500000000000000
	LuncBalanceA  = 500000000000000
	// chainB
	ChainBID          = "osmo-test-b"
	OsmoBalanceB      = 500000000000
	IonBalanceB       = 100000000000
	StakeBalanceB     = 440000000000
	StakeAmountB      = 400000000000
	GenesisFeeBalance = 100000000000
	WalletFeeBalance  = 100000000

	EpochDayDuration      = time.Second * 60
	EpochWeekDuration     = time.Second * 120
	TWAPPruningKeepPeriod = EpochDayDuration / 4

	// Denoms for testing Stride migration in v15.
	// Can be removed after v15 upgrade.
	StOsmoDenom               = "stOsmo"
	JunoDenom                 = "juno"
	StJunoDenom               = "stJuno"
	StarsDenom                = "stars"
	StStarsDenom              = "stStars"
	DefaultStrideDenomBalance = OsmoBalanceA

	// Stride pool ids to migrate
	// Can be removed after v15 upgrade.
	StOSMO_OSMOPoolId   = 833
	StJUNO_JUNOPoolId   = 817
	StSTARS_STARSPoolId = 810
)

var (
	StakeAmountIntA  = sdk.NewInt(StakeAmountA)
	StakeAmountCoinA = sdk.NewCoin(OsmoDenom, StakeAmountIntA)
	StakeAmountIntB  = sdk.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(OsmoDenom, StakeAmountIntB)

	// Pool balances for testing Stride migration in v15.
	// Can be removed after v15 upgrade.
	StridePoolBalances = fmt.Sprintf("%d%s,%d%s,%d%s,%d%s,%d%s", DefaultStrideDenomBalance, StOsmoDenom, DefaultStrideDenomBalance, JunoDenom, DefaultStrideDenomBalance, StJunoDenom, DefaultStrideDenomBalance, StarsDenom, DefaultStrideDenomBalance, StStarsDenom)

	InitBalanceStrA = fmt.Sprintf("%d%s,%d%s,%d%s,%d%s,%d%s", OsmoBalanceA, OsmoDenom, StakeBalanceA, StakeDenom, IonBalanceA, IonDenom, UstBalanceA, UstIBCDenom, LuncBalanceA, LuncIBCDenom)
	InitBalanceStrB = fmt.Sprintf("%d%s,%d%s,%d%s", OsmoBalanceB, OsmoDenom, StakeBalanceB, StakeDenom, IonBalanceB, IonDenom)
	OsmoToken       = sdk.NewInt64Coin(OsmoDenom, IbcSendAmount)  // 3,300uosmo
	StakeToken      = sdk.NewInt64Coin(StakeDenom, IbcSendAmount) // 3,300ustake
	tenOsmo         = sdk.Coins{sdk.NewInt64Coin(OsmoDenom, 10_000_000)}
	fiftyOsmo       = sdk.Coins{sdk.NewInt64Coin(OsmoDenom, 50_000_000)}
	WalletFeeTokens = sdk.NewCoin(E2EFeeToken, sdk.NewInt(WalletFeeBalance))
)

func addAccount(path, moniker, amountStr string, accAddr sdk.AccAddress, forkHeight int) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(path)
	config.Moniker = moniker

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}
	coins = coins.Add(sdk.NewCoin(E2EFeeToken, sdk.NewInt(GenesisFeeBalance)))

	balances := banktypes.Balance{Address: accAddr.String(), Coins: coins.Sort()}
	genAccount := authtypes.NewBaseAccount(accAddr, nil, 0, 0)

	// TODO: Make the SDK make it far cleaner to add an account to GenesisState
	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	genDoc.InitialHeight = int64(forkHeight)

	authGenState := authtypes.GetGenesisStateFromAppState(util.Cdc, appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return fmt.Errorf("failed to get accounts from any: %w", err)
	}

	if accs.Contains(accAddr) {
		return fmt.Errorf("failed to add account to genesis state; account already exists: %s", accAddr)
	}

	// Add the new account to the set of genesis accounts and sanitize the
	// accounts afterwards.
	accs = append(accs, genAccount)
	accs = authtypes.SanitizeGenesisAccounts(accs)

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return fmt.Errorf("failed to convert accounts into any's: %w", err)
	}

	authGenState.Accounts = genAccs

	authGenStateBz, err := util.Cdc.MarshalJSON(&authGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	appState[authtypes.ModuleName] = authGenStateBz

	bankGenState := banktypes.GetGenesisStateFromAppState(util.Cdc, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenStateBz, err := util.Cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}

	appState[banktypes.ModuleName] = bankGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}

	genDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(genDoc, genFile)
}

func updateModuleGenesis[V proto.Message](appGenState map[string]json.RawMessage, moduleName string, protoVal V, updateGenesis func(V)) error {
	if err := util.Cdc.UnmarshalJSON(appGenState[moduleName], protoVal); err != nil {
		return err
	}
	updateGenesis(protoVal)
	newGenState := protoVal

	bz, err := util.Cdc.MarshalJSON(newGenState)
	if err != nil {
		return err
	}
	appGenState[moduleName] = bz
	return nil
}

func initGenesis(chain *internalChain, votingPeriod, expeditedVotingPeriod time.Duration, forkHeight int) error {
	// initialize a genesis file
	configDir := chain.nodes[0].configDir()
	for _, val := range chain.nodes {
		if chain.chainMeta.Id == ChainAID {
			if err := addAccount(configDir, "", InitBalanceStrA+","+StridePoolBalances, val.keyInfo.GetAddress(), forkHeight); err != nil {
				return err
			}
		} else if chain.chainMeta.Id == ChainBID {
			if err := addAccount(configDir, "", InitBalanceStrB+","+StridePoolBalances, val.keyInfo.GetAddress(), forkHeight); err != nil {
				return err
			}
		}
	}

	// copy the genesis file to the remaining validators
	for _, val := range chain.nodes[1:] {
		_, err := util.CopyFile(
			filepath.Join(configDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		if err != nil {
			return err
		}
	}

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(chain.nodes[0].configDir())
	config.Moniker = chain.nodes[0].moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, staketypes.ModuleName, &staketypes.GenesisState{}, updateStakeGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, poolitypes.ModuleName, &poolitypes.GenesisState{}, updatePoolIncentiveGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, incentivestypes.ModuleName, &incentivestypes.GenesisState{}, updateIncentivesGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, minttypes.ModuleName, &minttypes.GenesisState{}, updateMintGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, txfeestypes.ModuleName, &txfeestypes.GenesisState{}, updateTxfeesGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, gammtypes.ModuleName, &gammtypes.GenesisState{}, updateGammGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, banktypes.ModuleName, &banktypes.GenesisState{}, updateBankGenesis(appGenState))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, poolmanagertypes.ModuleName, &poolmanagertypes.GenesisState{}, updatePoolManagerGenesis(appGenState))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, epochtypes.ModuleName, &epochtypes.GenesisState{}, updateEpochGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, twaptypes.ModuleName, &twaptypes.GenesisState{}, updateTWAPGenesis(appGenState))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, crisistypes.ModuleName, &crisistypes.GenesisState{}, updateCrisisGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, govtypes.ModuleName, &govtypes.GenesisState{}, updateGovGenesis(votingPeriod, expeditedVotingPeriod))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, genutiltypes.ModuleName, &genutiltypes.GenesisState{}, updateGenUtilGenesis(chain))
	if err != nil {
		return err
	}

	bz, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc.AppState = bz

	genesisJson, err := tmjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}

	// write the updated genesis file to each validator
	for _, val := range chain.nodes {
		if err := util.WriteFile(filepath.Join(val.configDir(), "config", "genesis.json"), genesisJson); err != nil {
			return err
		}
	}
	return nil
}

func updateBankGenesis(appGenState map[string]json.RawMessage) func(s *banktypes.GenesisState) {
	return func(bankGenState *banktypes.GenesisState) {
		strideMigrationDenoms := []string{StOsmoDenom, JunoDenom, StJunoDenom, StarsDenom, StStarsDenom}
		denomsToRegister := append([]string{StakeDenom, IonDenom, OsmoDenom, AtomDenom, LuncIBCDenom, UstIBCDenom}, strideMigrationDenoms...)
		for _, denom := range denomsToRegister {
			setDenomMetadata(bankGenState, denom)
		}

		// Update pool balances with initial liquidity.
		gammGenState := &gammtypes.GenesisState{}
		util.Cdc.MustUnmarshalJSON(appGenState[gammtypes.ModuleName], gammGenState)

		for _, poolAny := range gammGenState.Pools {
			poolBytes := poolAny.GetValue()

			var balancerPool balancer.Pool
			util.Cdc.MustUnmarshal(poolBytes, &balancerPool)

			coins := sdk.NewCoins()
			for _, asset := range balancerPool.PoolAssets {
				coins = coins.Add(asset.Token)
			}

			coins = coins.Add(balancerPool.TotalShares)

			bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
				Address: balancerPool.Address,
				Coins:   coins,
			})
		}
	}
}

func updateStakeGenesis(stakeGenState *staketypes.GenesisState) {
	stakeGenState.Params = staketypes.Params{
		BondDenom:         OsmoDenom,
		MaxValidators:     100,
		MaxEntries:        7,
		HistoricalEntries: 10000,
		UnbondingTime:     240000000000,
		MinCommissionRate: sdk.ZeroDec(),
	}
}

func updatePoolIncentiveGenesis(pooliGenState *poolitypes.GenesisState) {
	pooliGenState.LockableDurations = []time.Duration{
		time.Second * 120,
		time.Second * 180,
		time.Second * 240,
	}
	pooliGenState.Params = poolitypes.Params{
		MintedDenom: OsmoDenom,
	}
}

func updateIncentivesGenesis(incentivesGenState *incentivestypes.GenesisState) {
	incentivesGenState.LockableDurations = []time.Duration{
		time.Second,
		time.Second * 120,
		time.Second * 180,
		time.Second * 240,
	}
	incentivesGenState.Params = incentivestypes.Params{
		DistrEpochIdentifier: "day",
	}
}

func updateMintGenesis(mintGenState *minttypes.GenesisState) {
	mintGenState.Params.MintDenom = OsmoDenom
	mintGenState.Params.EpochIdentifier = "day"
}

func updateTxfeesGenesis(txfeesGenState *txfeestypes.GenesisState) {
	txfeesGenState.Basedenom = OsmoDenom
	txfeesGenState.Feetokens = []txfeestypes.FeeToken{
		{Denom: E2EFeeToken, PoolID: 1},
	}
}

func updateGammGenesis(gammGenState *gammtypes.GenesisState) {
	gammGenState.Params.PoolCreationFee = tenOsmo
	// setup fee pool, between "e2e_default_fee_token" and "uosmo"
	uosmoFeeTokenPool := setupPool(1, "uosmo", E2EFeeToken)

	gammGenState.Pools = []*types1.Any{uosmoFeeTokenPool}

	for poolId := uint64(2); poolId <= StOSMO_OSMOPoolId; poolId++ {
		var pool *types1.Any
		switch poolId {
		case StOSMO_OSMOPoolId:
			pool = setupPool(StOSMO_OSMOPoolId, StOsmoDenom, OsmoDenom)
		case StJUNO_JUNOPoolId:
			pool = setupPool(StJUNO_JUNOPoolId, StJunoDenom, JunoDenom)
		case StSTARS_STARSPoolId:
			pool = setupPool(StSTARS_STARSPoolId, StStarsDenom, StarsDenom)
		default:
			// repeated dummy pool. We must do this to be able to
			// test the migration all the way up to the largest pool id
			// of StOSMO_OSMOPoolId.
			pool = setupPool(poolId, OsmoDenom, AtomDenom)
		}
		gammGenState.Pools = append(gammGenState.Pools, pool)
	}

	// Note that we set the next pool number as 1 greater than the latest created pool.
	// This is to ensure that migrations are performed correctly.
	gammGenState.NextPoolNumber = StOSMO_OSMOPoolId + 1
}

func updatePoolManagerGenesis(appGenState map[string]json.RawMessage) func(*poolmanagertypes.GenesisState) {
	return func(s *poolmanagertypes.GenesisState) {
		gammGenState := &gammtypes.GenesisState{}
		util.Cdc.MustUnmarshalJSON(appGenState[gammtypes.ModuleName], gammGenState)
		s.NextPoolId = gammGenState.NextPoolNumber
		s.PoolRoutes = make([]poolmanagertypes.ModuleRoute, 0, s.NextPoolId-1)
		for poolId := uint64(1); poolId < s.NextPoolId; poolId++ {
			s.PoolRoutes = append(s.PoolRoutes, poolmanagertypes.ModuleRoute{
				PoolId: poolId,
				// Note: we assume that all pools created are balancer pools.
				// If changes are needed, modify gamm genesis first.
				PoolType: poolmanagertypes.Balancer,
			})
		}
	}
}

func updateEpochGenesis(epochGenState *epochtypes.GenesisState) {
	epochGenState.Epochs = []epochtypes.EpochInfo{
		// override week epochs which are in default integrations, to be 2min
		epochtypes.NewGenesisEpochInfo("week", time.Second*120),
		// override day epochs which are in default integrations, to be 1min
		epochtypes.NewGenesisEpochInfo("day", time.Second*60),
	}
}

func updateTWAPGenesis(appGenState map[string]json.RawMessage) func(twapGenState *twaptypes.GenesisState) {
	return func(twapGenState *twaptypes.GenesisState) {
		gammGenState := &gammtypes.GenesisState{}
		util.Cdc.MustUnmarshalJSON(appGenState[gammtypes.ModuleName], gammGenState)

		// Lower keep period from defaults to allos us to test pruning.
		twapGenState.Params.RecordHistoryKeepPeriod = time.Second * 15

		for _, poolAny := range gammGenState.Pools {
			poolBytes := poolAny.GetValue()

			var balancerPool balancer.Pool
			util.Cdc.MustUnmarshal(poolBytes, &balancerPool)

			denoms := []string{}
			for _, poolAsset := range balancerPool.PoolAssets {
				denoms = append(denoms, poolAsset.Token.Denom)
			}

			denomPairs := twaptypes.GetAllUniqueDenomPairs(denoms)

			for _, denomPair := range denomPairs {
				// sp0 = denom0 quote, denom1 base.
				sp0, err := balancerPool.SpotPrice(sdk.Context{}, denomPair.Denom0, denomPair.Denom1)
				if err != nil {
					panic(err)
				}

				// sp1 = denom0 base, denom1 quote.
				sp1, err := balancerPool.SpotPrice(sdk.Context{}, denomPair.Denom1, denomPair.Denom0)
				if err != nil {
					panic(err)
				}

				twapRecord := twaptypes.TwapRecord{
					PoolId:                      balancerPool.Id,
					Asset0Denom:                 denomPair.Denom0,
					Asset1Denom:                 denomPair.Denom0,
					Height:                      1,
					Time:                        time.Date(2023, 02, 1, 0, 0, 0, 0, time.UTC), // some time in the past.
					P0LastSpotPrice:             sp0,
					P1LastSpotPrice:             sp1,
					P0ArithmeticTwapAccumulator: sdk.ZeroDec(),
					P1ArithmeticTwapAccumulator: sdk.ZeroDec(),
					GeometricTwapAccumulator:    sdk.ZeroDec(),
					LastErrorTime:               time.Time{}, // no previous error
				}
				twapGenState.Twaps = append(twapGenState.Twaps, twapRecord)
			}
		}
	}
}

func updateCrisisGenesis(crisisGenState *crisistypes.GenesisState) {
	crisisGenState.ConstantFee.Denom = OsmoDenom
}

func updateGovGenesis(votingPeriod, expeditedVotingPeriod time.Duration) func(*govtypes.GenesisState) {
	return func(govGenState *govtypes.GenesisState) {
		govGenState.VotingParams.VotingPeriod = votingPeriod
		govGenState.VotingParams.ExpeditedVotingPeriod = expeditedVotingPeriod
		govGenState.DepositParams.MinDeposit = tenOsmo
		govGenState.DepositParams.MinExpeditedDeposit = fiftyOsmo
	}
}

func updateGenUtilGenesis(c *internalChain) func(*genutiltypes.GenesisState) {
	return func(genUtilGenState *genutiltypes.GenesisState) {
		// generate genesis txs
		genTxs := make([]json.RawMessage, 0, len(c.nodes))
		for _, node := range c.nodes {
			if !node.isValidator {
				continue
			}

			stakeAmountCoin := StakeAmountCoinA
			if c.chainMeta.Id != ChainAID {
				stakeAmountCoin = StakeAmountCoinB
			}
			createValmsg, err := node.buildCreateValidatorMsg(stakeAmountCoin)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			signedTx, err := node.signMsg(createValmsg)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			txRaw, err := util.Cdc.MarshalJSON(signedTx)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}
			genTxs = append(genTxs, txRaw)
		}
		genUtilGenState.GenTxs = genTxs
	}
}

func setDenomMetadata(genState *banktypes.GenesisState, denom string) {
	genState.DenomMetadata = append(genState.DenomMetadata, banktypes.Metadata{
		Description: fmt.Sprintf("Registered denom %s for e2e testing", denom),
		Display:     denom,
		Base:        denom,
		Symbol:      denom,
		Name:        denom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: 0,
			},
		},
	})
}

// sets up a pool with 1% fee, equal weights, and given denoms with supply of 100000000000,
// and a given pool id.
func setupPool(poolId uint64, denomA, denomB string) *types1.Any {
	feePoolParams := balancer.NewPoolParams(sdk.MustNewDecFromStr("0.01"), sdk.ZeroDec(), nil)
	feePoolAssets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denomA, sdk.NewInt(100000000000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denomB, sdk.NewInt(100000000000)),
		},
	}
	pool1, err := balancer.NewBalancerPool(poolId, feePoolParams, feePoolAssets, "", time.Unix(0, 0))
	if err != nil {
		panic(err)
	}
	anyPool, err := types1.NewAnyWithValue(&pool1)
	if err != nil {
		panic(err)
	}
	return anyPool
}
