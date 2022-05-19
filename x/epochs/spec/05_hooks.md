# Hooks

<!-- markdownlint-disable MD013 -->
``` {.go}
  // the first block whose timestamp is after the duration is counted as the end of the epoch
  AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
  // new epoch is next block of epoch end block
  BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
```

## How modules receive hooks

On hook receiver function of other modules, they need to filter
`epochIdentifier` and only do executions for only specific
epochIdentifier. Filtering epochIdentifier could be in `Params` of other
modules so that they can be modified by governance. Governance can
change epoch from `week` to `day` as their need.

## Panic isolation

If a given epoch hook panics, its state update is reverted, but we keep
proceeding through the remaining hooks. This allows more advanced epoch
logic to be used, without concern over state machine halting, or halting
subsequent modules.

This does mean that if there is behavior you expect from a prior epoch
hook, and that epoch hook reverted, your hook may also have an issue. So
do keep in mind "what if a prior hook didn't get executed" in the safety
checks you consider for a new epoch hook.
