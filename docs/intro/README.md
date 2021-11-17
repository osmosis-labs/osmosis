# What is Osmosis?

Osmosis is an advanced automated market maker (AMM) protocol that allows developers to build customized AMMs with sovereign liquidity pools. Built using the Cosmos SDK, Osmosis utilizes Inter-Blockchain Communication (IBC) to enable cross-chain transactions.

Osmosis allows users to launch liquidity pools with unique parameters, like bonding curves and multi-weighted asset pools. The incentive structure of Osmosis is also adaptable. Governance implements liquidity reward (LP) rewards for specific pools, allowing for strategically targeted incentives.

Osmosis is a fair-launched, customizable automated market maker for interchain assets that allows the creation and management of non-custodial, self-balancing, interchain token index similar to one of Balancer.

Inspired by [Balancer](http://balancer.finance/whitepaper) and Sunny Aggarwal's '[DAOfying Uniswap Automated Market Maker Pools](https://www.sunnya97.com/blog/daoifying-uniswap-automated-market-maker-pools)', the goal for Osmosis is to provide the best-in-class tools that extend the use of AMMs within the Cosmos ecosystem beyond traditional token swap-type use cases. Bonding curves, while have found its primary use case in decentralized exchange mechanisms, its potential use case can be further extended through the customizability that Osmosis offers. Through the customizability offered by Osmosis such custom-curve AMMs, dynamic adjustments of swap fees, multi-token liquidity pools–the AMM can offer decentralized formation of token fundraisers, interchain staking, options market, and more for the Cosmos ecosystem.

Whereas most Cosmos zones have focused their incentive scheme on the delegators, Osmosis attempts to align the interests of multiple stakeholders of the ecosystem such as LPs, DAO members, as well as delegators. One mechanism that is introduced is how staked liquidity providers have sovereign ownership over their pools, and through the pool governance process allow them to adjust the parameters depending on the pool’s competition and market conditions. Osmosis is a sovereign Cosmos zone that derives its sovereignty not only from its application-specific blockchain architecture but also the collective sovereignty of the LPs that has aligned interest to different tokens that they are providing liquidity for.

## Why Osmosis?

### On customizability of liquidity pools
Most major AMMs limit the changeable parameters of liquidity pools. For example, Uniswap only allows the creation of a two-token pool of equal ratio with the swap fee of 0.3%. The simplicity of Uniswap protocol allowed quick onboarding of the average user that previously had little to no experience in market making.

However, as the DeFi market size grows and market participants such as arbitrageurs and liquidity providers mature, the need for liquidity pools to react to market conditions becomes apparent. The optimal swap fee for a AMM trade may depend on various factors such as block times, slippage, transaction fee, market volatility and more. There is no one-size-fits-all solution as the mix of characteristics of blockchain protocol, tokens in the liquidity pool, market conditions, and others can change the optimal strategy for the liquidity providers and the market makers to carry out.

The tools Osmosis provides allow the market participants to self-identify opportunities and allow them to react by adjusting the various parameters. An optimal equilibrium between fee and liquidity can be reached through autonomous experiments and iterations, rather than a setting a centrally planned 'most acceptable compromise' value. This extends the addressable market for AMMs and bonding curves to beyond simple token swaps, as limitation on the customizability of liquidity pools may have been the inhibiting factor for more experimental use-cases of AMMs.

### Self-governing liquidity pools
As important as the ability to change the parameters of a liquidity pool is, the feature would mean very little without a method to coordinate a decision amongst the stakeholders. The pool governance feature of Osmosis allows a diverse spectrum of liquidity pools with risk tolerance and strategies to not only exist, but evolve.

In Osmosis, the liquidity pool shares are not only used to calculate the fractional ownership of a liquidity pool, but also the right to participate in the strategic decision making of the liquidity pool as well. To incentivize long-term liquidity commitment, shares must be locked up for an extended period. Longer term commitments are awarded by additional voting power / additional liquidity mining revenue. The long-term liquidity commitment by the liquidity providers prevent the impact of potential vampire attacks, where ownership of the shares are delegated and potentially used to migrate liquidity to an external AMM. This provides equity of power amongst liquidity providers, where those with greater skin-in-the-game are given their rightful power to steer the strategic direction of its pool in proportion to the risk they are taking with their assets.

As AMMs mostly guarantee a level of constant total value output, those who may disagree with the changes made to the pool are able to withdraw their funds with little to no loss of their principals. As Osmosis expects the market to self-discover the optimal value of each adjustable parameter, if a significant dissenting opinion exists–they are able to start a competing liquidity pool with their own strategy.

### AMM as serviced infrastructure
The number and complexity of decentralized financial products are consistently increasing. Instruments such as pegged assets, derivatives, options, and tokenized leveraged positions each have their own characteristics that produce optimal market efficiency when paired with the correct bonding curve. That being said, the traditional notion of AMMs have evolved around putting the AMM first, and the financial product being traded second.

As AMMs substantially increase the market accessibility for these instruments, assets with diverse characteristics either had to:
1. Compromise efficiency and trade on existing AMMs with non-optimal bonding curves or
2. Take on the massive task of building one's own AMM that is able to maximize efficiency

To solve this issue, Osmosis introduces the idea of an 'AMM as a serviced infrastructure'. Fairly often, adjustment of the value function and a few additional parameters are all that's needed to provide a highly-efficient, highly-accessible AMM for the majority of decentralized financial instruments. By providing the ability for the creator of the pool to simply define the bonding curve value function and reuse the majority of the key AMM infrastructure, the barrier to creating a tailor-made and efficient automated market maker can be reduced.


## AMM

Automated market makers (AMMs) are decentralized finance protocols that allow for the swapping of assets without a centralized intermediary. Smart contracts replace trading desks and order books in "making the market."
Trades are executed using assets from liquidity pools. Users create pools for specific tokens and deposit assets into them. Users who supply assets to a pool are called liquidity providers (LPs).
AMM pools are permissionless, meaning a user can make a pool for any asset. In Osmosis, pool creators are able to customize the transaction fees and exit fees paid by liquidity providers when withdrawing assets from the pool.
Permissionless pools are key to decentralization, but they also create risks. Some users list fake tokens, hoping to trick others into buying the wrong asset. A common version of this scam is a token with a slight mispelling of a popular token (e.g., OSMOO). It is very important to make sure one is purchasing the correct asset before executing a trade.

## LP Tokens

When users deposit assets into a liquidity pool, they receive LP tokens. These tokens represent their share of the total pool.
For example, if Pool #1 is the OSMO<>ATOM pool, users can deposit OSMO and ATOM tokens into the pool and receive back Pool1 share tokens. These tokens do not correspond to an exact quantity of tokens, but rather the proportional ownership of the pool.
When users remove their liquidity from the pool, they get back the percentage of liquidity that their LP tokens represent.
Since buying and selling from the pool changes the quantities of assets within a pool, users are highly unlikely to withdraw the same amount of each token that they initially deposited. They usually receive more of one and less of another, based on the trades executed from the pool.



## Liquidity Mining

Liquidity mining (also called yield farming) is when users earn tokens for providing liquidity to a DeFi protocol. This mechanism is used to offset the impermanent loss experienced by LPs. Liquidity mining rewards create an additional incentive for LPs besides transaction fees. These rewards are particularly useful for nascent protocols. Liquidity mining helps to bootstrap initial liquidity, facilitating increased usage and more fees for LPs.
Information on Osmosis' incentive mining program can be found in this [section](https://osmosis.gitbook.io/o/osmo/token-issuance/liquidity-rewards).
[IMG1] [IMG2]

## Impermanent Loss
Liquidity providers earn through fees and special pool rewards. However, they are also risking a scenario in which they would have been better off holding the assets rather than supplying them. This outcome is called impermanent loss.
Impermanent loss is the difference in net worth between HODLing and LPing. Liquidity mining helps to offset impermanent loss for LPs.
When the price of the assets in the pool change at different rates, LPs end up owning larger amounts of the asset that increased less in price (or decreased more in price). For example, if the price of OSMO moons relative to ATOM, LPs in the OSMO-ATOM pool end up with larger portions of the less valuable asset (ATOM).
[IMG3]
Impermanent loss is mitigated in part by the transaction fees earned by LPs. When the profits made from swap fees outweigh an LP’s impermanent loss, the pool is self-sustainable.
To further offset impermanent loss, particularly in the early stages of a protocol when volatility is high, AMMs utilize liquidity mining rewards. Liquidity rewards bootstrap the ecosystem as usage and fee revenues are still ramping up.
Osmosis also has many new features and innovations in development to decrease impermanent loss as well.

## Long-Term Liquidity
Liquidity mining rewards tend to attract short-term “mercenary farmers” who quickly deposit and withdraw their liquidity after harvesting the yield. These farmers are only interested in the speculative value of the governance tokens that they are earning. They usually bounce between protocols in search of the best yield.
Mercenary farmers often create the mirage of protocol adoption, but when these farmers leave, it results in significant liquidity volatility. Users of the AMM have difficulty executing trades without encountering slippage. Therefore, long-term liquidity is crucial to the success of an AMM.
Osmosis’ design includes two mechanisms to incentivize long-term liquidity: [exit fees](https://osmosis.gitbook.io/o/liquidity-providing/fees) and [bonded liquidity gauges](https://osmosis.gitbook.io/o/liquidity-providing/blg).

## IBC

The inter-blockchain communication protocol (IBC) creates communication between independent blockchains. IBC achieves this by specifying a set of structures that can be implemented by any distributed ledger that satisfies a small number of requirements.
IBC facilitates cross-chain applications for token transfers, swaps, multi-chain contracts, and data sharding. At launch, Osmosis utilizes IBC for token transfers. Over time, Osmosis will add new features that are made possible through IBC.
