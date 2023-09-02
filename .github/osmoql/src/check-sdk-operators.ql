/**
 * @name Wrong operator between sdk types.
 * @kind problem
 * @problem.severity warning
 * @id osmo-ql/wrong-binary-operators
 */

/* 
problem.severity is set to warning, because in some cases using native operators between sdk.Int(s) or sdk.Dec(s) is justified
ex: https://github.com/osmosis-labs/osmosis/blob/main/x/twap/logic.go#L58-L63
*/

import go

from BinaryExpr b
where
  (b.getOperator() = "==")
  and (
    // find sdk.Dec(s)
    (b.getLeftOperand().getType().hasQualifiedName("github.com/cosmos/cosmos-sdk/types", "Dec") or b.getRightOperand().getType().hasQualifiedName("github.com/cosmos/cosmos-sdk/types", "Dec")) or
    // find sdk.Int(s)
    (b.getLeftOperand().getType().hasQualifiedName("github.com/cosmos/cosmos-sdk/types", "Int") or b.getRightOperand().getType().hasQualifiedName("github.com/cosmos/cosmos-sdk/types", "Int")) or 
    // find osmomath.BigDec(s)
    (b.getLeftOperand().getType().hasQualifiedName("github.com/osmosis-labs/osmosis/osmomath", "BigDec") or b.getRightOperand().getType().hasQualifiedName("github.com/osmosis-labs/osmosis/osmomath", "BigDec"))
  )
select b, "Use SDK operators instead of native operators when dealing with sdk.Int or sdk.Dec."