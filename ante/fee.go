package ante

import (
	"bytes"
	errorsmod "cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	treasurytypes "github.com/osmosis-labs/osmosis/v27/x/treasury/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

// DeductFeeDecorator deducts fees from the first signer of the tx.
// If the first signer does not have the funds to pay for the fees, we return an InsufficientFunds error.
// We call next AnteHandler if fees successfully deducted.
//
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	ak             ante.AccountKeeper
	bankKeeper     BankKeeper
	feegrantKeeper ante.FeegrantKeeper
	txFeesKeeper   txfeestypes.TxFeesKeeper
	oracleKeeper   OracleKeeper
	treasuryKeeper TreasuryKeeper
}

func NewDeductFeeDecorator(txk txfeestypes.TxFeesKeeper, ak ante.AccountKeeper, bk BankKeeper, fk ante.FeegrantKeeper, tk TreasuryKeeper, ok OracleKeeper) DeductFeeDecorator {
	return DeductFeeDecorator{
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		txFeesKeeper:   txk,
		treasuryKeeper: tk,
		oracleKeeper:   ok,
	}
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// checks to make sure the auth module account has been set to collect tx fees in base token, to be used for staking rewards
	if addr := dfd.ak.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("fee collector module account (%s) has not been set", authtypes.FeeCollectorName)
	}

	// checks to make sure a separate module account has been set to collect tx fees not in base token
	if addrNonNativeFee := dfd.ak.GetModuleAddress(txfeestypes.NonNativeTxFeeCollectorName); addrNonNativeFee == nil {
		return ctx, fmt.Errorf("fee collector for staking module account (%s) has not been set", txfeestypes.NonNativeTxFeeCollectorName)
	}

	baseDenom, err := dfd.txFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return ctx, fmt.Errorf("could not retrieve base denom: %w", err)
	}

	msgs := feeTx.GetMsgs()
	taxes := FilterMsgAndComputeTax(ctx, dfd.treasuryKeeper, msgs...)
	taxesInBaseDenom := sdk.NewCoin(baseDenom, taxes.AmountOf(baseDenom))
	for _, denom := range taxes.Denoms() {
		if denom == baseDenom {
			continue
		}
		exchangeRate, err := dfd.oracleKeeper.GetMelodyExchangeRate(ctx, denom)
		if err != nil {
			return ctx, fmt.Errorf("could not retrieve exchange rate for %s: %w", denom, err)
		}
		taxesInBaseDenom = taxesInBaseDenom.AddAmount(taxes.AmountOf(denom).ToLegacyDec().Mul(exchangeRate).TruncateInt())
	}

	// fee can be in any denom (checked for validity later)
	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	// set the fee payer as the default address to deduct fees from
	deductFeesFrom := feePayer

	// If a fee granter was set, deduct fee from the fee granter's account.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "fee grants is not enabled")
		} else if !bytes.Equal(feeGranter, feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())
			if err != nil {
				return ctx, errorsmod.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		// if no errors, change the account that is charged for fees to the fee granter
		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	fees := feeTx.GetFee()

	// if we are simulating, set the fees to 1 note as they don't matter.
	// set it as coming from the burn addr
	if simulate && fees.IsZero() {
		fees = sdk.NewCoins(sdk.NewInt64Coin("note", 1))
		burnAcctAddr, _ := sdk.AccAddressFromBech32("symphony1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqymqs4m")
		// were doing 1 extra get account call alas
		burnAcct := dfd.ak.GetAccount(ctx, burnAcctAddr)
		if burnAcct != nil {
			deductFeesFromAcc = burnAcct
		}
	}

	// deducts the fees and transfer them to the module account
	if !fees.IsZero() {
		err = DeductFees(dfd.txFeesKeeper, dfd.bankKeeper, ctx, deductFeesFromAcc, fees, taxesInBaseDenom)
		if err != nil {
			return ctx, err
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
		sdk.NewAttribute("taxes", taxes.String()),
	)})

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account and transfers them to the set module account.
func DeductFees(txFeesKeeper txfeestypes.TxFeesKeeper, bankKeeper BankKeeper, ctx sdk.Context, acc authtypes.AccountI, fees sdk.Coins, baseDenomTax sdk.Coin) error {
	// Checks the validity of the fee tokens (sorted, have positive amount, valid and unique denomination)
	if !fees.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	// pulls base denom from TxFeesKeeper (should be NOTE)
	baseDenom, err := txFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	// checks if input fee is NOTE (assumes only one fee token exists in the fees array (as per the check in mempoolFeeDecorator))
	if fees[0].Denom == baseDenom {
		deductedFees, anyNegative := fees.SafeSub(baseDenomTax)
		if anyNegative {
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees (%s) to apply tax (%s)", fees[0], baseDenomTax)
		}
		if baseDenomTax.IsPositive() {
			err = bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), treasurytypes.ModuleName, sdk.Coins{baseDenomTax})
			if err != nil {
				return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
			}
		}

		// sends to FeeCollectorName module account, which distributes staking rewards
		err = bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), authtypes.FeeCollectorName, deductedFees)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	} else {
		// sends to FeeCollectorForStakingRewardsName module account
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), txfeestypes.NonNativeTxFeeCollectorName, fees)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	return nil
}
