# Osmosis
### Balancer meets Interchain Staking

## Background
Osmosis is an onchain generalized multi-token automated market maker and a self-balancing weighted portfolio built on the Cosmos-SDK with extendability as a custom staking token design tool for cross-chain incentive alignment, risk diversification, token distribution, and chain valuation by converging DeFi and staking together.

One benefit that a multi-token AMM DEX provides is that it allows a flexible AMM model where token weights, fees, and other parameters can be adjusted according to market conditions and competition. It allows an index fund of Cosmos ecosystem tokens that can be held passively, and also generate interest while providing utility in the form of DEX liquidity within Cosmos.

We extend the idea of the 'index fund' introduced in Balancer's generalized AMM model as a way to use the multi-token pool's LP token as a staking token for a new zone. Similar to how Balancer provided parameterization of the AMM DEX by allowing custom ratios, multiple tokens, and adjustable trading fees, Osmosis leverages the the generalized automated market maker as a mechanism to create a customizable staking token for a new zone.

This is a similar idea to 'interchain staking' introduced by Jae Kwon, but uses a very simple mechanism of Balancer + IBC + Non-native token staking to achieve the same goal.

Osmosis is a 'two birds with one stone' project that:
1. Provide a very flexible AMM DEX to operate within the Cosmos ecosystem
2. Allow cross-chain incentive alignment, risk hedging, fair launches to be programmed into the staking token

## Generalized Automated Market Maker
Similar to Balancer, Osmosis' `x/gamm` is a multi-token, parameterized automated market maker application. Liquidity pools can be created with up to 8 tokens to specified to different weighted ratios. Liquidity pools can function as an AMM DEX similar to Uniswap, where the average users can interact with the liquidity pool to swap tokens.

Currently, the following features are working:
* Token swaps
	* Calculating swap slippage
* Liquidity pool
	* Create new liquidity pools
		* Set ratio between tokens (i.e. 40% ATOM / 50% IRIS / 10% KAVA)
		* Add up to 8 tokens
	* Add liquidity to existing liquidity pools
		* Mint LP tokens which represent a share in the liquidity pool
		* Add liquidity according to the specified ratio
			* i.e. Add $4 ATOM / $5 IRIS / $1 KAVA for the 40% ATOM / 50% IRIS / 10% KAVA pool
		* Add single token liquidity
			* Automatically swapped internally
	* Withdraw from liquidity pool
		* Liquidity Pool receives LP token and gives the % share of tokens in the pool
		* Allow single-token withdrawals by automatically swapping the unwanted tokens internally

### Staking LP Tokens
We are merging the DeFi primitive of multi-token AMM and Cosmos proof-of-stake system by allowing the LP tokens (similar to the utility of the BPT token in Balancer) to be used as a staking token for a new zone.

The intent of this is to allow customization and parameterization of a zone's staking token, similar to how balancer allows customization and parameterization of the AMM LP. Currently, a new Cosmos zone (which only allows native zone staking token to be delegaed) is fully reliant on the economic characteristics of its own staking token for security. By allowing a custom designed LP token to be used as a staking token, the creators of the new zone can essentially 'design' its staking token by customizing the GAMM LP pool which the LP token is used for the zone's staking token.

For example, if I want to create a new zone called Network A which wants to:
1. Align my zone's success to the success of the Hub
2. Reduce volatility of the staking token
3. Have my zone's native token have impact on the voting power of my zone

I can essentially create a '40% ATOM / 30% USDC / 30% zoneToken' pool on Osmosis and use the Osmosis LP token as my zone's staking token.

Because the ATOM accounts for 40% of the voting power of 1 staking token on my zone, the success of ATOMs increase the security of my zone as well. Furthermore, it incentivizes ATOM holders to be more interested in participating in my zone without the common downsides of airdrops (i.e. no skin-in-the-game, price volatility, etc). Also, because 30% of the voting power is derived from a stablecoin my zone's security is relatively more secure against price fluctuations of ATOM tokens and the zone token–which allows a more stable security guarantee. Lastly, because my zone's utlity token can be used as a fraction of the voting power, it provides a level of sovereignty to my zone as well. The best part of this is that the assets that are included as part of the Osmosis pool can be selected according to its characteristics such as economic value, community, fiat liquidity, sovereignty, etc. Furthermore, an additional layer of customizability can be configured by setting each of these asset's Osmosis pool ratios at different weights.

Also, because an LP token can be used to withdraw from the pool, whoever chose to participate in minting the staking token can always choose to exit the pool and receive a portion of their tokens back. This reduces the risk of a token investment going to zero, allowing a more equitable token distribution mechanism.
