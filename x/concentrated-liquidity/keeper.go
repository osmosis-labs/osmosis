package concentrated_liquidity

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	paramSpace paramtypes.Subspace

	// keepers
	swaprouterKeeper types.SwaprouterKeeper
	bankKeeper       types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, bankKeeper types.BankKeeper, paramSpace paramtypes.Subspace) *Keeper {
	// ParamSubspace must be initialized within app/keepers/keepers.go
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		cdc:        cdc,
		bankKeeper: bankKeeper,
	}
}

// GetParams returns the total set of concentrated-liquidity module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the concentrated-liquidity module's parameters with the provided parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// TODO: implement minting, spec, tests
func (k Keeper) InitializePool(ctx sdk.Context, poolI swaproutertypes.PoolI, creatorAddress sdk.AccAddress) error {
	pool, ok := poolI.(types.ConcentratedPoolExtension)
	if !ok {
		return errors.New("invalid pool type when setting concentrated pool")
	}
	//poolId := pool.GetId()

	// Add the share token's meta data to the bank keeper.
	// poolShareBaseDenom := types.GetPoolShareDenom(poolId)
	// poolShareDisplayDenom := fmt.Sprintf("GAMM-%d", poolId)
	// k.bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
	// 	Description: fmt.Sprintf("The share token of the gamm pool %d", poolId),
	// 	DenomUnits: []*banktypes.DenomUnit{
	// 		{
	// 			Denom:    poolShareBaseDenom,
	// 			Exponent: 0,
	// 			Aliases: []string{
	// 				"attopoolshare",
	// 			},
	// 		},
	// 		{
	// 			Denom:    poolShareDisplayDenom,
	// 			Exponent: types.OneShareExponent,
	// 			Aliases:  nil,
	// 		},
	// 	},
	// 	Base:    poolShareBaseDenom,
	// 	Display: poolShareDisplayDenom,
	// })

	// // Mint the initial pool shares share token to the sender
	// err := k.MintPoolShareToAccount(ctx, pool, creatorAddress, pool.GetTotalShares())
	// if err != nil {
	// 	return err
	// }

	// k.RecordTotalLiquidityIncrease(ctx, pool.GetTotalPoolLiquidity(ctx))

	return k.setPool(ctx, pool)
}

// Set the swaprouter keeper.
func (k *Keeper) SetSwapRouterKeeper(swaprouterKeeper types.SwaprouterKeeper) {
	k.swaprouterKeeper = swaprouterKeeper
}
