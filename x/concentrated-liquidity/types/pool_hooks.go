package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// Action prefixes for pool actions
const (
	CreatePositionPrefix     = "CreatePosition"
	WithdrawPositionPrefix   = "WithdrawPosition"
	SwapExactAmountInPrefix  = "SwapExactAmountIn"
	SwapExactAmountOutPrefix = "SwapExactAmountOut"
)

// Helper function to generate before action prefix
func BeforeActionPrefix(action string) string {
	return fmt.Sprintf("before%s", action)
}

// Helper function to generate after action prefix
func AfterActionPrefix(action string) string {
	return fmt.Sprintf("after%s", action)
}

// GetAllActionPrefixes returns all the action prefixes corresponding to valid hooks
func GetAllActionPrefixes() []string {
	result := []string{}
	for _, prefix := range []string{
		CreatePositionPrefix,
		WithdrawPositionPrefix,
		SwapExactAmountInPrefix,
		SwapExactAmountOutPrefix,
	} {
		result = append(result, BeforeActionPrefix(prefix), AfterActionPrefix(prefix))
	}

	return result
}

// --- Sudo Message Wrappers ---

type BeforeCreatePositionSudoMsg struct {
	BeforeCreatePosition BeforeCreatePositionMsg `json:"before_create_position"`
}

type AfterCreatePositionSudoMsg struct {
	AfterCreatePosition AfterCreatePositionMsg `json:"after_create_position"`
}

type BeforeWithdrawPositionSudoMsg struct {
	BeforeWithdrawPosition BeforeWithdrawPositionMsg `json:"before_withdraw_position"`
}

type AfterWithdrawPositionSudoMsg struct {
	AfterWithdrawPosition AfterWithdrawPositionMsg `json:"after_withdraw_position"`
}

type BeforeSwapExactAmountInSudoMsg struct {
	BeforeSwapExactAmountIn BeforeSwapExactAmountInMsg `json:"before_swap_exact_amount_in"`
}

type AfterSwapExactAmountInSudoMsg struct {
	AfterSwapExactAmountIn AfterSwapExactAmountInMsg `json:"after_swap_exact_amount_in"`
}

type BeforeSwapExactAmountOutSudoMsg struct {
	BeforeSwapExactAmountOut BeforeSwapExactAmountOutMsg `json:"before_swap_exact_amount_out"`
}

type AfterSwapExactAmountOutSudoMsg struct {
	AfterSwapExactAmountOut AfterSwapExactAmountOutMsg `json:"after_swap_exact_amount_out"`
}

// --- Message structs ---

type BeforeCreatePositionMsg struct {
	PoolId         uint64             `json:"pool_id"`
	Owner          sdk.AccAddress     `json:"owner"`
	TokensProvided []wasmvmtypes.Coin `json:"tokens_provided"`
	Amount0Min     osmomath.Int       `json:"amount_0_min"`
	Amount1Min     osmomath.Int       `json:"amount_1_min"`
	LowerTick      int64              `json:"lower_tick"`
	UpperTick      int64              `json:"upper_tick"`
}

type AfterCreatePositionMsg struct {
	PoolId         uint64             `json:"pool_id"`
	Owner          sdk.AccAddress     `json:"owner"`
	TokensProvided []wasmvmtypes.Coin `json:"tokens_provided"`
	Amount0Min     osmomath.Int       `json:"amount_0_min"`
	Amount1Min     osmomath.Int       `json:"amount_1_min"`
	LowerTick      int64              `json:"lower_tick"`
	UpperTick      int64              `json:"upper_tick"`
}

type BeforeWithdrawPositionMsg struct {
	PoolId           uint64         `json:"pool_id"`
	Owner            sdk.AccAddress `json:"owner"`
	PositionId       uint64         `json:"position_id"`
	AmountToWithdraw osmomath.Dec   `json:"amount_to_withdraw"`
}

type AfterWithdrawPositionMsg struct {
	PoolId           uint64         `json:"pool_id"`
	Owner            sdk.AccAddress `json:"owner"`
	PositionId       uint64         `json:"position_id"`
	AmountToWithdraw osmomath.Dec   `json:"amount_to_withdraw"`
}

type BeforeSwapExactAmountInMsg struct {
	PoolId            uint64           `json:"pool_id"`
	Sender            sdk.AccAddress   `json:"sender"`
	TokenIn           wasmvmtypes.Coin `json:"token_in"`
	TokenOutDenom     string           `json:"token_out_denom"`
	TokenOutMinAmount osmomath.Int     `json:"token_out_min_amount"`
	SpreadFactor      osmomath.Dec     `json:"spread_factor"`
}

type AfterSwapExactAmountInMsg struct {
	PoolId            uint64           `json:"pool_id"`
	Sender            sdk.AccAddress   `json:"sender"`
	TokenIn           wasmvmtypes.Coin `json:"token_in"`
	TokenOutDenom     string           `json:"token_out_denom"`
	TokenOutMinAmount osmomath.Int     `json:"token_out_min_amount"`
	SpreadFactor      osmomath.Dec     `json:"spread_factor"`
}

type BeforeSwapExactAmountOutMsg struct {
	PoolId           uint64           `json:"pool_id"`
	Sender           sdk.AccAddress   `json:"sender"`
	TokenInDenom     string           `json:"token_in_denom"`
	TokenInMaxAmount osmomath.Int     `json:"token_in_max_amount"`
	TokenOut         wasmvmtypes.Coin `json:"token_out"`
	SpreadFactor     osmomath.Dec     `json:"spread_factor"`
}

type AfterSwapExactAmountOutMsg struct {
	PoolId           uint64           `json:"pool_id"`
	Sender           sdk.AccAddress   `json:"sender"`
	TokenInDenom     string           `json:"token_in_denom"`
	TokenInMaxAmount osmomath.Int     `json:"token_in_max_amount"`
	TokenOut         wasmvmtypes.Coin `json:"token_out"`
	SpreadFactor     osmomath.Dec     `json:"spread_factor"`
}
