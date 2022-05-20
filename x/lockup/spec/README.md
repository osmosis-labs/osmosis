# Lockup

## Abstract

Lockup module provides an interface for users to lock tokens (also known as bonding) into the module to get incentives.

After tokens have been added to a specific pool and turned into LP shares through the GAMM module, users can then lock these LP shares with a specific duration in order to begin earing rewards.

To unlock these LP shares, users must trigger the unlock timer and wait for the unlock period that was set initially to be completed. After the unlock period is over, users can turn LP shares back into their respective share of tokens.

This module provides interfaces for other modules to iterate the locks efficiently and grpc query to check the status of locked coins.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Keeper](05_keeper.md)**\
6. **[Hooks](06_hooks.md)**\
7. **[Queries](07_queries.md)**\
8. **[Params](08_params.md)**
9. **[Endblocker](09_endblocker.md)**

## Overview 

There are currently three incentivize lockup periods; `1 day` (24h), `1 week` (168h), and `2 weeks` (336h). When locking tokens in the 2 week period, the liquidity provider is effectively earning rewards for a combination of the 1 day, 1 week, and 2 week bonding periods.

The 2 week period refers to how long it takes to unbond the LP shares. The liquidity provider can keep their LP shares bonded to the 2 week lockup period indefinitely. Unbonding is only required when the liquidity provider desires access to the underlying assets.

If the liquidity provider begins the unbonding process for their 2 week bonded LP shares, they will earn rewards for all three bonding periods during the first day of unbonding.

After the first day passes, they will only receive rewards for the 1 day and 1 week lockup periods. After seven days pass, they will only receive the 1 day rewards until the 2 weeks is complete and their LP shares are unlocked. The below chart is a visual example of what was just explained.

<br/>
<p style="text-align:center;">
<img src="/img/bonding.png" height="300"/>
</p>

</br>
</br>

## Transactions

### lock-tokens

Bond tokens in a LP for a set duration

```sh
osmosisd tx lockup lock-tokens [tokens] --duration --from --chain-id
```

::: details Example

To lockup `15.527546134174465309gamm/pool/3` tokens for a `one day` bonding period from `WALLET_NAME` on the osmosis mainnet:

```bash
osmosisd tx lockup lock-tokens 15527546134174465309gamm/pool/3 --duration="24h" --from WALLET_NAME --chain-id osmosis-1
```

To lockup `25.527546134174465309gamm/pool/13` tokens for a `one week` bonding period from `WALLET_NAME` on the osmosis testnet:

```bash
osmosisd tx lockup lock-tokens 25527546134174465309gamm/pool/13 --duration="168h" --from WALLET_NAME --chain-id osmo-test-4
```

To lockup `35.527546134174465309 gamm/pool/197` tokens for a `two week` bonding period from `WALLET_NAME` on the osmosis mainnet:

```bash
osmosisd tx lockup lock-tokens 35527546134174465309gamm/pool/197 --duration="336h" --from WALLET_NAME --chain-id osmosis-1
```
:::


### begin-unlock-by-id

Begin the unbonding process for tokens given their unique lock ID

```sh
osmosisd tx lockup begin-unlock-by-id [id] --from --chain-id
```

::: details Example

To begin the unbonding time for all bonded tokens under id `75` from `WALLET_NAME` on the osmosis mainnet:

```bash
osmosisd tx lockup begin-unlock-by-id 75 --from WALLET_NAME --chain-id osmosis-1
```
:::
::: warning Note
The ID corresponds to the unique ID given to your lockup transaction (explained more in lock-by-id section)
:::

### begin-unlock-tokens

Begin unbonding process for all bonded tokens in a wallet

```sh
osmosisd tx lockup begin-unlock-tokens --from --chain-id
```

::: details Example

To begin unbonding time for ALL pools and ALL bonded tokens in `WALLET_NAME` on the osmosis mainnet:


```bash
osmosisd tx lockup begin-unlock-tokens --from=WALLET_NAME --chain-id=osmosis-1 --yes
```
:::

## Queries

### account-locked-beforetime

Query an account's unlocked records after a specified time (UNIX) has passed

In other words, if an account unlocked all their bonded tokens the moment the query was executed, only the locks that would have completed their bond time requirement by the time the `TIMESTAMP` is reached will be returned.

::: details Example

In this example, the current UNIX time is `1639776682`, 2 days from now is approx `1639971082`, and 15 days from now is approx `1641094282`.

An account's `ADDRESS` is locked in both the `1 day` and `1 week` gamm/pool/3. To query the `ADDRESS` with a timestamp 2 days from now `1639971082`:

```bash
osmosisd query lockup account-locked-beforetime ADDRESS 1639971082
```

In this example will output the `1 day` lock but not the `1 week` lock:

