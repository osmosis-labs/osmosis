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

## End of Epoch
Iterate through all active gauges:
 1. find `CurrentReward` using `denom` and `duration`
 2. transfer distributed `reward` from gauge to `CurrentReward`
    - reward_in_epoch = gauge.size / remain_epoch
    - current_reward.reward = current_reward.reward + reward_in_epoch
    - gauge.distributed_coins = gauge.distributed_coins + reward_in_epoch
 3. find if there are any `UnlockingLock`s that has less remaing `endtime` than `duration` of the gauge
    - begin_time = EpochStartTime + Lockable Duration
    - end_time = begin_time + EpochDuration (exclusive)
    - count(UnlockingLocks(begin_time, end_time) ) > 0
 4. if above is `true`, update `HistoricalReward` as follows:
    - find relevant `HistoricalReward`s
      - prev_historical_reward = historical_reward(current_reward.period - 1)
      - curr_historical_reward = historical_reward(current_reward.period)
    - update `reward`
      - stake_per_reward = current_reward.reward / current_reward.stake
      - curr_historical_reward = prev_historical_reward + stake_per_reward

## Estimate Reward
- `Period` is the last epoch number of claimed reward
- `Reward` is the reward accumulated until `Period`
### Lock
1. if `Period` is the same as last epoch number, return `Reward`
2. if not, for all gauges that have less `duration` of lock, return
    - additional_reward = Reward + (AccumReward[currEpoch] - AccumReward[period]) * Coin
### UnlockingLock
1. find epoch range using `duration` of gauge and last epoch
    - start_epoch: period
    - end_epoch: ((gauge.distibuteTo.duration + curr_epoch.time) ≤ lock.endTime) ? curr_epoch : lock.endEpoch
2. determine if there are any rewards
    - isCandidate = (gauge.distibuteTo.duration + epoch.time) ≤ lock.endTime
3. return `Reward`
    - additional_reward = Reward + (AccumReward[currEpoch] - AccumReward[period]) * Coin

## Claim Reward
1. calculate reward (_in a similar manner as in reward estimation_)
2. distribute reward
3. update `Period` to last epoch number
