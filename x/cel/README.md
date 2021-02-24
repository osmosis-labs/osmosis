# CEL module

integrates cel-go

## Introduction

CEL is a non turing complete embeddable scripting language that can be used to describe custom expressions
to be used by the more complex logic. Some of the specifics of CEL are:

- Variables are not explicitly passed, but the underlying environment resolves the names. This environment
  has to be provided by the expression caller.
- Cannot access mutable storage. CEL module will support storage without sync and access interface.
- Has native support of
-   constructing protobuf and json messages
-   adding custom functions and types
-   disabling specific functions and types

We will wrap expressions into an object unit `Cell`. A cell consists of multiple expressions and a shared state.
The number of expressions are not bounded, but the state has to be finite. 
This is because when the entire parameter set has to be provided as its execution environment at evaluation.
An expression is defined as a pair of CEL expression and its input/output type.

Expressions can be either pure or effectful. A pure expression is intended to be readonly upon the input parameters
and the state, and returns a value. An effectful expression returns a list of effects which is possibly side effectful
commands expected to be processed by the cel module.

Custom functions and types cannot be generated runtime(as they has to be written in golang and compiled). A set of 
precompiled functions and types is called plugin. An expression can have a set of plugins to be attatched when it is
registered to the cel module. The module will provide the plugin information to the execution environment at the 
evaluation.

## Components

### Expression

```go
// Runtime
type Expression struct {
  ID string
  CellID string
  Expr *cel.Expr
  Inputs []*cel.Decl
  Output *cel.Type
  Plugins []string
}
```

### Cell

```go
type Cell struct {
  ID string
  State CellState
}

// Runtime
type CellState interface {
  ID() string
  StateFrame() StateFrame
  StateVars() map[string]interface{}
  Plugins() []Plugin
}

// Runtime
type StateFrame interface {
  Types() []interface{}
  Decls() []*exprpb.Decl
}

// Compile time
type Plugin interface {
  ID() string
  StateFrame() StateFrame
  Funcs() []*functions.Overload
}
```

### Effect

Effects are encoded as either protobuf message or json.

```go
type EffectType enum

const (
  Abort EffectType = iota
  SetState
  Transfer
  ModuleEffect
  // ...
)
```

### Custom functions

```go
var _ Plugin = PluginAmm{}

type PluginAMM struct {}

func (_ PluginAMM) ID() string {
  return "AMM"
}

func (_ PluginAMM) StateFrame() StateFrame {
  return PluginStateFrame {
    Funcs: []*functions.Overload{
      &functions.Overload{
        Operator: "amm_curve_calculation",
        Function: func(args ...ref.Val) ref.Val {
          if len(args) != 3 {
            return types.NewErr("invalid number of argument given to amm_curve_calculation")
          }
          poolA, ok := args[0].(types.Uint)
          if !ok {
            return types.ValOrErr(args[0], "type error on argument 0: expected type uint")
          }
          poolB, ok := args[1].(types.Uint)
          if !ok {
            return types.ValOrErr(args[1], "type error on argument 1: expected type uint")
          }
          value, ok := args[2].(types.Uint)
          if !ok {
            return types.ValOrErr(args[2], "type error on argument 2: expected type uint")
          }
          newPoolA := poolA.Add(value)
          newPoolB := poolA.Mul(poolB).Div(poolB)
          tokensOut := poolB.Sub(newPoolB)

          reg := types.NewRegistry
          reg.RegiserType(&)
          return types.NewDynamicList(types.NewRegistry())
        },
      },
      &functions.Overload{
        Operator: "amm_module_swap_effect_constructor",
        Function: func(args ...ref.Val) ref.Val {
          if len(args) != 3 {
            return types.NewErr("invalid number of argument given to amm_module_swap_effect_constructor")
          }
          acc, ok := args[0].(Account)
          if !ok {
            return types.ValOrErr(args[0], "type error on argument 0: expected type Account")
          }
          value, ok := args[1].(types.Uint)
          if !ok {
            return types.ValOrErr(args[1], "type error on argument 1: expected type uint")
          }

          reg := types.NewRegistry()
          reg.RegisterType(&EffectAMMSwap{})
          return NewMessage(&EffectAMMSwap{
            Account: acc,
            Value: value,
          })
        },
      }
    },
  }
}
```
In this example, the function "amm_curve_calculation" takes three parameters(size of each pool and requested swap value)
and returns the appropriate result state. Note that this function will be not practically useful as it would be easier
to write simple functions such as swap in cel expression, and cel does not support multi variable return. 

The function "amm_module_swap_effect_constructor" is however a useful function that constructs an effect message type.
The effect will be passed to the amm module which will handle the effect in a predefined way.

## Execution process

1. An external actor provides(Cell ID, Expression ID, Expression Input)
2. The module retrieves the cell, the EnvType, and the expression.
3. Retrieve and deserialize the state using the EnvType.
4. Setup cel Env using EnvType and the state.
5. Provide the state and the user input to the expression, evaluate it.
6. Return the output or process the effects.

## Memo

- CellAccount
