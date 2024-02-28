package authenticator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"

	"github.com/osmosis-labs/osmosis/v23/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v23/x/twap"

	errorsmod "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/utils"
)

type PeriodType string

const (
	Day   PeriodType = "day"
	Week  PeriodType = "week"
	Month PeriodType = "month"
	Year  PeriodType = "year"
)

type PriceStrategy string

const (
	Twap          PriceStrategy = "twap"
	AbsoluteValue PriceStrategy = "absolute_value"
	// Spot          PriceStrategy = "spot"
)

type SpendLimitAuthenticator struct {
	store             sdk.KVStore
	storeKey          storetypes.StoreKey
	quoteDenom        string
	bankKeeper        bankkeeper.Keeper
	poolManagerKeeper *poolmanager.Keeper
	twapKeeper        *twap.Keeper
	priceStrategy     PriceStrategy

	allowedDelta osmomath.Uint
	periodType   PeriodType
}

var _ iface.Authenticator = &SpendLimitAuthenticator{}

// NewSpendLimitAuthenticator creates a new SpendLimitAuthenticator. Creators must make sure to use a properly prefixed
// store with this authenticator. That is, prefix.NewStore(authenticatorsStoreKey, []byte("spendLimitAuthenticator"))
func NewSpendLimitAuthenticator(storeKey storetypes.StoreKey, quoteDenom string, priceStrategy PriceStrategy, bankKeeper bankkeeper.Keeper, poolManagerKeeper *poolmanager.Keeper, twapKeeper *twap.Keeper) SpendLimitAuthenticator {
	// Ideally we'd validate that the store has been properly prefixed here, but the prefix store doesn't expose its prefix
	if !(priceStrategy == AbsoluteValue || priceStrategy == Twap) {
		panic("invalid price strategy")
	}
	if priceStrategy == Twap && twapKeeper == nil {
		panic("twap keeper must be provided when using twap price strategy")
	}
	return SpendLimitAuthenticator{
		storeKey:          storeKey,
		quoteDenom:        quoteDenom,
		bankKeeper:        bankKeeper,
		poolManagerKeeper: poolManagerKeeper,
		twapKeeper:        twapKeeper,
		priceStrategy:     priceStrategy,
	}
}

type InitData struct {
	AllowedDelta uint64     `json:"allowed"`
	PeriodType   PeriodType `json:"period"`
}

func (sla SpendLimitAuthenticator) Type() string {
	return "SpendLimitAuthenticator"
}

func (sla SpendLimitAuthenticator) StaticGas() uint64 {
	return 5000
}

func (sla SpendLimitAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	var initData InitData
	if err := json.Unmarshal(data, &initData); err != nil {
		return nil, errorsmod.Wrap(err, "failed to unmarshal initialization data")
	}
	sla.allowedDelta = sdk.NewUint(initData.AllowedDelta)
	if sla.allowedDelta.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "allowed delta must be positive")
	}
	sla.periodType = initData.PeriodType
	if !(sla.periodType == Day || sla.periodType == Week || sla.periodType == Month || sla.periodType == Year) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period type %s", sla.periodType)
	}
	return sla, nil
}

func (sla SpendLimitAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) error {
	// We never authenticate ourselves. We just confirm execution after the fact if the balances changed too much
	return nil
}

func (sla SpendLimitAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, feePayer sdk.AccAddress, msg sdk.Msg, msgIndex uint64,
	authenticatorId string) error {
	sla.store = prefix.NewStore(ctx.KVStore(sla.storeKey), []byte(sla.Type()))
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
	return nil
}

