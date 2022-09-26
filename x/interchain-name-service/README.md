# Interchain Name Service

The `Interchain Name Service` module allows for the mapping of [interchain accounts](https://github.com/cosmos/interchain-accounts-demo) to a human-readable name, starting with Osmosis [Bech32](https://docs.cosmos.network/master/basics/accounts.html) addresses.

# Fee Mechanism

A property tax mechanism that starts out as fixed but adjusts annually based on market demand ensures ensure efficient allocation of interchain domains. This promotes healthy name-buying activity to bootstrap initial activity but prevents inefficent, rent-seeking squatting for high-demand names in the long term.

Note when we say **start of the year** or similar terminology below, we don't mean the start of the _calendar_ year, i.e. 00:00 January 1st, rather the beginning the current year-long tax term. For example, if a name was minted on April 1st, 2022, the beginning of the next year would be April 1st, 2023, _not_ January 1st, 2023.

We start with definitions of relevant constants set during the instantiation of our contract.

- $p_{\text{mint}}$ - the price of initially creating and owning a name, expressed as `mint_price`
- $r$ - the annual property tax as a percentage of the market valuation of a name, expressed as basis points in `annual_tax_bps`
- $\Delta t_g$ - the grace period where only the owner may bid on her name, expressed in seconds as `owner_grace_period`

Other formal variables for reference are defined as

- $p_i(x)$ - the current valuation of a name $x$
- $p^*_{i,j}(x)$ - a bid for $x$ at year $i$ by user $j$
- $\tau_i(x)$ - the property tax paid for owning $x$ at year $i$

When a user (Alice) first buys a domain $x_a$, she sends $p_{\text{mint}}(1 + ry)$ in escrow to indicate her intention to own the domain for $y$ years.

At any year $i$ outside the grade period $\Delta t_g$, Bob (who we will denote with $b$) may bid $p^*_{i,b}$ for ownership of $x_a$, sending $p_{i,b}(1 + ry')$ in escrow to indicate his intention to own the domain for $y'$ years. At any point, Alice may **accept** a bid for $x_a$, receiving $p^*_{i,b}$ and all unpaid rent refunded from the contract.

The name's annual tax is calculated as a percentage of its valuation, which is the maximum amount anyone is willing to bid to own it.

To formalize this a bit, let's call the year that the Alice buys the domain year $0$. At the beginning of any given year $i: i>0$, the valuation of $x_a$ is

$$
p_i = \max_{k : k < i} \left( \max_j \, p^*_{k,j}(x_a) \right)
,
$$

and the tax for year $i$ would be

$$
\tau_i = rp_i
.
$$

The **challenge period** for a year will last from the start of the year $t = 0$ to $t = \Delta t_\text{year} - \Delta t_g$, where $\Delta t_\text{year}$ is the duration of a year. Bids from non-owners are only alllowed during the challenge period.

Afterwards, the **grace period** will last from $t=\Delta t_\text{year} - \Delta t_g$ to $t=\Delta t_\text{year}$. During the grace period, only the owner can make bids on her name. If the property tax from the current valuation exceeds the tax she is currently paying, she can either

1. Match (or even exceed) the current best bid, increasing her taxes for the next year or
1. Sell, and the name will transfer to the user who made the highest bid. As mentioned above, she will be able to get back all "unpaid tax" in escrow and the user's bid.

Thanks to [Vitalik's article](https://vitalik.eth.limo/general/2022/09/09/ens.html) for the thought-provoking inspiration and to @AlpinYukseloglu + @ValarDragon for ironing out more of the details.

# Local setup

1. Follow the instructions to install and run [localosmosis](https://docs.osmosis.zone/developing/dapps/get_started/cosmwasm-localosmosis.html#setup-localosmosis).

2. Install [beaker](https://docs.osmosis.zone/developing/tools/beaker/#installation) and then setup the initial rust project.

```
cd x/interchain-name-service
cargo build
```

3. Compile, deploy, and instantiate the `name-service` contract.

```
beaker wasm deploy name-service --signer-account test1 --no-wasm-opt --raw '{"required_denom":"uosmo","mint_price":"200","annual_tax_bps":"100", "owner_grace_period":"7776000"}'
```

4. Execute example transactions on localosmosis!

```
beaker wasm execute name-service --raw '{"register":{"name":"alice.ibc","years":"5"}}' --signer-account test1 --funds 200uosmo
```

```
beaker wasm query name-service --raw '{"resolve_record": {"name": "alice.ibc"}}'
```
