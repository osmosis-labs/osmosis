package testutils

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
)

var _ authenticator.Authenticator = &SpyAuthenticator{}

type SpyTrackRequest struct {
	AuthenticatorId string                 `json:"authenticator_id"`
	Account         sdk.AccAddress         `json:"account"`
	Msg             authenticator.LocalAny `json:"msg"`
	MsgIndex        uint64                 `json:"msg_index"`
}

type SpyAddRequest struct {
	Account         sdk.AccAddress `json:"account"`
	Data            []byte         `json:"data"`
	AuthenticatorId string         `json:"authenticator_id"`
}

type SpyRemoveRequest struct {
	Account         sdk.AccAddress `json:"account"`
	Data            []byte         `json:"data"`
	AuthenticatorId string         `json:"authenticator_id"`
}

type LatestCalls struct {
	Authenticate           authenticator.AuthenticationRequest
	Track                  SpyTrackRequest
	ConfirmExecution       authenticator.AuthenticationRequest
	OnAuthenticatorAdded   SpyAddRequest
	OnAuthenticatorRemoved SpyRemoveRequest
}

type FailureFlag = uint8

const (
	AUTHENTICATE_FAIL FailureFlag = 1 << iota
	CONFIRM_EXECUTION_FAIL
)

func Has(fixed, toCheck FailureFlag) bool {
	return fixed&toCheck != 0
}

type SpyAuthenticatorData struct {
	Name    string      `json:"name"`
	Failure FailureFlag `json:"failure"` // bit flag representing authenticator failure
}

// SpyAuthenticator tracks latest call and can be used to test the authenticator
type SpyAuthenticator struct {
	KvStoreKey storetypes.StoreKey
	Name       string
	Failure    FailureFlag
}

func NewSpyAuthenticator(kvStoreKey storetypes.StoreKey) SpyAuthenticator {
	return SpyAuthenticator{KvStoreKey: kvStoreKey}
}

func (s SpyAuthenticator) Type() string {
	return "Spy"
}

func (s SpyAuthenticator) StaticGas() uint64 {
	return 1000
}

func (s SpyAuthenticator) Initialize(config []byte) (authenticator.Authenticator, error) {
	var spyData SpyAuthenticatorData
	err := json.Unmarshal(config, &spyData)
	if err != nil {
		return nil, err
	}
	s.Name = spyData.Name
	s.Failure = spyData.Failure
	return s, nil
}

func (s SpyAuthenticator) Authenticate(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	s.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.Authenticate = request
		return calls
	})

	if Has(s.Failure, AUTHENTICATE_FAIL) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "not authenticated")
	}
	return nil
}

func (s SpyAuthenticator) Track(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	s.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.Track = SpyTrackRequest{
			AuthenticatorId: request.AuthenticatorId,
			Account:         request.Account,
			Msg:             request.Msg,
			MsgIndex:        request.MsgIndex,
		}
		return calls
	})
	return nil
}

func (s SpyAuthenticator) ConfirmExecution(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	// intentionlly call update before check to test state revert
	s.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.ConfirmExecution = request
		return calls
	})

	if Has(s.Failure, CONFIRM_EXECUTION_FAIL) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "not authenticated")
	}
	return nil
}

func (s SpyAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	spy, err := s.Initialize(config)
	if err != nil {
		return err
	}
	spyAuth, ok := spy.(SpyAuthenticator)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to cast authenticator to SpyAuthenticator")
	}

	spyAuth.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.OnAuthenticatorAdded = SpyAddRequest{
			Account:         account,
			Data:            config,
			AuthenticatorId: authenticatorId,
		}
		return calls
	})
	return nil
}

func (s SpyAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	spy, err := s.Initialize(config)
	if err != nil {
		return err
	}
	spyAuth, ok := spy.(SpyAuthenticator)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to cast authenticator to SpyAuthenticator")
	}

	spyAuth.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.OnAuthenticatorRemoved = SpyRemoveRequest{
			Account:         account,
			Data:            config,
			AuthenticatorId: authenticatorId,
		}
		return calls
	})
	return nil
}

func (s SpyAuthenticator) GetLatestCalls(ctx sdk.Context) LatestCalls {
	if s.Name == "" {
		panic("SpyAuthenticator is not initialized")
	}
	kvStore := s.storeByName(ctx)
	bz := kvStore.Get([]byte("calls"))
	var calls LatestCalls
	_ = json.Unmarshal(bz, &calls)
	return calls
}

func (s SpyAuthenticator) UpdateLatestCalls(ctx sdk.Context, f func(calls LatestCalls) LatestCalls) LatestCalls {
	if s.Name == "" {
		panic("SpyAuthenticator is not initialized")
	}

	kvStore := s.storeByName(ctx)
	bz := kvStore.Get([]byte("calls"))
	var calls LatestCalls

	_ = json.Unmarshal(bz, &calls)

	calls = f(calls)
	newBz, _ := json.Marshal(calls)
	kvStore.Set([]byte("calls"), newBz)

	return calls
}

func (s SpyAuthenticator) ResetLatestCalls(ctx sdk.Context) {
	s.store(ctx).Delete([]byte("calls"))
}

func (s SpyAuthenticator) store(ctx sdk.Context) storetypes.KVStore {
	return prefix.NewStore(ctx.KVStore(s.KvStoreKey), []byte(s.Type()))
}

func (s SpyAuthenticator) storeByName(ctx sdk.Context) storetypes.KVStore {
	return prefix.NewStore(ctx.KVStore(s.KvStoreKey), []byte(s.Name))
}