```bash
locks:
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

If querying the same `ADDRESS` with a timestamp 15 days from now `1641094282`:

```bash
osmosisd query lockup account-locked-beforetime ADDRESS 1641094282
```

In this example will output both the `1 day` and `1 week` lock:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```
:::


### account-locked-coins

Query an account's locked (bonded) LP tokens

```sh
osmosisd query lockup account-locked-coins [address]
```

:::: details Example

```bash
osmosisd query lockup account-locked-coins osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

An example output:

```bash
coins:
- amount: "413553955105681228583"
  denom: gamm/pool/1
- amount: "32155370994266157441309"
  denom: gamm/pool/10
- amount: "220957857520769912023"
  denom: gamm/pool/3
- amount: "31648237936933949577"
  denom: gamm/pool/42
- amount: "14162624050980051053569"
  denom: gamm/pool/5
- amount: "1023186951315714985896914"
  denom: gamm/pool/9
```
::: warning Note
All GAMM amounts listed are 10^18. Move the decimal place to the left 18 places to get the GAMM amount listed in the GUI.

You may also specify a --height flag to see bonded LP tokens at a specified height (note: if running a pruned node, this may result in an error)
:::
::::

### account-locked-longer-duration

Query an account's locked records that are greater than or equal to a specified lock duration

```sh
osmosisd query lockup account-locked-longer-duration [address] [duration]
```

::: details Example

Here is an example of querying an `ADDRESS` for all `1 day` or greater bonding periods:

```bash
osmosisd query lockup account-locked-longer-duration osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h
```

An example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```
:::


### account-locked-longer-duration-denom

Query an account's locked records for a denom that is locked equal to or greater than the specified duration AND match a specified denom

```sh
osmosisd query lockup account-locked-longer-duration-denom [address] [duration] [denom]
```

::: details Example

Here is an example of an `ADDRESS` that is locked in both the `1 day` and `1 week` for both the gamm/pool/3 and gamm/pool/1, then queries the `ADDRESS` for all bonding periods equal to or greater than `1 day` for just the gamm/pool/3:

```bash
osmosisd query lockup account-locked-longer-duration-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h gamm/pool/3
```

An example output:

```bash
locks:
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

As shown, the gamm/pool/3 is returned but not the gamm/pool/1 due to the denom filter.
:::


###  account-locked-longer-duration-not-unlocking

Query an account's locked records for a denom that is locked equal to or greater than the specified duration AND is not in the process of being unlocked

```sh
osmosisd query lockup account-locked-longer-duration-not-unlocking [address] [duration]
```

::: details Example

Here is an example of an `ADDRESS` that is locked in both the `1 day` and `1 week` gamm/pool/3, begins unlocking process for the `1 day` bond, and queries the `ADDRESS` for all bonding periods equal to or greater than `1 day` that are not unbonding:

```bash
osmosisd query lockup account-locked-longer-duration-not-unlocking osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h
```

An example output:

```bash
locks:
- ID: "571839"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

The `1 day` bond does not show since it is in the process of unbonding.
:::


### account-locked-pasttime

Query the locked records of an account with the unlock time beyond timestamp (UNIX)

```bash
osmosisd query lockup account-locked-pasttime [address] [timestamp]
```

::: details Example

Here is an example of an account that is locked in both the `1 day` and `1 week` gamm/pool/3. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`)

```bash
osmosisd query lockup account-locked-pasttime osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082
```

The example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` lock ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input.
:::


### account-locked-pasttime-denom

Query the locked records of an account with the unlock time beyond timestamp (unix) and filter by a specific denom

```bash
osmosisd query lockup account-locked-pasttime-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 [timestamp] [denom]
```

::: details Example

Here is an example of an account that is locked in both the `1 day` and `1 week` gamm/pool/3 and `1 day` and `1 week` gamm/pool/1. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`) and filters for gamm/pool/3

```bash
osmosisd query lockup account-locked-pasttime-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082 gamm/pool/3
```

The example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` lock ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input. Additionally, neither of the `1 day` or `1 week` lock IDs for gamm/pool/1 showed due to the denom filter.
:::


### account-locked-pasttime-not-unlocking

Query the locked records of an account with the unlock time beyond timestamp (unix) AND is not in the process of unlocking

```sh
osmosisd query lockup account-locked-pasttime [address] [timestamp]
```

::: details Example

Here is an example of an account that is locked in both the `1 day` and `1 week` gamm/pool/3. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`) AND is not unlocking:

```bash
osmosisd query lockup account-locked-pasttime osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082
```

The example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` lock ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input. Additionally, if ID 572027 were to begin the unlocking process, the query would have returned blank.
:::


### account-unlockable-coins

Query an address's LP shares that have completed the unlocking period and are ready to be withdrawn

