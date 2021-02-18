# CEL module

integrates cel-go

## Introduction

A cel-go expression type is defined by (EnvType, Input parameter type, Output parameter type).
Expressions those shares the same functionality(e.g. a curve and its parameter updater) are under the same EnvType.
We call a set of expressions a cell. Expressions under a cell should share same EnvType(which acts like a state). 
Expressions can have different input/output parameter types. Output type includes effects.
For the sake of simplicity, we read the entire variable set defined by EnvType. This could be optimized later.
For this reason we allow finite set of variables used inside the expression.
Later we can parameterize EnvSet to use specific variable only.

## Components

### EnvType

An EnvType is, roughly, a JSON object type, such as:

```ts
interface User {
  balance: number;
  moniker: string;
  minter:  address;
}
```

An EnvType defines the state structure that is going to be used by a set of expressions.
When executing an expression, the module will load corresponding EnvType from the state,
and initializes a cel-go Env providing the EnvSet declarations to cel.NewEnv.

Custom functions(such as precompiled curve calculation) is EnvType-dependent. 
For Osmosis, my opinion is that EnvType extension should be done on chain level governance.

### Cell

A cell is somewhat like a contract. Cell can consists of multiple expressions, which shares the same state.
New expressions can be dynamically added / existing expressions can be updated through cell-level governance.

### Expression

Three types of expressions:
1. Input -> Output (pure function)
2. (State, Input) -> Output (reader function)
3. (State, Input) -> Effect (mutable function)

(State modification is not what cel-go is defined for, but I think we can hack it to return "effects" that 
has some mutable effects. We need some sort of mutability for cell level governance anyway. Note this is 
how functional languages handles side effects, still maintaining purity. This is a hack so would be better to separate 2/3)

TODO: I belive that, we can add a macro that "calls" another pure/reader expression inside an expression. 
As long as it does not call mutable function(effects cannot be processed inside an expression), and the callee
is within the same cell, it is a good sideway to use output of other expressions. 
Within a single expression invocation from an external actor, there still will be no state modification.

## Execution process

1. An external actor provides(Cell ID, Expression ID, Expression Input)
2. The module retrieves the cell, the EnvType, and the expression.
3. Retrieve and deserialize the state using the EnvType.
4. Setup cel Env using EnvType and the state.
5. Provide the state and the user input to the expression, evaluate it.
6. Return the output or process the effects.

## KVStore State

```go
type Cell struct {
  ID string
  EnvType EnvType
}

// EnvType should be an interface.
// Macros are EnvType specific. New EnvType is added through chain upgrade. 
type EnvType struct {
  Types []*exprpb.Decl
  State map[string]string
}

func KeyCell(id string) []byte {
  return join([]byte(0x00), id)
}

type Expr struct {
  InputTypes []*exprpb.Decl
  OutputType *exprpb.Type
}

func KeyExpr(cellID string, exprID string) []byte {
  return join([]byte{0x00}, cellID, exprID)
}
```


## Memo

- CellAccount
