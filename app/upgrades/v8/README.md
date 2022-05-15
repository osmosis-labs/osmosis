# V8 Upgrade

The v8 upgrade is an emergency upgrade coordinated according to osmosis governance proposals [225](https://www.mintscan.io/osmosis/proposals/225), [226](https://www.mintscan.io/osmosis/proposals/226).   And thus by implication of 225, incentive proposals [222](https://www.mintscan.io/osmosis/proposals/222), [223](https://www.mintscan.io/osmosis/proposals/223), and [224](https://www.mintscan.io/osmosis/proposals/224).

## Adjusting Incentives for 222, 223, 224
Like the weekly Gauge Weight updates, the implementations for these proposals simply modify the weights of gauges between pools in the Upgrade:

 * `ApplyProp222Change`
 * `ApplyProp223Change`
 * `ApplyProp224Change`

The specification of Minimum and Maximum values will be applied to the spreadsheet that is shared in each Weekly update.

## UnPoolWhitelistedPool for 226
The implementation of 226 will introduce a new method for unpooling:

`UnPoolWhitelistedPool`

Let's review the states a position in a pool may be to be able to understand the unpooling process better.  Coins are pooled together to form shares of a GAMM.  These may be locked for a period of time, to receive addtional incentives.  Finally, locked tokens may enter into Superfluid Delegations.

```
  ┌─────────────────────────┐   ┌─────────────────────────┐
  │  sdk.Coin               │   │  sdk.Coin               │
  │  Denom:  UST            │   │  Denom:  uOSMO          │
  │  Amount: 5.647          │   │  Amount: 1              │
  └───────────┬─────────────┘   └───────────┬─────────────┘
              │                             │
  ┌───────────▼─────────────────────────────▼─────────────┐
  │                       JoinPool()                      │
  └───────────────────────────┬───────────────────────────┘
                              │
                 ┌────────────▼────────────┐
                 │  sdk.Coin               │
                 │  Denom: GAMM            │
                 │  Amount: 100000         │
                 └────────────┬────────────┘
                              │
  ┌───────────────────────────▼───────────────────────────┐
  │                      LockTokens()                     │
  └───────────────────────────┬───────────────────────────┘
                              │
                 ┌────────────▼────────────┐
                 │  types.PeriodLock       │
                 └────────────┬────────────┘
                              │
  ┌───────────────────────────▼───────────────────────────┐
  │                  SuperfluidDelegate()                 │
  └───────────────────────────┬───────────────────────────┘
                              │
                 ┌────────────▼────────────┐
                 │ types.SuperfluidAsset   │
                 └─────────────────────────┘

```
### Unpooling Steps
To unpool, we'll need to carefully consider each of these concepts above.  For instance, a user may have already begun unbonding.

We will start with the most deeply locked assets, and iteratively unroll them until we end up with individual sdk.Coin entities, some of which may be locked.

In the code, the following comment block may be found:
```
	// 0) Check if its for a whitelisted unpooling poolID
	// 1) Consistency check that lockID corresponds to sender, and contains correct LP shares. (Should also be validated by caller)
	// 2) Get remaining duration on the lock.
	// 3) If superfluid delegated, superfluid undelegate
	// 4) Break underlying lock. This will clear any metadata if things are superfluid unbonding
	// 5) ExitPool with these unlocked LP shares
	// 6) Make 1 new lock for every asset in collateral. Many code paths need this assumption to hold
	// 7) Make new lock begin unlocking
```