```html
<!--
order: 9
-->
```

# Endblocker

## Withdraw tokens after unlock time mature

Once time is over, endblocker withdraw coins from matured locks and
coins are sent from lockup `ModuleAccount`.

**State modifications:**

- Fetch all unlockable `PeriodLock`s that `Owner` has not withdrawn
    yet
- Remove `PeriodLock` records from the state
- Transfer the tokens from lockup `ModuleAccount` to the
    `MsgUnlockTokens.Owner`.

## Remove synthetic locks after removal time mature

For synthetic lockups, no coin movement is made, but lockup record and
reference queues are removed.

**State modifications:**

- Fetch all synthetic lockups that is matured
- Remove `SyntheticLock` records from the state along with reference
    queues
