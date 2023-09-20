package testutils

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
)

// This is a very naive implementation of an authenticator that tracks sends and blocks if the total amount sent is greater than 3_000
var _ authenticator.Authenticator = &MaxAmountAuthenticator{}
var _ authenticator.AuthenticatorData = &MaxAmountAuthenticatorData{}

type MaxAmountAuthenticatorData struct {
	Amount sdk.Int
}
type MaxAmountAuthenticator struct {
	KvStoreKey sdk.StoreKey
}

func (m MaxAmountAuthenticator) StaticGas() uint64 {
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
func (m MaxAmountAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) bool {
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
