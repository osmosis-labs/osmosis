package keepers

import (
	"context"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/tendermint/tendermint/libs/log"
	"time"
)

type AuthzKeeperInterface interface {
	Logger(ctx types.Context) log.Logger
	DispatchActions(ctx types.Context, grantee types.AccAddress, msgs []types.Msg) ([][]byte, error)
	SaveGrant(ctx types.Context, grantee, granter types.AccAddress, authorization authz.Authorization, expiration time.Time) error
	DeleteGrant(ctx types.Context, grantee types.AccAddress, granter types.AccAddress, msgType string) error
	GetAuthorizations(ctx types.Context, grantee types.AccAddress, granter types.AccAddress) []authz.Authorization
	GetCleanAuthorization(ctx types.Context, grantee types.AccAddress, granter types.AccAddress, msgType string) (authz.Authorization, time.Time)
	IterateGrants(ctx types.Context, handler func(granterAddr types.AccAddress, granteeAddr types.AccAddress, grant authz.Grant) bool)
	ExportGenesis(ctx types.Context) *authz.GenesisState
	InitGenesis(ctx types.Context, data *authz.GenesisState)
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
	K authzkeeper.Keeper
}

func NewKeeperWrapper(k authzkeeper.Keeper) *KeeperWrapper {
	return &KeeperWrapper{K: k}
}

// Implementing KeeperInterface

func (kw *KeeperWrapper) Keeper() authzkeeper.Keeper {
	return kw.K
}

func (kw *KeeperWrapper) Logger(ctx types.Context) log.Logger {
	return kw.K.Logger(ctx)
}

func (kw *KeeperWrapper) DispatchActions(ctx types.Context, grantee types.AccAddress, msgs []types.Msg) ([][]byte, error) {
	// TODO: Here we could ensure that authenticators are properly called.
	//       (if we don't want to make authz grants an authenticator)
	return kw.K.DispatchActions(ctx, grantee, msgs)
}

func (kw *KeeperWrapper) SaveGrant(ctx types.Context, grantee, granter types.AccAddress, authorization authz.Authorization, expiration time.Time) error {
	return kw.K.SaveGrant(ctx, grantee, granter, authorization, expiration)
}

func (kw *KeeperWrapper) DeleteGrant(ctx types.Context, grantee types.AccAddress, granter types.AccAddress, msgType string) error {
	return kw.K.DeleteGrant(ctx, grantee, granter, msgType)
}

func (kw *KeeperWrapper) GetAuthorizations(ctx types.Context, grantee types.AccAddress, granter types.AccAddress) []authz.Authorization {
	return kw.K.GetAuthorizations(ctx, grantee, granter)
}

func (kw *KeeperWrapper) GetCleanAuthorization(ctx types.Context, grantee types.AccAddress, granter types.AccAddress, msgType string) (authz.Authorization, time.Time) {
	return kw.K.GetCleanAuthorization(ctx, grantee, granter, msgType)
}

func (kw *KeeperWrapper) IterateGrants(ctx types.Context, handler func(granterAddr types.AccAddress, granteeAddr types.AccAddress, grant authz.Grant) bool) {
	kw.K.IterateGrants(ctx, handler)
}

func (kw *KeeperWrapper) ExportGenesis(ctx types.Context) *authz.GenesisState {
	return kw.K.ExportGenesis(ctx)
}

func (kw *KeeperWrapper) InitGenesis(ctx types.Context, data *authz.GenesisState) {
	kw.K.InitGenesis(ctx, data)
}

func (kw *KeeperWrapper) Grants(c context.Context, req *authz.QueryGrantsRequest) (*authz.QueryGrantsResponse, error) {
	return kw.K.Grants(c, req)
}

func (kw *KeeperWrapper) GranterGrants(c context.Context, req *authz.QueryGranterGrantsRequest) (*authz.QueryGranterGrantsResponse, error) {
	return kw.K.GranterGrants(c, req)
}

func (kw *KeeperWrapper) GranteeGrants(c context.Context, req *authz.QueryGranteeGrantsRequest) (*authz.QueryGranteeGrantsResponse, error) {
	return kw.K.GranteeGrants(c, req)
}

func (kw *KeeperWrapper) Grant(goCtx context.Context, msg *authz.MsgGrant) (*authz.MsgGrantResponse, error) {
	return kw.K.Grant(goCtx, msg)
}

func (kw *KeeperWrapper) Revoke(goCtx context.Context, msg *authz.MsgRevoke) (*authz.MsgRevokeResponse, error) {
	return kw.K.Revoke(goCtx, msg)
}

func (kw *KeeperWrapper) Exec(goCtx context.Context, msg *authz.MsgExec) (*authz.MsgExecResponse, error) {
	return kw.K.Exec(goCtx, msg)
}
