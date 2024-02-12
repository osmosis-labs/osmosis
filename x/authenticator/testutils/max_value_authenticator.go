package testutils

import (
	"encoding/json"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// This is a very naive implementation of an authenticator that tracks sends and blocks if the total amount sent is greater than 3_000
var _ iface.Authenticator = &MaxAmountAuthenticator{}

type MaxAmountAuthenticatorData struct {
	Amount osmomath.Int
}
type MaxAmountAuthenticator struct {
	KvStoreKey storetypes.StoreKey
}

func (m MaxAmountAuthenticator) StaticGas() uint64 {
	return 0
}

func (m MaxAmountAuthenticator) Type() string {
	return "MaxAmountAuthenticator"
}

func (m MaxAmountAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	return m, nil
}

func (m MaxAmountAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	if request.Msg.TypeURL != "/cosmos.bank.v1beta1.MsgSend" {
		return iface.NotAuthenticated()
	}
	// unmarshal the message.value into the bank.MsgSend struct
	var send banktypes.MsgSend
	err := proto.Unmarshal(request.Msg.Bytes, &send)
	if err != nil {
		return iface.NotAuthenticated()
	}
	if m.GetAmount(ctx).Add(send.Amount[0].Amount).GTE(sdk.NewInt(3_000)) {
		return iface.NotAuthenticated()
	}

	return iface.Authenticated()
}

func (m MaxAmountAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (m MaxAmountAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
	if request.Msg.TypeURL != "/cosmos.bank.v1beta1.MsgSend" {
		return iface.Confirm()
	}
	// unmarshal the message.value into the bank.MsgSend struct
	var send banktypes.MsgSend
	err := proto.Unmarshal(request.Msg.Bytes, &send)
	if err != nil {
		return iface.Confirm()
	}
	m.SetAmount(ctx, m.GetAmount(ctx).Add(send.Amount[0].Amount))
	return iface.Confirm()
}

func (m MaxAmountAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}

func (m MaxAmountAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
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