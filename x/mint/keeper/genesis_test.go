package keeper_test

import (
	"testing"

	store "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	epochtypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"

	simapp "github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/app/params"

	"github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
)

const (
	mintSubspace = "mint"
)

var (
	mintStoreKeys = sdk.NewKVStoreKeys([]string{authtypes.StoreKey, banktypes.StoreKey, distrtypes.StoreKey, epochtypes.StoreKey, types.StoreKey}...)
)

func TestMintInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	validateGenesis := types.ValidateGenesis(*types.DefaultGenesisState())
	require.NoError(t, validateGenesis)

	developerAccount := app.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	initialVestingCoins := app.BankKeeper.GetBalance(ctx, developerAccount, sdk.DefaultBondDenom)

	expectedVestingCoins, ok := sdk.NewIntFromString("225000000000000")
	require.True(t, ok)
	require.Equal(t, expectedVestingCoins, initialVestingCoins.Amount)
	require.Equal(t, int64(0), app.MintKeeper.GetLastHalvenEpochNum(ctx))
}

// TestMintExportGenesis exports genesis and attempts to validate it
func TestMintImportExportGenesis(t *testing.T) {
	authStoreKey, ok := mintStoreKeys[authtypes.StoreKey]
	require.True(t, ok)

	accountKeeper := newAccountKeeper(t, authStoreKey)

	bankStoreKey, ok := mintStoreKeys[banktypes.StoreKey]
	require.True(t, ok)

	bankKeeper := newBankKeeper(t, bankStoreKey, accountKeeper)

	mintStoreKey, ok := mintStoreKeys[types.StoreKey]
	require.True(t, ok)

	mintKeeper := newMintKeeper(t, mintStoreKey, accountKeeper, bankKeeper)
	ctx := newContext(t)

	// To easily initialize other keepers
	app := simapp.SetupNoInitChain()

	expectedGenesis := types.DefaultGenesisState()

	// Reset genesis to default
	mintKeeper.InitGenesis(ctx, accountKeeper, app.BankKeeper, expectedGenesis)

	exported := app.MintKeeper.ExportGenesis(ctx)
	require.NotNil(t, exported)

	require.Equal(t, *expectedGenesis, *exported)
}

func newAccountKeeper(t *testing.T, authStoreKey *sdk.KVStoreKey) authkeeper.AccountKeeper {
	encodingCfg := params.MakeEncodingConfig()
	cdc := encodingCfg.Marshaler

	paramSpace := paramtypes.NewSubspace(cdc, encodingCfg.Amino, authStoreKey, authStoreKey, "auth")

	accountKeeper := authkeeper.NewAccountKeeper(
		encodingCfg.Marshaler,
		authStoreKey,
		paramSpace,
		authtypes.ProtoBaseAccount,
		simapp.ModuleAccountPermissions,
	)
	return accountKeeper
}

func newBankKeeper(t *testing.T, bankStoreKey *sdk.KVStoreKey, accountKeeper authkeeper.AccountKeeper) bankkeeper.BaseKeeper {
	encodingCfg := params.MakeEncodingConfig()
	cdc := encodingCfg.Marshaler

	paramSpace := paramtypes.NewSubspace(cdc, encodingCfg.Amino, bankStoreKey, bankStoreKey, "bank")

	bankKeeper := bankkeeper.NewBaseKeeper(
		encodingCfg.Marshaler,
		bankStoreKey,
		accountKeeper,
		paramSpace,
		map[string]bool{},
	)
	return bankKeeper
}

func newMintKeeper(t *testing.T, mintStoreKey *sdk.KVStoreKey, accountKeeper authkeeper.AccountKeeper, bankKeeper bankkeeper.BaseKeeper) keeper.Keeper {
	encodingCfg := params.MakeEncodingConfig()
	cdc := encodingCfg.Marshaler

	paramSpace := paramtypes.NewSubspace(cdc, encodingCfg.Amino, mintStoreKey, mintStoreKey, "mint")

	// To easily initialize other keepers
	app := simapp.SetupNoInitChain()

	mintKeeper := keeper.NewKeeper(
		encodingCfg.Marshaler,
		mintStoreKey,
		paramSpace,
		accountKeeper,
		bankKeeper,
		app.DistrKeeper,
		app.EpochsKeeper,
		authtypes.FeeCollectorName,
	)
	return mintKeeper
}

func newContext(t *testing.T) sdk.Context {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	cms := store.NewCommitMultiStore(db, logger)

	for _, key := range mintStoreKeys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	}

	require.NoError(t, cms.LoadLatestVersion())

	return sdk.NewContext(cms, tmproto.Header{}, false, logger)
}
