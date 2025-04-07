package keeper_test

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/stretchr/testify/suite"

	cdcutil "github.com/cosmos/cosmos-sdk/codec/testutil"

	epochskeeper "github.com/osmosis-labs/osmosis/v27/x/epochs/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/epochs/types"
)

type KeeperTestSuite struct {
	suite.Suite
	Ctx          sdk.Context
	EpochsKeeper *epochskeeper.Keeper
	queryClient  types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	ctx, epochsKeeper := Setup()
	s.Ctx = ctx
	s.EpochsKeeper = epochsKeeper
	queryRouter := baseapp.NewGRPCQueryRouter()
	cfg := module.NewConfigurator(nil, nil, queryRouter)
	types.RegisterQueryServer(cfg.QueryServer(), epochskeeper.NewQuerier(*s.EpochsKeeper))
	grpcQueryService := &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: queryRouter,
		Ctx:             s.Ctx,
	}
	interfaceRegistry := cdcutil.CodecOptions{AccAddressPrefix: "symphony", ValAddressPrefix: "symphonyvaloper"}.NewInterfaceRegistry()
	grpcQueryService.SetInterfaceRegistry(interfaceRegistry)
	s.queryClient = types.NewQueryClient(grpcQueryService)

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func Setup() (sdk.Context, *epochskeeper.Keeper) {
	epochsStoreKey := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContext(epochsStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	epochsKeeper := epochskeeper.NewKeeper(epochsStoreKey)
	epochsKeeper = epochsKeeper.SetHooks(types.NewMultiEpochHooks())
	ctx.WithBlockHeight(1).WithChainID("osmosis-1").WithBlockTime(time.Now().UTC())
	epochsKeeper.InitGenesis(ctx, *types.DefaultGenesis())
	SetEpochStartTime(ctx, epochsKeeper)
	return ctx, epochsKeeper
}

func SetEpochStartTime(ctx sdk.Context, epochsKeeper *epochskeeper.Keeper) {
	for _, epoch := range epochsKeeper.AllEpochInfos(ctx) {
		epoch.StartTime = ctx.BlockTime()
		epochsKeeper.DeleteEpochInfo(ctx, epoch.Identifier)
		err := epochsKeeper.AddEpochInfo(ctx, epoch)
		if err != nil {
			panic(err)
		}
	}
}
