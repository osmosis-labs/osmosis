# Superfluid staking

Superfluid staking is to use the OSMO tokens put on liquidity pools to be used both for swap & staking.

## Potential problems

### Is it secure to use it both for staking and swap?

In current design, liquidity is not used enough and therefore, using it both for swap and staking won't be a problem.
As we update the pools to use liquidity more efficiently, almost all of those liquidity could be used for swap.
In that case, the amount of OSMO in the pool does not participate in increasing the security of the chain.

### How to determine the amount of OSMO delegation from the pool to validators?

Time Average Osmo Amount per LP token * Risk Rating of LP token

### On slash event, OSMO will be burnt from the pool directly or LP tokens will be burnt?

I think burning LP token would be nicer and pretty easier.
LP token's value could be calculated based on how much OSMO is in it and also calculate burn amount from that.

If burn OSMO, it will cause a problem of price change.

### The reward from staking, it should be withdrawn to the user's wallet on each epoch with yield farming incentives or withdraw it by doing claim?

### Do we really need to select validator for superfluid staking?

What if we think that they are constantly supporting the security of chain and provide specific percentage of fee pool to that rather than dividing by OSMO put the liquidity and the OSMO delegated?

### To start superfluid staking, users should lock LP tokens for 3 weeks

### Superfluid staking will be applied for only secure pairs like ATOM/OSMO or AKT/OSMO

### Base thoughts of how OSMO-X LP token pair secures Osmosis chain

OSMO-X LPs secures OSMO token price by providing worth of X tokens to OSMO pair.
It prevents OSMO token price from going down easily.

The TVL on staking is the security of the chain.
The TVL on liquidity is the security of OSMO token price.

Therefore, the incentives could be provided in the same way with stakers.

What stakers get = OSMO tokens from inflation + OSMO tokens on transaction fee
What LPs get = OSMO yield farming from inflation + trading fees - impermanent loss

What LPs don't get here is transaction fees.
Should we actually provide inflation allocated for stakers to LPs? Won't it let the OSMO stakers leave to become LPs?

Or when we implement superfluid staking, should we just remove allocation percentage for pool-incentives?

### Security of LP token

LP token's security should be counted based on two tokens put on the pool.
For instance, for ATOM/OSMO pair, both ATOM and OSMO participate in increasing the security of Osmosis chain.
I think not only OSMO amount, making both of them to participate in the security would be better.

Let's say ATOM's security level is 3 and OSMO's security level is 1.
And pair has 1million ATOM and 5million OSMO.
Here, let's say 1 LP token is representative of 1ATOM and 5 OSMO.
In this case, security level of an LP token could be 10 not 5.

We will need to track average rating between ATOM/OSMO pair and also volatility between these pairs.
We can calculate like `LP(OSMO)*(1-Volatility)+LP(ATOM)*(1-Volatility)`.

### For superfluid staking, price stability is quite important 

To make the LP token price stable, what Osmosis integrate auto pool rebalancer on endblocker and giving the auto-rebalance rewards to stakers?

Like if price changes are made within the Osmosis pool, just remove it on endblocker, to remove bots that get incentives?
It will ensure that the price is maintained in a block for Osmosis zone and stakers will get higher income.
And also hackers should hack all the relevant pools rather than only a single pool.
And also it will eliminate the competition between bots to earn incentives.

### Feedback from Dev

Better to focus on the code side of imagining how we can re-structure staking reward distribution to look at GAMM shares as well.
