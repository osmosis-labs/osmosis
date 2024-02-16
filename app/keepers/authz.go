package keepers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/utils"
)

type AuthzKeeperInterface interface {
	Logger(ctx sdk.Context) log.Logger
	DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) ([][]byte, error)
	SaveGrant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error
	DeleteGrant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error
	GetAuthorizations(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) ([]authz.Authorization, error)
	IterateGrants(ctx sdk.Context, handler func(granterAddr sdk.AccAddress, granteeAddr sdk.AccAddress, grant authz.Grant) bool)
	ExportGenesis(ctx sdk.Context) *authz.GenesisState
	InitGenesis(ctx sdk.Context, data *authz.GenesisState)
	Grants(c context.Context, req *authz.QueryGrantsRequest) (*authz.QueryGrantsResponse, error)
	GranterGrants(c context.Context, req *authz.QueryGranterGrantsRequest) (*authz.QueryGranterGrantsResponse, error)
	GranteeGrants(c context.Context, req *authz.QueryGranteeGrantsRequest) (*authz.QueryGranteeGrantsResponse, error)
	Grant(goCtx context.Context, msg *authz.MsgGrant) (*authz.MsgGrantResponse, error)
	Revoke(goCtx context.Context, msg *authz.MsgRevoke) (*authz.MsgRevokeResponse, error)
	Exec(goCtx context.Context, msg *authz.MsgExec) (*authz.MsgExecResponse, error)

	Keeper() authzkeeper.Keeper
}

var _ AuthzKeeperInterface = &KeeperWrapper{}

type KeeperWrapper struct {
	K                    authzkeeper.Keeper
	authenticatorStorage utils.AuthenticatorStorage
	transientStore       *authenticator.TransientStore
}

func NewKeeperWrapper(k authzkeeper.Keeper, authenticatorKeeper utils.AuthenticatorStorage, transientStore *authenticator.TransientStore) *KeeperWrapper {
	return &KeeperWrapper{K: k, authenticatorStorage: authenticatorKeeper, transientStore: transientStore}
}

// Implementing KeeperInterface

func (kw *KeeperWrapper) Keeper() authzkeeper.Keeper {
	return kw.K
}

func (kw *KeeperWrapper) Logger(ctx sdk.Context) log.Logger {
	return kw.K.Logger(ctx)
}

func (kw *KeeperWrapper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) ([][]byte, error) {
	results, err := kw.K.DispatchActions(ctx, grantee, msgs)
	if err != nil {
		return nil, err
	}
	err = utils.ConfirmExecutionWithoutTx(ctx, kw.authenticatorStorage, msgs)
	if err != nil {
		return nil, err
	}

	kw.transientStore.UpdateFrom(ctx)
	return results, err
}

func (kw *KeeperWrapper) SaveGrant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error {
	return kw.K.SaveGrant(ctx, grantee, granter, authorization, expiration)
}

func (kw *KeeperWrapper) DeleteGrant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error {
	return kw.K.DeleteGrant(ctx, grantee, granter, msgType)
}

func (kw *KeeperWrapper) GetAuthorizations(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) ([]authz.Authorization, error) {
	return kw.K.GetAuthorizations(ctx, grantee, granter)
}

func (kw *KeeperWrapper) IterateGrants(ctx sdk.Context, handler func(granterAddr sdk.AccAddress, granteeAddr sdk.AccAddress, grant authz.Grant) bool) {
	kw.K.IterateGrants(ctx, handler)
}

func (kw *KeeperWrapper) ExportGenesis(ctx sdk.Context) *authz.GenesisState {
	return kw.K.ExportGenesis(ctx)
}

func (kw *KeeperWrapper) InitGenesis(ctx sdk.Context, data *authz.GenesisState) {
	kw.K.InitGenesis(ctx, data)
}

func (kw KeeperWrapper) Grants(c context.Context, req *authz.QueryGrantsRequest) (*authz.QueryGrantsResponse, error) {
	return kw.K.Grants(c, req)
}

func (kw KeeperWrapper) GranterGrants(c context.Context, req *authz.QueryGranterGrantsRequest) (*authz.QueryGranterGrantsResponse, error) {
	return kw.K.GranterGrants(c, req)
}

func (kw KeeperWrapper) GranteeGrants(c context.Context, req *authz.QueryGranteeGrantsRequest) (*authz.QueryGranteeGrantsResponse, error) {
	return kw.K.GranteeGrants(c, req)
}

func (kw KeeperWrapper) Grant(goCtx context.Context, msg *authz.MsgGrant) (*authz.MsgGrantResponse, error) {
	return kw.K.Grant(goCtx, msg)
}

func (kw KeeperWrapper) Revoke(goCtx context.Context, msg *authz.MsgRevoke) (*authz.MsgRevokeResponse, error) {
	return kw.K.Revoke(goCtx, msg)
}

func (kw KeeperWrapper) Exec(goCtx context.Context, msg *authz.MsgExec) (*authz.MsgExecResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, err
	}
	msgs, err := msg.GetMessages()
	if err != nil {
		return nil, err
	}
	results, err := kw.DispatchActions(ctx, grantee, msgs)
	if err != nil {
		return nil, err
	}
	return &authz.MsgExecResponse{Results: results}, nil
}

// MODULE FORM HERE

type AppModuleWrapper struct {
	authzmodule.AppModule
	keeperWrapper AuthzKeeperInterface // Replace original Keeper with KeeperWrapper
}

func NewAppModuleWrapper(original authzmodule.AppModule, keeperWrapper AuthzKeeperInterface) AppModuleWrapper {
	return AppModuleWrapper{
		AppModule:     original,
		keeperWrapper: keeperWrapper,
	}
}

func (amw AppModuleWrapper) RegisterServices(cfg module.Configurator) {
	authz.RegisterQueryServer(cfg.QueryServer(), amw.keeperWrapper)
	authz.RegisterMsgServer(cfg.MsgServer(), amw.keeperWrapper)

	m := authzkeeper.NewMigrator(amw.keeperWrapper.Keeper())
	err := cfg.RegisterMigration(authz.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", authz.ModuleName, err))
	}
}

func (amw AppModuleWrapper) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState authz.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	amw.keeperWrapper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

func (amw AppModuleWrapper) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := amw.keeperWrapper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

func (amw AppModuleWrapper) Name() string {
	return amw.AppModule.Name()
}

func (amw AppModuleWrapper) RegisterInvariants(ir sdk.InvariantRegistry) {
	amw.AppModule.RegisterInvariants(ir)
}

func (amw AppModuleWrapper) NewHandler() sdk.Handler {
	return amw.AppModule.NewHandler()
}

func (amw AppModuleWrapper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	amw.AppModule.BeginBlock(ctx, req)
}

func (amw AppModuleWrapper) GenerateGenesisState(simState *module.SimulationState) {
	amw.AppModule.GenerateGenesisState(simState)
}

func (amw AppModuleWrapper) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	amw.AppModule.RegisterStoreDecoder(sdr)
}

func (amw AppModuleWrapper) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return amw.AppModule.WeightedOperations(simState)
}
