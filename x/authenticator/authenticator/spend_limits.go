package authenticator

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/utils"
	"math/big"
	"time"
)

type Period string

const (
	Day  Period = "day"
	Week Period = "week"
	Year Period = "year"
)

type SpendLimitAuthenticator struct {
	store        sdk.KVStore
	quoteDenom   string
	bankKeeper   bankkeeper.Keeper
	allowedDelta sdk.Int
	period       Period
}

var _ Authenticator = &SpendLimitAuthenticator{}

func NewSpendLimitAuthenticator(store sdk.KVStore, quoteDenom string, bankKeeper bankkeeper.Keeper) SpendLimitAuthenticator {
	return SpendLimitAuthenticator{
		store:      store,
		quoteDenom: quoteDenom,
		bankKeeper: bankKeeper,
	}
}

func (sla SpendLimitAuthenticator) Type() string {
	return "SpendLimitAuthenticator"
}

func (sla SpendLimitAuthenticator) StaticGas() uint64 {
	return 5000
}

func (sla SpendLimitAuthenticator) Initialize(data []byte) (Authenticator, error) {
	var initData struct {
		AllowedDelta int64  `json:"allowed_delta"`
		Period       Period `json:"period"`
	}

	if err := json.Unmarshal(data, &initData); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal initialization data")
	}

	sla.allowedDelta = sdk.NewInt(initData.AllowedDelta)
	sla.period = initData.Period
	return sla, nil
}

func (sla SpendLimitAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int8,
	simulate bool,
) (AuthenticatorData, error) {
	return nil, nil
}

func (sla SpendLimitAuthenticator) Authenticate(
	ctx sdk.Context,
	msg sdk.Msg,
	_ AuthenticatorData,
) AuthenticationResult {
	account := msg.GetSigners()[0]
	sla.DeleteBlockBalances(ctx, account)
	sla.DeletePastPeriods(account, ctx.BlockTime()) // TODO: implement this

	// Store the balances
	balances := sla.bankKeeper.GetAllBalances(ctx, account)
	sla.SetBlockBalance(account, ctx.BlockHeight(), balances)

	// We never authenticate ourselves. We just block authentication after the fact if the balances changed too much
	return NotAuthenticated()
}

func (sla SpendLimitAuthenticator) AuthenticationFailed(ctx sdk.Context, _ AuthenticatorData, msg sdk.Msg) {
	account := msg.GetSigners()[0]
	sla.DeleteBlockBalances(ctx, account)
}

func (sla SpendLimitAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, _ AuthenticatorData) ConfirmationResult {
	account := msg.GetSigners()[0]

	prevBalances := sla.GetBlockBalance(account, ctx.BlockHeight())
	currentBalances := sla.bankKeeper.GetAllBalances(ctx, account)

	totalPrevValue := sdk.NewInt(0)
	totalCurrentValue := sdk.NewInt(0)

	for _, coin := range prevBalances {
		price := sla.getPriceInQuoteDenom(coin)
		totalPrevValue = totalPrevValue.Add(price.Mul(coin.Amount))
	}

	for _, coin := range currentBalances {
		price := sla.getPriceInQuoteDenom(coin)
		totalCurrentValue = totalCurrentValue.Add(price.Mul(coin.Amount))
	}

	delta := totalPrevValue.Sub(totalCurrentValue)

	// Get the total spent so far in the current period
	spentSoFar := sla.GetSpentInPeriod(account, ctx.BlockTime())

	if delta.Add(spentSoFar).GT(sla.allowedDelta) {
		return Block(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "spend limit exceeded"))
	}

	// Update the total spent so far in the current period
	sla.SetSpentInPeriod(account, ctx.BlockTime(), delta.Add(spentSoFar))
	osmoutils.DeleteAllKeysFromPrefix(ctx, sla.store, getAccountKey(account))

	return Confirm()
}

func (sla SpendLimitAuthenticator) getPriceInQuoteDenom(coin sdk.Coin) sdk.Int {
	// ToDo: Get current price (which pool do we base this on?)
	return sdk.NewInt(1)
}

// STATE

func (sla SpendLimitAuthenticator) GetBlockBalance(account sdk.AccAddress, blockHeight int64) []sdk.Coin {
	var coins []sdk.Coin
	_ = json.Unmarshal(sla.store.Get(getAccountBlockKey(account, blockHeight)), &coins)
	return coins
}

func (sla SpendLimitAuthenticator) SetBlockBalance(account sdk.AccAddress, blockHeight int64, coins []sdk.Coin) {
	bz, _ := json.Marshal(coins)
	sla.store.Set(getAccountBlockKey(account, blockHeight), bz)
}

func (sla SpendLimitAuthenticator) DeleteBlockBalances(ctx sdk.Context, account sdk.AccAddress) {
	osmoutils.DeleteAllKeysFromPrefix(ctx, sla.store, getAccountKey(account))
}

func (sla SpendLimitAuthenticator) GetSpentInPeriod(account sdk.AccAddress, t time.Time) sdk.Int {
	return sdk.NewIntFromBigInt(new(big.Int).SetBytes(sla.store.Get(getPeriodKey(account, sla.period, t))))
}

func (sla SpendLimitAuthenticator) SetSpentInPeriod(account sdk.AccAddress, t time.Time, spent sdk.Int) {
	sla.store.Set(getPeriodKey(account, sla.period, t), spent.BigInt().Bytes())
}

func (sla SpendLimitAuthenticator) DeletePastPeriods(account sdk.AccAddress, t time.Time) {
	// TODO
}

// KEYS
func getPeriodKey(account sdk.AccAddress, period Period, t time.Time) []byte {
	switch period {
	case Day:
		return utils.BuildKey(account, "day", t.Format("2006-01-02"))
	case Week:
		return utils.BuildKey(account, "week", t.Format("2006-01"))
	case Year:
		return utils.BuildKey(account, "year", t.Format("2006"))
	}
	return nil
}

func getAccountKey(account sdk.AccAddress) []byte {
	return utils.BuildKey("block_balance", account.String())
}

func getAccountBlockKey(account sdk.AccAddress, blockHeight int64) []byte {
	return utils.BuildKey("block_balance", account.String(), sdk.Uint64ToBigEndian(uint64(blockHeight)))
}
