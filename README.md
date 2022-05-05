# Osmosis

![Logo!](assets/logo.png)

[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://img.shields.io/badge/repo%20status-Active-green.svg?style=flat-square)](https://www.repostatus.org/#active)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/osmosis-labs/osmosis)
[![Go Report Card](https://goreportcard.com/badge/github.com/osmosis-labs/osmosis?style=flat-square)](https://goreportcard.com/report/github.com/osmosis-labs/osmosis)
[![Version](https://img.shields.io/github/tag/osmosis-labs/osmosis.svg?style=flat-square)](https://github.com/osmosis-labs/osmosis/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/osmosis-labs/osmosis.svg?style=flat-square)](https://github.com/osmosis-labs/osmosis/blob/main/LICENSE)
[![Lines Of Code](https://img.shields.io/tokei/lines/github/osmosis-labs/osmosis?style=flat-square)](https://github.com/osmosis-labs/osmosis)
[![GitHub Super-Linter](https://img.shields.io/github/workflow/status/osmosis-labs/osmosis/Lint?style=flat-square&label=Lint)](https://github.com/marketplace/actions/super-linter)
[![Discord](https://badgen.net/badge/icon/discord?icon=discord&label)](https://discord.gg/osmosis)

Osmosis is a fair-launched, customizable automated market maker for interchain
assets that allows the creation and management of non-custodial, self-balancing,
interchain token index similar to one of Balancer.

Inspired by [Balancer](http://balancer.finance/whitepaper) and Sunny Aggarwal's '[DAOfying Uniswap Automated Market Maker Pools](https://www.sunnya97.com/blog/daoifying-uniswap-automated-market-maker-pools)', the goal for Osmosis is to
provide the best-in-class tools that extend the use of AMMs within the Cosmos
ecosystem beyond traditional token swap-type use cases. Bonding curves, while
have found its primary use case in decentralized exchange mechanisms, its potential
use case can be further extended through the customizability that Osmosis offers.
Through the customizability offered by Osmosis such custom-curve AMMs, dynamic
adjustments of swap fees, multi-token liquidity pools–the AMM can offer decentralized
formation of token fundraisers, interchain staking, options market, and more for
the Cosmos ecosystem.

Whereas most Cosmos zones have focused the  ir incentive scheme on the delegators,
Osmosis attempts to align the interests of multiple stakeholders of the ecosystem
such as LPs, DAO members, as well as delegators. One mechanism that is introduced
is how staked liquidity providers have sovereign ownership over their pools, and
through the pool governance process allow them to adjust the parameters depending
on the pool’s competition and market conditions. Osmosis is a sovereign Cosmos
zone that derives its sovereignty not only from its application-specific blockchain
architecture but also the collective sovereignty of the LPs that has aligned
interest to different tokens that they are providing liquidity for.

## System Requirements

This system spec has been tested by many users and validators and found to be comfortable:

* Quad Core or larger AMD or Intel (amd64) CPU
  * ARM CPUs like the Apple M1 are not supported at this time.
* 64GB RAM (A lot can be in swap)
* 1TB NVMe Storage
* 100MBPS bidirectional internet connection

You can run Osmosis on lower-spec hardware for each component, but you may find that it is not highly performant or prone to crashing.

## Documentation

For the most up to date documentation please visit [docs.osmosis.zone](https://docs.osmosis.zone/)

## Joining the Mainnet

[Please visit the official instructions on how to join the Mainnet here.](https://docs.osmosis.zone/developing/network/join-mainnet.html#install-osmosis-binary)

Thank you for supporting a healthy blockchain network and community by running an Osmosis node!

## Why Osmosis?

### On customizability of liquidity pools

Most major AMMs limit the changeable parameters of liquidity pools. For example,
Uniswap only allows the creation of a two-token pool of equal ratio with the swap
fee of 0.3%. The simplicity of Uniswap protocol allowed quick onboarding of the
average user that previously had little to no experience in market making.

However, as the DeFi market size grows and market participants such as arbitrageurs
and liquidity providers mature, the need for liquidity pools to react to market
conditions becomes apparent. The optimal swap fee for a AMM trade may depend on
various factors such as block times, slippage, transaction fee, market volatility
and more. There is no one-size-fits-all solution as the mix of characteristics of
blockchain protocol, tokens in the liquidity pool, market conditions, and others
can change the optimal strategy for the liquidity providers and the market makers
to carry out.

The tools Osmosis provides allow the market participants to self-identify opportunities
and allow them to react by adjusting the various parameters. An optimal equilibrium
between fee and liquidity can be reached through autonomous experiments and iterations,
rather than a setting a centrally planned 'most acceptable compromise' value. This
extends the addressable market for AMMs and bonding curves to beyond simple token
swaps, as limitation on the customizability of liquidity pools may have been the
inhibiting factor for more experimental use-cases of AMMs.

### Self-governing liquidity pools

As important as the ability to change the parameters of a liquidity pool is, the
feature would mean very little without a method to coordinate a decision amongst
the stakeholders. The pool governance feature of Osmosis allows a diverse spectrum
of liquidity pools with risk tolerance and strategies to not only exist, but evolve.

In Osmosis, the liquidity pool shares are not only used to calculate the fractional
ownership of a liquidity pool, but also the right to participate in the strategic
decision making of the liquidity pool as well. To incentivize long-term liquidity
commitment, shares must be locked up for an extended period. Longer term commitments
are awarded by additional voting power / additional liquidity mining revenue. The
long-term liquidity commitment by the liquidity providers prevent the impact of
potential vampire attacks, where ownership of the shares are delegated and potentially
used to migrate liquidity to an external AMM. This provides equity of power amongst
liquidity providers, where those with greater skin-in-the-game are given their
rightful power to steer the strategic direction of its pool in proportion to the
risk they are taking with their assets.

As AMMs mostly guarantee a level of constant total value output, those who may
disagree with the changes made to the pool are able to withdraw their funds with
little to no loss of their principals. As Osmosis expects the market to self-discover
the optimal value of each adjustable parameter, if a significant dissenting opinion
exists–they are able to start a competing liquidity pool with their own strategy.

### AMM as serviced infrastructure

The number and complexity of decentralized financial products are consistently
increasing. Instruments such as pegged assets, derivatives, options, and tokenized
leveraged positions each have their own characteristics that produce optimal market
efficiency when paired with the correct bonding curve. That being said, the traditional
notion of AMMs have evolved around putting the AMM first, and the financial product
being traded second.

As AMMs substantially increase the market accessibility for these instruments,
assets with diverse characteristics either had to:

1. Compromise efficiency and trade on existing AMMs with non-optimal bonding curves or
2. Take on the massive task of building one's own AMM that is able to maximize efficiency

To solve this issue, Osmosis introduces the idea of an 'AMM as a serviced infrastructure'.
Fairly often, adjustment of the value function and a few additional parameters are
all that's needed to provide a highly-efficient, highly-accessible AMM for the
majority of decentralized financial instruments. By providing the ability for the
creator of the pool to simply define the bonding curve value function and reuse
the majority of the key AMM infrastructure, the barrier to creating a tailor-made
and efficient automated market maker can be reduced.
