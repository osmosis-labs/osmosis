package testutils

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
)

var _ authenticator.Authenticator = &StatefulAuthenticator{}
var _ authenticator.AuthenticatorData = &StatefulAuthenticatorData{}

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

func (s StatefulAuthenticator) StaticGas() uint64 {
	return 1000
}

func (s StatefulAuthenticator) Initialize(data []byte) (authenticator.Authenticator, error) {
	return s, nil
}

func (s StatefulAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int8, simulate bool) (authenticator.AuthenticatorData, error) {
	// TODO: We probably want the context here. Specifically a read-only cachecontext
	return StatefulAuthenticatorData{Value: s.GetValue(ctx)}, nil
}

func (s StatefulAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) authenticator.AuthenticationResult {
	// TODO: the get should probably happen in the method above and here we should just have this:
	statefulData, ok := authenticationData.(StatefulAuthenticatorData)
	if !ok {
		return authenticator.Rejected("", sdkerrors.Wrap(sdkerrors.ErrInvalidType, "authenticationData is not StatefulAuthenticatorData"))
	}
	//ctx.GasMeter().ConsumeGas(100_000_000, "loads of gas")
	s.SetValue(ctx, statefulData.Value+1)
	return authenticator.Authenticated()

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

func (s StatefulAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) authenticator.ConfirmationResult {
	s.SetValue(ctx, s.GetValue(ctx)+1)
	return authenticator.Confirm()
}
