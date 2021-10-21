<!--
order: 8
-->

# F1 Distribution
The basic concept is from [F1 Fee Distribution]( https://drops.dagstuhl.de/opus/volltexte/2020/11974/pdf/OASIcs-Tokenomics-2019-10.pdf) which is implemented in Cosmos SDK (_distribution module_).

## Definition
- `Period`
  - time duration in which bonded stakes remain constant
- `CurrentReward`
  - represents current rewards and current period for a gauge
- `HistoricalReward`
  - represents historical rewards for a gauge. Accumulated reward ratio is the sum from the zeroth period until current period of (rewards/stakes)

## Bonded stake is changed when
- a new lock is created
- additional lock is added to existing lock
- a lock that is being unlocked has less remaining `endtime` than `duration` of gauge

### Lock tokens
- update all `HistoricalReward` related to this lock
- if add tokens to existing lock, the accumulated rewards so far will be recorded

### Unlock tokens
- send rewards for this lock to owner
- in case of partial unlock:
  - update all `HistoricalReward` related to this lock 
  - the accumulated rewards so far will be recorded, not send to owner

## End of Epoch
Iterate through all active gauges:
 1. get `CurrentReward` using `denom` and `duration` of gauge
 2. get the total stakes of locks that can be rewarded
 3. if the total stakes was changed during this period:
    - make a new `HistoricalReward` based on `CurrentReward` 
    - reset `CurrentReward` with new period
    - if `HistoricalReward` has already been updated in this epoch, this process will be skipped
 4. transfer distributed `reward` from gauge to `CurrentReward`

## Estimate Reward for a lock
### Lock
- return the recorded rewards + expected rewards if unlock tokens at now
### UnlockingLock
- return the recorded rewards + expected rewards until `endtime`

## Claim Reward
1. calculate and send reward
2. reset reward data for this lock
