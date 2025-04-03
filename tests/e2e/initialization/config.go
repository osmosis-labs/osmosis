package initialization

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	poolitypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v27/x/protorev/types"
	twaptypes "github.com/osmosis-labs/osmosis/v27/x/twap/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"

	types1 "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/osmosis-labs/osmosis/v27/tests/e2e/util"
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
	MelodyDenom         = "note"
	IonDenom            = "uion"
	StakeDenom          = "stake"
	AtomDenom           = "uatom"
	DaiDenom            = "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"
	MelodyIBCDenom      = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	StakeIBCDenom       = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B7787"
	E2EFeeToken         = "e2e-default-feetoken"
	UstIBCDenom         = "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC"
	LuncIBCDenom        = "ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0"
	MinGasPrice         = "0.000"
	IbcSendAmount       = 3300000000
	ValidatorWalletName = "val"
	// chainA
	ChainAID       = "melody-test-a"
	MelodyBalanceA = 20000000000000
	IonBalanceA    = 100000000000
	StakeBalanceA  = 110000000000
	StakeAmountA   = 100000000000
	UstBalanceA    = 500000000000000
	LuncBalanceA   = 500000000000000
	DaiBalanceA    = "100000000000000000000000"
	// chainB
	ChainBID          = "melody-test-b"
	MelodyBalanceB    = 500000000000
	IonBalanceB       = 100000000000
	StakeBalanceB     = 440000000000
	StakeAmountB      = 400000000000
	GenesisFeeBalance = 100000000000
	WalletFeeBalance  = 100000000

	EpochDayDuration      = time.Second * 60
	EpochWeekDuration     = time.Second * 120
	TWAPPruningKeepPeriod = EpochDayDuration / 4

	DaiMelodyPoolId = 674
)

var (
	StakeAmountIntA  = osmomath.NewInt(StakeAmountA)
	StakeAmountCoinA = sdk.NewCoin(MelodyDenom, StakeAmountIntA)
	StakeAmountIntB  = osmomath.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(MelodyDenom, StakeAmountIntB)

	DaiMelodyPoolBalances = fmt.Sprintf("%s%s", DaiBalanceA, DaiDenom)

	InitBalanceStrA = fmt.Sprintf("%d%s,%d%s,%d%s,%d%s,%d%s", MelodyBalanceA, MelodyDenom, StakeBalanceA, StakeDenom, IonBalanceA, IonDenom, UstBalanceA, UstIBCDenom, LuncBalanceA, LuncIBCDenom)
	InitBalanceStrB = fmt.Sprintf("%d%s,%d%s,%d%s", MelodyBalanceB, MelodyDenom, StakeBalanceB, StakeDenom, IonBalanceB, IonDenom)
	MelodyToken     = sdk.NewInt64Coin(MelodyDenom, IbcSendAmount) // 3,300note
	StakeToken      = sdk.NewInt64Coin(StakeDenom, IbcSendAmount)  // 3,300ustake
	tenMelody       = sdk.Coins{sdk.NewInt64Coin(MelodyDenom, 10_000_000)}
	fiftyMelody     = sdk.Coins{sdk.NewInt64Coin(MelodyDenom, 50_000_000)}
	WalletFeeTokens = sdk.NewCoin(E2EFeeToken, osmomath.NewInt(WalletFeeBalance))
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
	coins = coins.Add(sdk.NewCoin(E2EFeeToken, osmomath.NewInt(GenesisFeeBalance)))

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
		addr, err := val.keyInfo.GetAddress()
		if err != nil {
			return err
		}
		if chain.chainMeta.Id == ChainAID {
			if err := addAccount(configDir, "", InitBalanceStrA+","+DaiMelodyPoolBalances, addr, forkHeight); err != nil {
				return err
			}
		} else if chain.chainMeta.Id == ChainBID {
			if err := addAccount(configDir, "", InitBalanceStrB+","+DaiMelodyPoolBalances, addr, forkHeight); err != nil {
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

	err = updateModuleGenesis(appGenState, govtypes.ModuleName, &govtypesv1.GenesisState{}, updateGovGenesis(votingPeriod, expeditedVotingPeriod))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, genutiltypes.ModuleName, &genutiltypes.GenesisState{}, updateGenUtilGenesis(chain))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, protorevtypes.ModuleName, &protorevtypes.GenesisState{}, updateProtorevGenesis)
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
		denomsToRegister := []string{StakeDenom, IonDenom, MelodyDenom, AtomDenom, LuncIBCDenom, UstIBCDenom, DaiDenom}
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
		BondDenom:         MelodyDenom,
		MaxValidators:     100,
		MaxEntries:        7,
		HistoricalEntries: 10000,
		UnbondingTime:     240000000000,
		MinCommissionRate: osmomath.ZeroDec(),
	}
}

func updatePoolIncentiveGenesis(pooliGenState *poolitypes.GenesisState) {
	pooliGenState.LockableDurations = []time.Duration{
		time.Second * 60,
		time.Second * 120,
		time.Second * 240,
	}
	pooliGenState.Params = poolitypes.Params{
		MintedDenom: MelodyDenom,
	}
}

