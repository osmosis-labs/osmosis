package authenticator

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/utils"
)

type PeriodType string

const (
	Day   PeriodType = "day"
	Week  PeriodType = "week"
	Month PeriodType = "month"
	Year  PeriodType = "year"
)

type SpendLimitAuthenticator struct {
	store        sdk.KVStore
	quoteDenom   string
	bankKeeper   bankkeeper.Keeper
	allowedDelta osmomath.Uint
	periodType   PeriodType
}

var _ Authenticator = &SpendLimitAuthenticator{}

// NewSpendLimitAuthenticator creates a new SpendLimitAuthenticator. Creators must make sure to use a properly prefixed
// store with this authenticator. That is, prefix.NewStore(authenticatorsStoreKey, []byte("spendLimitAuthenticator"))
func NewSpendLimitAuthenticator(store sdk.KVStore, quoteDenom string, bankKeeper bankkeeper.Keeper) SpendLimitAuthenticator {
	// Ideally we'd validate that the store has been properly prefixed here, but the prefix store doesn't expose its prefix
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
		AllowedDelta uint64     `json:"allowed"`
		PeriodType   PeriodType `json:"period"`
	}

	if err := json.Unmarshal(data, &initData); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal initialization data")
	}

	sla.allowedDelta = sdk.NewUint(initData.AllowedDelta)
	if sla.allowedDelta.IsZero() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "allowed delta must be positive")
	}
	sla.periodType = initData.PeriodType
	if !(sla.periodType == Day || sla.periodType == Week || sla.periodType == Month || sla.periodType == Year) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period type %s", sla.periodType)
	}
	return sla, nil
}

func (sla SpendLimitAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int8,
	simulate bool,
) (AuthenticatorData, error) {
	return SignatureData{}, nil // No data needed for this authenticator
}

func (sla SpendLimitAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) AuthenticationResult {
	// Get the current period based on block time
	currentPeriod := formatPeriodTime(ctx.BlockTime(), sla.periodType)

	// Check if the period has changed
	activePeriod := sla.GetActivePeriod(account)
	if activePeriod != currentPeriod {
		// Delete past periods
		sla.DeletePastPeriods(account, ctx.BlockTime())
		// Update the current period
		sla.SetActivePeriod(account, currentPeriod)
	}

	// Store the balances
	balances := sla.bankKeeper.GetAllBalances(ctx, account)
	sla.SetBalance(account, balances)

	// We never authenticate ourselves. We just  authentication after the fact if the balances changed too much
	return NotAuthenticated()
}

func (sla SpendLimitAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) ConfirmationResult {
	prevBalances := sla.GetBalance(account)
	currentBalances := sla.bankKeeper.GetAllBalances(ctx, account)

	totalPrevValue := osmomath.NewInt(0)
	totalCurrentValue := osmomath.NewInt(0)

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

	if delta.Add(spentSoFar).Int64() > int64(sla.allowedDelta.Uint64()) {
		return Block(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "spend limit exceeded"))
	}

	// Update the total spent so far in the current period
	sla.SetSpentInPeriod(account, ctx.BlockTime(), delta.Add(spentSoFar))
	sla.DeleteBalances(account) // This is not 100% necessary, but it's nice to clean up after ourselves

	return Confirm()
}

func (sla SpendLimitAuthenticator) getPriceInQuoteDenom(_ sdk.Coin) osmomath.Int {
	// ToDo: Get current price (which pool do we base this on?)
	return osmomath.NewInt(1)
}

// STATE

func (sla SpendLimitAuthenticator) GetBalance(account sdk.AccAddress) []sdk.Coin {
	var coins []sdk.Coin
	_ = json.Unmarshal(sla.store.Get(getBalanceKey(account)), &coins)
	return coins
}

func (sla SpendLimitAuthenticator) SetBalance(account sdk.AccAddress, coins []sdk.Coin) {
	bz, _ := json.Marshal(coins)
	sla.store.Set(getBalanceKey(account), bz)
}

func (sla SpendLimitAuthenticator) DeleteBalances(account sdk.AccAddress) {
	osmoutils.DeleteAllKeysFromPrefix(sdk.Context{}, sla.store, getBalanceKey(account))
}

func (sla SpendLimitAuthenticator) GetSpentInPeriod(account sdk.AccAddress, t time.Time) osmomath.Int {
	key := getPeriodKey(account, sla.periodType, t)
	return sdk.NewIntFromBigInt(new(big.Int).SetBytes(sla.store.Get(key)))
}

func (sla SpendLimitAuthenticator) SetSpentInPeriod(account sdk.AccAddress, t time.Time, spent osmomath.Int) {
	key := getPeriodKey(account, sla.periodType, t)
	sla.store.Set(key, spent.BigInt().Bytes())
}

func (sla SpendLimitAuthenticator) DeletePastPeriods(account sdk.AccAddress, t time.Time) {
	currentPeriodKey := getPeriodKey(account, sla.periodType, t)

	// Use iterator range to focus on keys before the current period.
	prefixKey := utils.BuildKey(account, string(sla.periodType))
	iter := sla.store.Iterator(prefixKey, currentPeriodKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		sla.store.Delete(iter.Key())
	}
}

// GetActivePeriod gets the current period for the given account (i.e "currentPeriod|osmo1.. => day|2021-01-01")
func (sla SpendLimitAuthenticator) GetActivePeriod(account sdk.AccAddress) string {
	key := getActivePeriodKey(account, sla.periodType)
	return string(sla.store.Get(key))
}

// SetActivePeriod sets the current period for the given account (i.e "activePeriod|osmo1.. => day|2021-01-01")
func (sla SpendLimitAuthenticator) SetActivePeriod(account sdk.AccAddress, current string) {
	key := getActivePeriodKey(account, sla.periodType)
	sla.store.Set(key, []byte(current))
}

// KEYS
func getPeriodKey(account sdk.AccAddress, period PeriodType, t time.Time) []byte {
	if !(period == Day || period == Week || period == Month || period == Year) {
		panic("invalid period type")
	}
	return utils.BuildKey(account, period, formatPeriodTime(t, period))
}

func getBalanceKey(account sdk.AccAddress) []byte {
	return utils.BuildKey("balance", account.String())
}

func getActivePeriodKey(account sdk.AccAddress, period PeriodType) []byte {
	return utils.BuildKey("activePeriod", account.String(), string(period))
}

func formatPeriodTime(t time.Time, periodType PeriodType) string {
	switch periodType {
	case Day:
		return t.Format("2006-01-02")
	case Week:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-v%02d", year, week)
	case Month:
		return t.Format("2006-01")
	case Year:
		return t.Format("2006")
	}
	return ""
}