func (sla SpendLimitAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) error {
	sla.store = prefix.NewStore(ctx.KVStore(sla.storeKey), []byte(sla.Type()))
	prevBalances := sla.GetBalance(request.Account)
	currentBalances := sla.bankKeeper.GetAllBalances(ctx, request.Account)

	totalPrevValue := osmomath.NewInt(0)
	totalCurrentValue := osmomath.NewInt(0)

	for _, coin := range prevBalances {
		price, err := sla.getPriceInQuoteDenom(ctx, coin)
		if err != nil {
			// ToDO: what do we want to do if we can't determine the price of an asset?
			return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "can't find price for %s", coin.Denom)
		}
		totalPrevValue = totalPrevValue.Add(price.MulInt(coin.Amount).RoundInt())
	}

	for _, coin := range currentBalances {
		price, err := sla.getPriceInQuoteDenom(ctx, coin)
		if err != nil {
			// ToDO: what do we want to do if we can't determine the price of an asset?
			return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "can't find price for %s", coin.Denom)
		}
		totalCurrentValue = totalCurrentValue.Add(price.MulInt(coin.Amount).RoundInt())
	}

	delta := totalPrevValue.Sub(totalCurrentValue)

	// Get the total spent so far in the current period
	spentSoFar := sla.GetSpentInPeriod(request.Account, ctx.BlockTime())

	if delta.Add(spentSoFar).Int64() > int64(sla.allowedDelta.Uint64()) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "spend limit exceeded")
	}

	// Update the total spent so far in the current period
	sla.SetSpentInPeriod(request.Account, ctx.BlockTime(), delta.Add(spentSoFar))
	sla.DeleteBalances(request.Account) // This is not 100% necessary, but it's nice to clean up after ourselves

	return nil
}

func (sla SpendLimitAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	var initData InitData
	if err := json.Unmarshal(data, &initData); err != nil {
		return errorsmod.Wrap(err, "failed to unmarshal initialization data")
	}
	if sdk.NewUint(initData.AllowedDelta).IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "allowed delta must be positive")
	}
	if !(initData.PeriodType == Day || initData.PeriodType == Week || initData.PeriodType == Month || initData.PeriodType == Year) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period type %s", initData.PeriodType)
	}
	return nil
}

func (sla SpendLimitAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	return nil
}

func (sla SpendLimitAuthenticator) getPriceInQuoteDenom(ctx sdk.Context, coin sdk.Coin) (osmomath.Dec, error) {
	switch sla.priceStrategy {
	case Twap:
		// This is a very bad and inefficient implementation that should be improved
		oneWeekAgo := ctx.BlockTime().Add(-time.Hour * 24 * 7)
		numPools := sla.poolManagerKeeper.GetNextPoolId(ctx)
		for i := uint64(1); i < numPools; i++ {
			price, err := sla.twapKeeper.GetArithmeticTwapToNow(ctx, i, coin.Denom, sla.quoteDenom, oneWeekAgo)
			if err == nil {
				return price, nil
			}
		}
		return osmomath.Dec{}, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "no price found for %s", coin.Denom)
	case AbsoluteValue:
		return osmomath.NewDec(1), nil
	default:
		return osmomath.Dec{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid price strategy %s", sla.priceStrategy)
	}
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
	osmoutils.DeleteAllKeysFromPrefix(sla.store, getBalanceKey(account))
}

func (sla SpendLimitAuthenticator) GetSpentInPeriod(account sdk.AccAddress, t time.Time) osmomath.Int {
	key := getPeriodKey(account, sla.periodType, t)
	var spent osmomath.Int
	err := json.Unmarshal(sla.store.Get(key), &spent)
	if err != nil {
		return osmomath.ZeroInt()
	}
	return spent
}

func (sla SpendLimitAuthenticator) SetSpentInPeriod(account sdk.AccAddress, t time.Time, spent osmomath.Int) {
	key := getPeriodKey(account, sla.periodType, t)
	bz, err := json.Marshal(spent)
	if err != nil {
		panic("couldn't marshal spent") // TODO: deal with this
	}
	sla.store.Set(key, bz)
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
	return utils.BuildKey("balance", account)
}

func getActivePeriodKey(account sdk.AccAddress, period PeriodType) []byte {
	return utils.BuildKey("activePeriod", account, string(period))
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
