package keeper

import (
	"github.com/google/cel-go/cel"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: sdkwrap errors

func (k Keeper) RegisterCell(
	ctx sdk.Context,
	cellID uint64,
	initState types.CellStateI,
	exprs []*types.Expr,
) error {
	if k.HasCell(ctx, cellID) {
		return errors.New("asdf")
	}

	cell := &types.Cell {
		Id: cellID,
		State: state,
	}

	k.SetCell(ctx, cellID, cell)

	for _, expr := range exprs {
		k.SetExpr(ctx, cellID, expr.Id, expr)
	}

	return nil
}

func (k Keeper) ExecuteExpression(
	ctx sdk.Context,
	cellID, exprID uint64,
	args map[string]interface{},
) (interface{}, error) {
	cell, err := k.GetCell(ctx, cellID)
	if err != nil {
		return nil, err
	}

	expr, err := k.GetExpr(ctx, cellID, exprID)
	if err != nil {
		return nil, err
	}

	// TODO: types

	decls := cell.State.Decls()
	decls = append(decls, expr.Inputs...)

	env, err := cel.NewEnv(decls)
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(expr.Expr)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	for k, v in range cell.State.Vars() {
		args[k] = v
	}

	out, _, err := prg.Eval(args)
	if err != nil {
		return nil, err
	}

	return out.Value(), nil
}
