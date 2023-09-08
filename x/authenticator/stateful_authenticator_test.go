package authenticator_test

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
)

var _ types.Authenticator = &StatefulAuthenticator{}
var _ types.AuthenticatorData = &StatefulAuthenticatorData{}

type StatefulAuthenticatorData struct {
	Value int
}

// StatefulAuthenticator is an experiment of how to write authenticators that handle state
type StatefulAuthenticator struct {
	KvStoreKey sdk.StoreKey
}

func (s StatefulAuthenticator) Type() string {
	return "Stateful"
}

func (s StatefulAuthenticator) Initialize(data []byte) (types.Authenticator, error) {
	return s, nil
}

func (s StatefulAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex uint8, simulate bool) (types.AuthenticatorData, error) {
	// TODO: We probably want the context here. Specifically a read-only cachecontext
	return StatefulAuthenticatorData{Value: 0}, nil
}

func (s StatefulAuthenticator) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData types.AuthenticatorData) (bool, error) {
	// TODO: the get should probably happen in the method above and here we should just have this:
	//statefulData, ok := authenticationData.(StatefulAuthenticatorData)
	//if !ok {
	//	return false, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "authenticationData is not StatefulAuthenticatorData")
	ctx.GasMeter().ConsumeGas(100_000_000, "loads of gas")
	statefulData := StatefulAuthenticatorData{Value: s.GetValue(ctx) + 1}
	s.SetValue(ctx, statefulData.Value)
	return true, nil
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

func (s StatefulAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData types.AuthenticatorData) bool {
	s.SetValue(ctx, s.GetValue(ctx)+1)
	return true
}
