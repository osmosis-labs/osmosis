package testutils

import (
	"encoding/json"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ iface.Authenticator = &SpyAuthenticator{}

type SpyTrackRequest struct {
	AuthenticatorId string         `json:"authenticator_id"`
	Account         sdk.AccAddress `json:"account"`
	Msg             iface.LocalAny `json:"msg"`
	MsgIndex        uint64         `json:"msg_index"`
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
	Authenticate           iface.AuthenticationRequest
	Track                  SpyTrackRequest
	ConfirmExecution       iface.AuthenticationRequest
	OnAuthenticatorAdded   SpyAddRequest
	OnAuthenticatorRemoved SpyRemoveRequest
}

type SpyAuthenticatorData struct {
	Name string `json:"name"`
}

// SpyAuthenticator tracks latest call and can be used to test the authenticator
type SpyAuthenticator struct {
	KvStoreKey storetypes.StoreKey
	Name       string
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

func (s SpyAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	var spyData SpyAuthenticatorData
	err := json.Unmarshal(data, &spyData)
	if err != nil {
		return nil, err
	}
	s.Name = spyData.Name
	return s, nil
}

func (s SpyAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) error {
	s.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.Authenticate = request
		return calls
	})
	return nil
}

func (s SpyAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, feePayer sdk.AccAddress, msg sdk.Msg, msgIndex uint64,
	authenticatorId string,
) error {
	encodedMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to encode msg")
	}

	s.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.Track = SpyTrackRequest{
			AuthenticatorId: authenticatorId,
			Account:         account,
			Msg:             iface.LocalAny{TypeURL: encodedMsg.TypeUrl, Value: encodedMsg.Value},
			MsgIndex:        msgIndex,
		}
		return calls
	})
	return nil
}

func (s SpyAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) error {
	s.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.ConfirmExecution = request
		return calls
	})
	return nil
}

func (s SpyAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	spy, err := s.Initialize(data)
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
			Data:            data,
			AuthenticatorId: authenticatorId,
		}
		return calls
	})
	return nil
}

func (s SpyAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	s.UpdateLatestCalls(ctx, func(calls LatestCalls) LatestCalls {
		calls.OnAuthenticatorRemoved = SpyRemoveRequest{
			Account:         account,
			Data:            data,
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
	kvStore := prefix.NewStore(prefix.NewStore(ctx.KVStore(s.KvStoreKey), []byte(s.Type())), []byte(s.Name))
	bz := kvStore.Get([]byte("calls"))
	var calls LatestCalls
	_ = json.Unmarshal(bz, &calls)
	return calls
}

func (s SpyAuthenticator) UpdateLatestCalls(ctx sdk.Context, f func(calls LatestCalls) LatestCalls) LatestCalls {
	if s.Name == "" {
		panic("SpyAuthenticator is not initialized")
	}

	kvStore := prefix.NewStore(prefix.NewStore(ctx.KVStore(s.KvStoreKey), []byte(s.Type())), []byte(s.Name))
	bz := kvStore.Get([]byte("calls"))
	var calls LatestCalls

	_ = json.Unmarshal(bz, &calls)

	calls = f(calls)
	newBz, _ := json.Marshal(calls)
	kvStore.Set([]byte("calls"), newBz)

	return calls
}

func (s SpyAuthenticator) ResetLatestCalls(ctx sdk.Context) {
	kvStore := prefix.NewStore(ctx.KVStore(s.KvStoreKey), []byte(s.Type()))
	kvStore.Delete([]byte("calls"))
}
