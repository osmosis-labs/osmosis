package testutils

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
)

var _ authenticator.Authenticator = &StatefulAuthenticator{}

type StatefulAuthenticatorData struct {
	Value int
}

// StatefulAuthenticator is an experiment of how to write authenticators that handle state
type StatefulAuthenticator struct {
	KvStoreKey storetypes.StoreKey
}

func (s StatefulAuthenticator) Type() string {
	return "Stateful"
}

func (s StatefulAuthenticator) StaticGas() uint64 {
	return 1000
}

func (s StatefulAuthenticator) Initialize(config []byte) (authenticator.Authenticator, error) {
	return s, nil
}

func (s StatefulAuthenticator) Authenticate(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	statefulData := StatefulAuthenticatorData{Value: s.GetValue(ctx)}
	if statefulData.Value > 10 {
		return fmt.Errorf("Value is too high: %d", statefulData.Value)
	}
	s.SetValue(ctx, statefulData.Value+1)
	return nil
}

func (s StatefulAuthenticator) Track(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	statefulData := StatefulAuthenticatorData{Value: s.GetValue(ctx)}
	if statefulData.Value > 10 {
		return fmt.Errorf("Value is too high: %d", statefulData.Value)
	}
	s.SetValue(ctx, statefulData.Value+1)
	return nil
}

func (s StatefulAuthenticator) SetValue(ctx sdk.Context, value int) {
	kvStore := prefix.NewStore(ctx.KVStore(s.KvStoreKey), []byte(s.Type()))
	statefulData := StatefulAuthenticatorData{Value: value}
	newBz, _ := json.Marshal(statefulData)
	kvStore.Set([]byte("value"), newBz)
}

func (s StatefulAuthenticator) GetValue(ctx sdk.Context) int {
	kvStore := prefix.NewStore(ctx.KVStore(s.KvStoreKey), []byte(s.Type()))
	bz := kvStore.Get([]byte("value")) // global value. On the real thing we may want the account
	var statefulData StatefulAuthenticatorData
	_ = json.Unmarshal(bz, &statefulData) // if we can't unmarshal, we just assume it's 0
	return statefulData.Value
}

func (s StatefulAuthenticator) ConfirmExecution(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	s.SetValue(ctx, s.GetValue(ctx)+1)
	return nil
}

func (s StatefulAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return nil
}

func (s StatefulAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return nil
}
