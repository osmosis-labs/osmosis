---
title: Module Home
---

# Modules

<div class="cards twoColumn">
  <a href="spec-epochs.html" class="card">
    <img src="/osmosis/img/time.svg" class="filter-blue" />
    <div class="title">
      Epochs
    </div>
    <div class="text">
      Allows other modules to be signaled once every period to run their desired function
    </div>
  </a>


  <a href="spec-gamm.html" class="card">
    <img src="/osmosis/img/handshake.svg" class="filter-blue" />
    <div class="title">
      GAMM
    </div>
    <div class="text">
      Provides the logic to create and interact with liquidity pools on Osmosis
    </div>
  </a>


  <a href="spec-incentives.html" class="card">
    <img src="/osmosis/img/incentives.svg" class="filter-blue" />
    <div class="title">
      Incentives
    </div>
    <div class="text">
      Creates gauges to provide incentives to users who lock specified tokens for a certain period of time
    </div>
  </a>


  <a href="spec-lockup.html" class="card">
    <img src="/osmosis/img/lock-bold.svg" class="filter-blue" />
    <div class="title">
      Lockup
    </div>
    <div class="text">
      Bonds LP shares for user-defined locking periods to earn rewards
    </div>
  </a>


  <a href="spec-mint.html" class="card">
    <img src="/osmosis/img/mint.svg" class="filter-blue" />
    <div class="title">
      Mint
    </div>
    <div class="text">
      Creates tokens to reward validators, incentivize liquidity, provide funds for governance, and pay developers
    </div>
  </a>


  <a href="" class="card">
    <img src="" class="filter-blue" />
    <div class="title">
      Pool-incentives
    </div>
    <div class="text">
      Test
    </div>
  </a>


  <a href="spec-gov.html" class="card">
    <img src="/osmosis/img/gov.svg" class="filter-blue" />
    <div class="title">
      Gov
    </div>
    <div class="text">
      On-chain governance which allows token holders to participate in a community led decision-making process
    </div>
  </a>
 </div>

## Module Flow

While module functions can be called in many different orders, here is a basic flow of module commands to bring assets onto Osmosis and then add/remove liquidity:

1. (IBC-Transfer) IBC Received
2. (GAMM) Swap Exact Amount In
3. (GAMM) Join Pool
4. (lockup) Lock-tokens
5. (lockup) Begin-unlock-tokens
6. (GAMM) Exit Pool