func updateIncentivesGenesis(incentivesGenState *incentivestypes.GenesisState) {
	incentivesGenState.LockableDurations = []time.Duration{
		time.Second * 60,
		time.Second * 120,
		time.Second * 240,
	}
	incentivesGenState = incentivestypes.DefaultGenesis()
	incentivesGenState.Params.DistrEpochIdentifier = "day"
}

func updateMintGenesis(mintGenState *minttypes.GenesisState) {
	mintGenState.Params.MintDenom = MelodyDenom
	mintGenState.Params.EpochIdentifier = "day"
}

func updateTxfeesGenesis(txfeesGenState *txfeestypes.GenesisState) {
	txfeesGenState.Basedenom = MelodyDenom
	txfeesGenState.Feetokens = []txfeestypes.FeeToken{
		{Denom: E2EFeeToken, PoolID: 1},
	}
}

func updateGammGenesis(gammGenState *gammtypes.GenesisState) {
	gammGenState.Params.PoolCreationFee = tenMelody
	// setup fee pool, between "e2e_default_fee_token" and "note"
	noteFeeTokenPool := setupPool(1, "note", E2EFeeToken)

	gammGenState.Pools = []*types1.Any{noteFeeTokenPool}

	// Notice that this is non-inclusive. The DAI/OSMO pool should be created in the
	// pre-upgrade logic of the upgrade configurer.
	for poolId := uint64(2); poolId < DaiMelodyPoolId; poolId++ {
		gammGenState.Pools = append(gammGenState.Pools, setupPool(poolId, MelodyDenom, AtomDenom))
	}

	// Note that we set the next pool number as 1 greater than the latest created pool.
	// This is to ensure that migrations are performed correctly.
	gammGenState.NextPoolNumber = DaiMelodyPoolId
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
		// override week epochs which are in default integrations, to be 60 seconds
		epochtypes.NewGenesisEpochInfo("week", time.Second*60),
		// override day epochs which are in default integrations, to be 5 seconds
		epochtypes.NewGenesisEpochInfo("day", time.Second*5),
	}
}

func updateTWAPGenesis(appGenState map[string]json.RawMessage) func(twapGenState *twaptypes.GenesisState) {
	return func(twapGenState *twaptypes.GenesisState) {
		gammGenState := &gammtypes.GenesisState{}
		util.Cdc.MustUnmarshalJSON(appGenState[gammtypes.ModuleName], gammGenState)

		// Lower keep period from defaults to allows us to test pruning.
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
					PoolId:      balancerPool.Id,
					Asset0Denom: denomPair.Denom0,
					Asset1Denom: denomPair.Denom0,
					Height:      1,
					Time:        time.Date(2023, 0o2, 1, 0, 0, 0, 0, time.UTC), // some time in the past.
					// Note: truncation is acceptable as x/twap is guaranteed to work only on pools with spot prices > 10^-18.
					P0LastSpotPrice:             sp0.Dec(),
					P1LastSpotPrice:             sp1.Dec(),
					P0ArithmeticTwapAccumulator: osmomath.ZeroDec(),
					P1ArithmeticTwapAccumulator: osmomath.ZeroDec(),
					GeometricTwapAccumulator:    osmomath.ZeroDec(),
					LastErrorTime:               time.Time{}, // no previous error
				}
				twapGenState.Twaps = append(twapGenState.Twaps, twapRecord)
			}
		}
	}
}

func updateCrisisGenesis(crisisGenState *crisistypes.GenesisState) {
	crisisGenState.ConstantFee.Denom = MelodyDenom
}

//nolint:unparam
func updateGovGenesis(votingPeriod, expeditedVotingPeriod time.Duration) func(*govtypesv1.GenesisState) {
	return func(govGenState *govtypesv1.GenesisState) {
		govGenState.Params.VotingPeriod = &votingPeriod
		govGenState.Params.ExpeditedVotingPeriod = &expeditedVotingPeriod
		govGenState.Params.MinDeposit = tenMelody
		govGenState.Params.ExpeditedMinDeposit = fiftyMelody
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

			const genesisSetupFailed = "genutil genesis setup failed: "
			if err != nil {
				panic(genesisSetupFailed + err.Error())
			}

			signedTx, err := node.signMsg(createValmsg)
			if err != nil {
				panic(genesisSetupFailed + err.Error())
			}

			txRaw, err := util.Cdc.MarshalJSON(signedTx)
			if err != nil {
				panic(genesisSetupFailed + err.Error())
			}
			genTxs = append(genTxs, txRaw)
		}
		genUtilGenState.GenTxs = genTxs
	}
}

func updateProtorevGenesis(protorevGenState *protorevtypes.GenesisState) {
	protorevGenState.DeveloperAddress = "symphony14acx3hq749jm04cmr4qe4hqhlw6va2sv3cg64j"
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
	feePoolParams := balancer.NewPoolParams(osmomath.MustNewDecFromStr("0.01"), osmomath.ZeroDec(), nil)
	feePoolAssets := []balancer.PoolAsset{
		{
			Weight: osmomath.NewInt(100),
			Token:  sdk.NewCoin(denomA, osmomath.NewInt(100000000000)),
		},
		{
			Weight: osmomath.NewInt(100),
			Token:  sdk.NewCoin(denomB, osmomath.NewInt(100000000000)),
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
