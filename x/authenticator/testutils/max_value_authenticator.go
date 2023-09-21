package testutils

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
)

// This is a very naive implementation of an authenticator that tracks sends and blocks if the total amount sent is greater than 3_000
var _ authenticator.Authenticator = &MaxAmountAuthenticator{}
var _ authenticator.AuthenticatorData = &MaxAmountAuthenticatorData{}

type MaxAmountAuthenticatorData struct {
	Amount osmomath.Int
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

func (m MaxAmountAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) authenticator.AuthenticationResult {
	send, ok := msg.(*banktypes.MsgSend)
	if !ok {
		return authenticator.NotAuthenticated()
	}
	if m.GetAmount(ctx).Add(send.Amount[0].Amount).GTE(sdk.NewInt(3_000)) {
		return authenticator.NotAuthenticated()
	}

	return authenticator.Authenticated()
}

func (m MaxAmountAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) authenticator.ConfirmationResult {
	send, ok := msg.(*banktypes.MsgSend)
	if !ok {
		return authenticator.Confirm()
	}
	m.SetAmount(ctx, m.GetAmount(ctx).Add(send.Amount[0].Amount))
	return authenticator.Confirm()
}

// The following methods for MaxAmountAuthenticator are similar to the set and get value methods for StatefulAuthenticator but set and get an int
func (m MaxAmountAuthenticator) SetAmount(ctx sdk.Context, amount osmomath.Int) {
	kvStore := prefix.NewStore(ctx.KVStore(m.KvStoreKey), []byte(m.Type()))
	maxAmountData := MaxAmountAuthenticatorData{Amount: amount}
	newBz, _ := json.Marshal(maxAmountData)
	kvStore.Set([]byte("amount"), newBz)
}

func (m MaxAmountAuthenticator) GetAmount(ctx sdk.Context) osmomath.Int {
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