```bash
osmosisd query lockup account-unlockable-coins ADDRESS
```



### account-unlocking-coins

Query an address's LP shares that are currently unlocking

```sh
osmosisd query lockup account-unlocking-coins [address]
```

::: details Example

```bash
osmosisd query lockup account-unlocking-coins osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Example output:

```bash
coins:
- amount: "15527546134174465309"
  denom: gamm/pool/3
```
:::


### lock-by-id

Query a lock record by its ID

```sh
osmosisd query lockup lock-by-id [id]
```

::: details Example

Every time a user bonds tokens to an LP, a unique lock ID is created for that transaction.

Here is an example viewing the lock record for ID 9:

```bash
osmosisd query lockup lock-by-id 9
```

And its output:

```bash
lock:
  ID: "9"
  coins:
  - amount: "2449472670508255020346507"
    denom: gamm/pool/2
  duration: 336h
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo16r39ghhwqjcwxa8q3yswlz8jhzldygy66vlm82
```

In summary, this shows wallet `osmo16r39ghhwqjcwxa8q3yswlz8jhzldygy66vlm82` bonded `2449472.670 gamm/pool/2` LP shares for a `2 week` locking period.
:::


### module-balance

Query the balance of all LP shares (bonded and unbonded)

```sh
osmosisd query lockup module-balance
```

::: details Example

```bash
osmosisd query lockup module-balance
```

An example output:

```bash
coins:
- amount: "118851922644152734549498647"
  denom: gamm/pool/1
- amount: "2165392672114512349039263626"
  denom: gamm/pool/10
- amount: "9346769826591025900804"
  denom: gamm/pool/13
- amount: "229347389639275840044722315"
  denom: gamm/pool/15
- amount: "81217698776012800247869"
  denom: gamm/pool/183
- amount: "284253336860259874753775"
  denom: gamm/pool/197
- amount: "664300804648059580124426710"
  denom: gamm/pool/2
- amount: "5087102794776326441530430"
  denom: gamm/pool/22
- amount: "178900843925960029029567880"
  denom: gamm/pool/3
- amount: "64845148811263846652326124"
  denom: gamm/pool/4
- amount: "177831279847453210600513"
  denom: gamm/pool/42
- amount: "18685913727862493301261661338"
  denom: gamm/pool/5
- amount: "23579028640963777558149250419"
  denom: gamm/pool/6
- amount: "1273329284855460149381904976"
  denom: gamm/pool/7
- amount: "625252103927082207683116933"
  denom: gamm/pool/8
- amount: "1148475247281090606949382402"
  denom: gamm/pool/9
```
:::


### module-locked-amount

Query the balance of all bonded LP shares

```sh
osmosisd query lockup module-locked-amount
```

::: details Example

```bash
osmosisd query lockup module-locked-amount
```

An example output:

```bash

  "coins":
    {
      "denom": "gamm/pool/1",
      "amount": "247321084020868094262821308"
    },
    {
      "denom": "gamm/pool/10",
      "amount": "2866946821820635047398966697"
    },
    {
      "denom": "gamm/pool/13",
      "amount": "9366580741745176812984"
    },
    {
      "denom": "gamm/pool/15",
      "amount": "193294911294343602187680438"
    },
    {
      "denom": "gamm/pool/183",
      "amount": "196722012808526595790871"
    },
    {
      "denom": "gamm/pool/197",
      "amount": "1157025085661167198918241"
    },
    {
      "denom": "gamm/pool/2",
      "amount": "633051132033131378888258047"
    },
    {
      "denom": "gamm/pool/22",
      "amount": "3622601406125950733194696"
    },
...

```

NOTE: This command seems to only work on gRPC and on CLI returns an EOF error.
:::



### output-all-locks 

Output all locks into a json file

```sh
osmosisd query lockup output-all-locks [max lock ID]
```

:::: details Example

This example command outputs locks 1-1000 and saves to a json file:

```bash
osmosisd query lockup output-all-locks 1000
```
::: warning Note
If a lockup has been completed, the lockup status will show as "0" (or successful) and no further information will be available. To get further information on a completed lock, run the lock-by-id query.
:::
::::


### total-locked-of-denom

Query locked amount for a specific denom in the duration provided

```sh
osmosisd query lockup total-locked-of-denom [denom] --min-duration
```

:::: details Example

This example command outputs the amount of `gamm/pool/2` LP shares that locked in the `2 week` bonding period:

```bash
osmosisd query lockup total-locked-of-denom gamm/pool/2 --min-duration "336h"
```

Which, at the time of this writing outputs `14106985399822075248947045` which is equivalent to `14106985.3998 gamm/pool/2`

NOTE: As of this writing, there is a bug that defaults the min duration to days instead of seconds. Ensure you specify the time in seconds to get the correct response.
:::
