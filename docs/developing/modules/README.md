---
title: Modules
---

# Modules

<div class="cards twoColumn">
  <a href="spec-epochs.html" class="card">
    <img src="/img/time.svg" class="filter-icon" />
    <div class="title">
      Epochs
    </div>
    <div class="text">
      Allows other modules to be signaled once every period to run their desired function
    </div>
  </a>


  <a href="spec-gamm.html" class="card">
    <img src="/img/handshake.svg" class="filter-icon" />
    <div class="title">
      GAMM
    </div>
    <div class="text">
      Provides the logic to create and interact with liquidity pools on Osmosis
    </div>
  </a>


  <a href="spec-incentives.html" class="card">
    <img src="/img/incentives.svg" class="filter-icon" />
    <div class="title">
      Incentives
    </div>
    <div class="text">
      Creates gauges to provide incentives to users who lock specified tokens for a certain period of time
    </div>
  </a>


  <a href="spec-lockup.html" class="card">
    <img src="/img/lock-bold.svg" class="filter-icon" />
    <div class="title">
      Lockup
    </div>
    <div class="text">
      Bonds LP shares for user-defined locking periods to earn rewards
    </div>
  </a>


  <a href="spec-mint.html" class="card">
    <img src="/img/mint.svg" class="filter-icon" />
    <div class="title">
      Mint
    </div>
    <div class="text">
      Creates tokens to reward validators, incentivize liquidity, provide funds for governance, and pay developers
    </div>
  </a>


  <a href="spec-pool-incentives.html" class="card">
    <img src="/img/pool.svg" class="filter-icon" />
    <div class="title">
      Pool-Incentives
    </div>
    <div class="text">
      Creates individual gauges for every lock duration that exists in a pool
    </div>
  </a>


  <a href="spec-gov.html" class="card">
    <img src="/img/gov.svg" class="filter-icon" />
    <div class="title">
      Gov
    </div>
    <div class="text">
      On-chain governance which allows token holders to participate in a community led decision-making process
    </div>
  </a>
 </div>


## Module Accounts

All modules have their own module account. Think of this as a wallet that can only be controlled by that module. Here are a few with their respective wallet addresses and permissions:

<table><thead><tr>
<th><code>Name</code></th>  
<th><code>Address</code></th> 
<th><code>Permissions</code></th></tr></thead> <tbody>
<tr><td>transfer</td>  
<td>

[osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3](https://www.mintscan.io/osmosis/account/osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3)

</td> 
<td>minter, burner</td></tr> 
<tr><td>bonded_tokens_pool</td>  
<td>

[osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09](https://www.mintscan.io/osmosis/account/osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09)

</td> 
<td>burner, staking</td></tr>

<tr><td>not_bonded_tokens_pool</td>  
<td>

[osmo1tygms3xhhs3yv487phx3dw4a95jn7t7lfqxwe3](https://www.mintscan.io/osmosis/account/osmo1tygms3xhhs3yv487phx3dw4a95jn7t7lfqxwe3)

</td> 
<td>burner, staking</td></tr> 

<tr><td>developer_vesting_unvested</td>  
<td>

[osmo1vqy8rqqlydj9wkcyvct9zxl3hc4eqgu3d7hd9k](https://www.mintscan.io/osmosis/account/osmo1vqy8rqqlydj9wkcyvct9zxl3hc4eqgu3d7hd9k)

</td> 
<td>minter</td></tr> 
<tr><td>gov</td> 
<td>

[osmo10d07y265gmmuvt4z0w9aw880jnsr700jjeq4qp](https://www.mintscan.io/osmosis/account/osmo10d07y265gmmuvt4z0w9aw880jnsr700jjeq4qp)

</td> 
<td>burner</td></tr> 
<tr><td>distribution</td> 
<td>

[osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld](https://www.mintscan.io/osmosis/account/osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld)

</td> 
<td>none</td></tr>
<tr><td>lockup</td>  
<td>

[osmo1njty28rqtpw6n59sjj4esw76enp4mg6g7cwrhc](https://www.mintscan.io/osmosis/account/osmo1njty28rqtpw6n59sjj4esw76enp4mg6g7cwrhc)

</td> 
<td>minter, burner</td></tr> 
<tr><td>incentives</td>  
<td>

[osmo1krxwf5e308jmclyhfd9u92kp369l083wequge6](https://www.mintscan.io/osmosis/account/osmo1krxwf5e308jmclyhfd9u92kp369l083wequge6)

</td> 
<td>minter, burner</td></tr> 
<tr><td>gamm</td>  
<td>

[osmo1c9y7crgg6y9pfkq0y8mqzknqz84c3etr0kpcvj](https://www.mintscan.io/osmosis/account/osmo1c9y7crgg6y9pfkq0y8mqzknqz84c3etr0kpcvj)

</td> 
<td>minter, burner</td></tr> 
<tr><td>mint</td> 
<td>

[osmo1m3h30wlvsf8llruxtpukdvsy0km2kum8q25g3j](https://www.mintscan.io/osmosis/account/osmo1m3h30wlvsf8llruxtpukdvsy0km2kum8q25g3j)

</td> 
<td>minter, burner</td></tr> 
<tr><td>claim</td> 
<td>

[osmo1m5dncvfv7lvpvycr23zja93fecun2kcv226glq](https://www.mintscan.io/osmosis/account/osmo1m5dncvfv7lvpvycr23zja93fecun2kcv226glq)

</td> 
<td>minter</td></tr>
<tr><td>poolincentives</td>  
<td>

[osmo1upfuxznarpja3sywq0tzd2kktg9wv8mcc0rlm9](https://www.mintscan.io/osmosis/account/osmo1upfuxznarpja3sywq0tzd2kktg9wv8mcc0rlm9)

</td> 
<td>none</td></tr> 
<tr><td>fee_collector</td>  
<td>

[osmo17xpfvakm2amg962yls6f84z3kell8c5lczssa0](https://www.mintscan.io/osmosis/account/osmo17xpfvakm2amg962yls6f84z3kell8c5lczssa0)

</td> 
<td>none</td></tr> 
</tbody></table>

## Module Flow

While module functions can be called in many different orders, here is a basic flow of module commands to bring assets onto Osmosis and then add/remove liquidity:

1. (IBC-Transfer) IBC Received
2. (GAMM) Swap Exact Amount In
3. (GAMM) Join Pool
4. (lockup) Lock-tokens
5. (lockup) Begin-unlock-tokens
6. (GAMM) Exit Pool
