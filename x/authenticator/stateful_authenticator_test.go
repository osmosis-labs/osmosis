package authenticator_test

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

func (s StatefulAuthenticator) Gas() uint64 {
	return 1000
}

func (s StatefulAuthenticator) Initialize(data []byte) (authenticator.Authenticator, error) {
	return s, nil
}

func (s StatefulAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int8, simulate bool) (authenticator.AuthenticatorData, error) {
	// TODO: We probably want the context here. Specifically a read-only cachecontext
	return StatefulAuthenticatorData{Value: s.GetValue(ctx)}, nil
}

func (s StatefulAuthenticator) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) (bool, error) {
	// TODO: the get should probably happen in the method above and here we should just have this:
	statefulData, ok := authenticationData.(StatefulAuthenticatorData)
	if !ok {
		return false, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "authenticationData is not StatefulAuthenticatorData")
	}
	//ctx.GasMeter().ConsumeGas(100_000_000, "loads of gas")
	s.SetValue(ctx, statefulData.Value+1)
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

func (s StatefulAuthenticator) AuthenticationFailed(ctx sdk.Context, authenticatorData authenticator.AuthenticatorData, msg sdk.Msg) {
}

func (s StatefulAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData authenticator.AuthenticatorData) bool {
	s.SetValue(ctx, s.GetValue(ctx)+1)
	return true
}

// This is a very naive implementation of an authenticator that tracks sends and blocks if the total amount sent is greater than 3_000
var _ authenticator.Authenticator = &MaxAmountAuthenticator{}
var _ authenticator.AuthenticatorData = &MaxAmountAuthenticatorData{}

type MaxAmountAuthenticatorData struct {
	Amount sdk.Int
}
type MaxAmountAuthenticator struct {
	KvStoreKey sdk.StoreKey
}

func (m MaxAmountAuthenticator) Gas() uint64 {
	return 0
}

func (m MaxAmountAuthenticator) Type() string {
	return "MaxAmountAuthenticator"
}

func (m MaxAmountAuthenticator) Initialize(data []byte) (authenticator.Authenticator, error) {
	return m, nil
}

func (m MaxAmountAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int8, simulate bool) (authenticator.AuthenticatorData, error) {
	return MaxAmountAuthenticatorData{}, nil
}

func (m MaxAmountAuthenticator) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) (bool, error) {
	send := msg.(*banktypes.MsgSend)
	if m.GetAmount(ctx).Add(send.Amount[0].Amount).GTE(sdk.NewInt(3_000)) {
		return false, nil
	}

	return true, nil
}

func (m MaxAmountAuthenticator) AuthenticationFailed(ctx sdk.Context, authenticatorData authenticator.AuthenticatorData, msg sdk.Msg) {
}

// TODO: Consider doing something like SetPubKey for determining if this authenticator was the one that authenticated the tx
func (m MaxAmountAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData authenticator.AuthenticatorData) bool {
	send := msg.(*banktypes.MsgSend)
	m.SetAmount(ctx, m.GetAmount(ctx).Add(send.Amount[0].Amount))
	return true
}

// The following methods for MaxAmountAuthenticator are similar to the set and get value methods for StatefulAuthenticator but set and get an int
func (m MaxAmountAuthenticator) SetAmount(ctx sdk.Context, amount sdk.Int) {
	kvStore := prefix.NewStore(ctx.KVStore(m.KvStoreKey), []byte(m.Type()))
	maxAmountData := MaxAmountAuthenticatorData{Amount: amount}
	newBz, _ := json.Marshal(maxAmountData)
	kvStore.Set([]byte("amount"), newBz)
}

func (m MaxAmountAuthenticator) GetAmount(ctx sdk.Context) sdk.Int {
	kvStore := prefix.NewStore(ctx.KVStore(m.KvStoreKey), []byte(m.Type()))
	bz := kvStore.Get([]byte("amount")) // global value. On the real thing we may want the account
	var amountData MaxAmountAuthenticatorData
	err := json.Unmarshal(bz, &amountData)
	// if we can't unmarshal, we just assume it's 0
	if err != nil {
		return sdk.NewInt(0)
	}
	return amountData.Amount
}
