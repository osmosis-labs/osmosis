# Solidly Stableswap

Stableswaps are pools that offer low slippage for two assets that are intended to be tightly correlated.
There is a price ratio they are expected to be at, and the AMM offers low slippage around this price.
There is still price impact for each trade, and as the liquidity becomes more lop-sided, the slippage drastically increases.

This package implements the Solidly stableswap curve, namely a CFMM with
invariant: $f(x, y) = xy(x^2 + y^2) = k$

It is generalized to the multi-asset setting as $f(a_1, ..., a_n) = a_1 * ... * a_n (a_1^2 + ... + a_n^2)$

## Choice of curve

{TODO: Include some high level summary of the curve}

## Pool configuration

One key concept, is that the pool has a native concept of

### Scaling factor handling

An important concept thats up to now, not been mentioned is how do we set the expected price ratio.
In the choice of curve section, we see that its the case that when `x_reserves ~= y_reserves`, that spot price is very close to `1`. However, there are a couple issues with just this in practice:

1. Precision of pegged coins may differ. Suppose `1 Foo = 10^12 base units`, whereas `1 WrappedFoo = 10^6 base units`, but `1 Foo` is expected to trade near the price of `1 Wrapped Foo`.
2. Relatedly, suppose theres a token called `TwoFoo` which should trade around `1 TwoFoo = 2 Foo`
3. For staking derivatives, where value accrues within the token, the expected price to concentrate around dynamically changes (very slowly).

To handle these cases, we introduce scaling factors. A scaling factor maps from "raw coin units" to "amm math units", by dividing.
To handle the first case, we would make `Foo` have a scaling factor of `10^6`, and `WrappedFoo` have a scaling factor of `1`.
This mapping is done via `raw coin units / scaling factor`.
We use a decimal object for amm math units, however we still have to be precise about how we round.
We introduce an enum `rounding mode` for this, with three modes: `RoundUp`, `RoundDown`, `RoundBankers`.

The reserve units we pass into all AMM equations would then be computed based off the following reserves:

```python
scaled_Foo_reserves = decimal_round(pool.Foo_liquidity / 10^6, RoundingMode)
descaled_Foo_reserves = scaled_Foo_reserves * 10^6
```

Similarly all token inputs would be scaled as such.
The AMM equations need to each ensure that rounding happens correctly,
for cases where the scaling factor doesn't perfectly divide into the liquidity.
We detail rounding modes and scaling details as pseudocode in the relevant sections of the spec.
(And rounding modes for 'descaling' from AMM eq output to real liquidity amounts, via multiplying by the respective scaling factor)

<!-- TODO come back and revise the scaling factor section for clarity -->

## Algorithm details

The AMM pool interfaces requires implementing the following stateful methods:

```golang
	SwapOutAmtGivenIn(tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error)
	SwapInAmtGivenOut(tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error)

	SpotPrice(baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error)

	JoinPool(tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error)
	JoinPoolNoSwap(tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error)
	ExitPool(numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error)
```

The "constant" part of CFMM's imply that we can reason about all their necessary algorithms from just the CFMM equation. There are still multiple ways to solve each method. We detail below the ways in which we do so. This is organized by first discussing variable substitutions we do, to be in a more amenable form, and then the details of how we implement each method.

<!---  This is true, but however the details of this curve actually lend itself better to using more general solutions for swaps via binary search. (Inspired from https://arxiv.org/abs/2111.13740 ) The remainder of the methods we implement via generic CFMM ideas.

We detail below both what the direct solution via the CFMM equation is, and the solution we use.
Then we show how these are used to give all of the swqp equations. --->

### CFMM function

Most operations we do only need to reason about two of the assets in a pool, and sometimes only one.
We wish to have a simpler CFMM function to work within these cases.
Due to the CFMM equation $f$ being a symmetric function, we can without loss of generality reorder the arguments to the function. Thus we put the assets of relevance at the beginning of the function. So if two assets $x, y$, we write: $f(x,y, a_3, ... a_n) = xy * a_3 * ... a_n (x^2 + y^2 + a_3^2 + ... + a_n^2)$.

We then take a more convenient expression to work with, via variable substition.

<!-- It took several very silly hacks to get github to compile this. Editors in the future, be aware of spacing in the equation wrt how it GH renders -->

$$
\begin{equation}
    v =
    \begin{cases}
      1, & \text{if } n=2 \\
      \prod\negthinspace \negthinspace \thinspace^{n}_{i=3} \space a_i, & \text{otherwise}
    \end{cases}
  \end{equation}
$$

$$
\begin{equation}
    w =
    \begin{cases}
      0, & \text{if}\ n=2 \\
      \sum\negthinspace \negthinspace \thinspace^{n}_{i=3} \space {a_i^2}, & \text{otherwise}
    \end{cases}
  \end{equation}
$$

$$\text{then } g(x,y,v,w) = xyv(x^2 + y^2 + w) = f(x,y, a_3, ... a_n)$$

As a corollary, notice that $g(x,y,v,w) = v * g(x,y,1,w)$, which will be useful when we have to compare before and after quantities. We will use $h(x,y,w) := g(x,y,1,w)$ as short-hand for this.

### Swaps

The question we need to answer for a swap is "suppose I want to swap $a$ units of $x$, how many units $b$ of $y$ would I get out".

Since we only deal with two assets at a time, we can then work with our prior definition of $g$. Let the input asset's reserves be $x$, the output asset's reserves be $y$, and we compute $v$ and $w$ given the other asset reserves, whose reserves are untouched throughout the swap.

First we note the direct way of solving this, its limitation, and then an iterative approximation approach that we implement.

#### Direct swap solution

The method to compute this under 0 swap fee is implied by the CFMM equation itself, since the constant refers to:
$g(x_0, y_0, v, w) = k = g(x_0 + a, y_0 - b, v, w)$. As $k$ is linearly related to $v$, and $v$ is unchanged throughout the swap, we can simplify the equation to be reasoning about $k' = \frac{k}{v}$ as the constant, and $h$ instead of $g$

We then model the solution by finding a function $\text{solve cfmm}(x, w, k') = y\text{ s.t. }h(x, y, w) = k'$.
Then we can solve the swap amount out by first computing $k'$ as $k' = h(x_0, y_0, w)$, and
computing $y_f := \text{solve cfmm}(x_0 + a, w, k')$. We then get that $b = y_0 - y_f$.

So all we need is an equation for $\text{solve cfmm}$! Its essentially inverting a multi-variate polynomial, and in this case is solvable: [wolfram alpha link](https://www.wolframalpha.com/input?i=solve+for+y+in+x+*+y+*+%28x%5E2+%2B+y%5E2+%2B+w%29+%3D+k)

Or if were clever with simplification in the two asset case, we can reduce it to: [desmos link](https://www.desmos.com/calculator/hag1f0wieg).

These functions are a bit complex, which is fine as they are easy to prove correct. However, they are relatively expensive to compute, the latter needs precision on the order of x^4, and requires computing multiple cubic roots.

Instead there is a more generic way to compute these, which we detail in the next subsection.

#### Iterative search solution

Instead of using the direct solution for $\text{solve cfmm}(x, w, k')$, instead notice that $h(x, y, w)$ is an increasing function in $y$.
So we can simply binary search for $y$ such that $h(x, y, w) = k'$, and we are guaranteed convergence within some error bound.

In order to do a binary search, we need bounds on $y$.
The lowest lowerbound is $0$, and the largest upperbound is $\infty$.
The maximal upperbound is obviously unworkable, and in general binary searching around wide ranges is unfortunate, as we expect most trades to be centered around $y_0$.
This would suggest that we should do something smarter to iteratively approach the right value for the upperbound at least.
Notice that $h$ is super-linearly related in $y$, and at most cubically related to $y$.
This means that $\forall c \in \mathbb{R}^+, c * h(x,y,w) < h(x,c*y,w) < c^3 * h(x,y,w)$.
We can use this fact to get a pretty-good initial upperbound guess for $y$ using the linear estimate.
In the lowerbound case, we leave it as lower-bounded by $0$, otherwise we would need to take a cubed root to get a better estimate.

```python
def iterative_search(x_f, y_0, w, k, err_tolerance):
  k_0 = h(x_f, y_0, w)
  lowerbound, upperbound = y_0, y_0
  k_ratio = k_0 / k
  if k_ratio < 1:
    # k_0 < k. Need to find an upperbound. Worst case assume a linear relationship, gives an upperbound
    # We could derive better bounds via reasoning about coefficients in the cubic,
    # however this is deemed as not worth it, since the solution is quite close
    # when we are in the "stable" part of the curve.
    upperbound = ceil(y_0 / k_ratio)
  elif k_ratio > 1:
    # need to find a lowerbound. We could use a cubic relation, but for now we just set it to 0.
    lowerbound = 0
  else:
    return y_0 # means x_f = x_0
  k_calculator = lambda y_est: h(x_f, y_est, w)
  max_iteration_count = 100
  return binary_search(lowerbound, upperbound, k_calculator, k, err_tolerance)

def binary_search(lowerbound, upperbound, approximation_fn, target, max_iteration_count, err_tolerance):
  iter_count = 0
  cur_k_guess = 0
  while (not satisfies_bounds(cur_k_guess, target, err_tolerance)) and iter_count < max_iteration_count:
    iter_count += 1
    cur_y_guess = (lowerbound + upperbound) / 2
    cur_k_guess = approximation_fn(cur_y_guess)

    if cur_k_guess > target:
      upperbound = cur_y_guess
    else if cur_k_guess < target:
      lowerbound = cur_y_guess

  if iter_count == max_iteration_count:
    return Error("max iteration count reached")

  return cur_y_guess
```

Now we want to wrap this binary search into `solve_y`. We changed the API slightly, from what was previously denoted, to have this "y_0" term, in order to derive initial bounds.
What remains is setting the error tolerance. We need two properties:

- The returned value to be within some correctness threshold of the true value
- The returned value to be rounded correctly (always ending with the user having fewer funds to avoid pool drain attacks). Mitigated by swap fees for normal swaps, but needed for 0-fee to be safe.

The error tolerance we set is defined in terms of error in `k`, which itself implies some error in `y`.
An error of `e_k` in `k`, implies an error `e_y` in `y` that is less than `e_k`. We prove this [here](#err_proof) (and show that `e_y` is actually much less than the error in `e_k`, but for simplicity ignore this fact). We want `y` to be within a factor of `10^(-12)` of its true value.
To ensure the returned value is always rounded correctly, we define the rounding behavior expected.

- If `x_in` is positive, then we take `y_out` units of `y` out of the pool. `y_out` should be rounded down. Note that `y_f < y_0` here. Therefore to round `y_out = y_0 - y_f` down, given fixed `y_0`, we want to round `y_f` up.
- If `x_in` is negative, then `y_out` is also negative. The reason is that this is called in CalcInAmtGivenOut, so confusingly `x_in` is the known amount out, as a negative quantity. `y_out` is negative as well, to express that we get that many tokens out. (Since negative, `-y_out` is how many we add into the pool). We want `y_out` to be a larger negative, which means we want to round it down. Note that `y_f > y_0` here. Therefore `y_out = y_0 - y_f` is more negative, the higher `y_f` is. Thus we want to round `y_f` up.

And therefore we round up in both cases.

We capture all of this, in the following `solve_y` pseudocode:

```python
# solve_cfmm returns y_f s.t. CFMM_eq(x_f, y_f, w) = k
# for the no-v variant of CFMM_eq
def solve_y(x_0, y_0, w, x_in):
  x_f = x_0 + x_in
  k = CFMM_eq(x_0, y_0, w)
  err_tolerance = {"within factor of 10^-12", RoundUp}
  y_f = iterative_search(x_f, y_0, w, k, err_tolerance):
  y_out = y_0 - y_f
  return y_out
```

#### Using this in swap methods

So now we put together the components discussed in prior sections to achieve pseudocode for the SwapExactAmountIn
and SwapExactAmountOut functions.

We assume existence of a function `pool.ScaledLiquidity(input, output, rounding_mode)` that returns `in_reserve, out_reserve, rem_reserves`, where each are scaled by their respective scaling factor using the provided rounding mode.

##### SwapExactAmountIn

So now we need to put together the prior components.
When we scale liquidity, we round down, as lower reserves -> higher slippage.
Similarly when we scale the token in, we round down as well.
These both ensure no risk of over payment.

The amount of tokens that we treat as going into the "0-swap fee" pool we defined equations off of is: `amm_in = in_amt_scaled * (1 - swapfee)`. (With `swapfee * in_amt_scaled` just being added to pool liquidity)

Then we simply call `solve_y` with the input reserves, and `amm_in`.

<!-- TODO: Maybe we just use normal pseudocode syntax -->

```python
def CalcOutAmountGivenExactAmountIn(pool, in_coin, out_denom, swap_fee):
  in_reserve, out_reserve, rem_reserves = pool.ScaledLiquidity(in_coin, out_denom, RoundingMode.RoundDown)
  in_amt_scaled = pool.ScaleToken(in_coin, RoundingMode.RoundDown)
  amm_in = in_amt_scaled * (1 - swap_fee)
  out_amt_scaled = solve_y(in_reserve, out_reserve, remReserves, in_amt_scaled)
  out_amt = pool.DescaleToken(out_amt_scaled, out_denom)
  return out_amt
```

##### SwapExactAmountOut

<!-- TODO: Explain overall context of this section -->

When we scale liquidity, we round down, as lower reserves -> higher slippage.
Similarly when we scale the exact token out, we round up to increase required token in.

We model the `solve_y` call as we are doing a known change to the `out_reserve`, and solving for the implied unknown change to `in_reserve`.
To handle the swapfee, we apply the swapfee on the resultant needed input amount.
We do this by having `token_in = amm_in / (1 - swapfee)`.

<!-- TODO: Maybe we just use normal pseudocode syntax -->

```python
def CalcInAmountGivenExactAmountOut(pool, out_coin, in_denom, swap_fee):
  in_reserve, out_reserve, rem_reserves = pool.ScaledLiquidity(in_denom, out_coin, RoundingMode.RoundDown)
  out_amt_scaled = pool.ScaleToken(in_coin, RoundingMode.RoundUp)

  amm_in_scaled = solve_y(out_reserve, in_reserve, remReserves, -out_amt_scaled)
  swap_in_scaled = ceil(amm_in_scaled / (1 - swapfee))
  in_amt = pool.DescaleToken(swap_in_scaled, in_denom)
  return in_amt
```

We see correctness of the swap fee, by imagining what happens if we took this resultant input amount, and ran `SwapExactAmountIn (seai)`. Namely, that `seai_amm_in = amm_in * (1 - swapfee) = amm_in`, as desired!

#### Precision handling

{Something we have to be careful of is precision handling, notes on why and how we deal with it.}

<a name="err_proof">

#### Proof that |e_y| < |e_k|

</a>

In the binary search code, we are going to find a `k'` that is 'close' to `k`. We define and bound this error term as `e_k`, namely `e_k = k - k'`. We then find an implied value of `y'`, but this has an error term off of the true value of `y`, that would lead to exactly `k`. We call this term `y'`, and similarly `e_y = y - y'`.

Recall we compute `k` as `k = xy(x^2 + y^2 + w)`. So `k' = xy'(x^2 + (y')^2 + w)`. We seek to relate the error terms, so:

```tex
k - k' = xy(x^2 + y^2 + w) - xy'(x^2 + (y')^2 + w)
e_k = (y - y')(x^3 + xw) + x(y^3 - (y')^3)
e_k = e_y(x^3 + xw) + x(y^3 - (y')^3)
```

Notice that $(y')^3 = (y - e_y)^3 = y^3 - 3y^2e_y + 3ye_y^2 - e_y^3$.
We assume that `e_y` is sufficiently small, that $e_y^2$ is approximately `0`.
So we say $(y')^3 \approx y^3 - 3y^2e_y$.

Thus $e_k \approx e_y(x^3 + xw) + x(y^3 - y^3 + 3y^2e_y) = e_y(x^3 + xw) + 3xy^2e_y = e_y(x^3 + xw + 3xy^2)$.
Therefore
$$e_y = \frac{e_k}{x^3 + xw + 3xy^2}$$

Since `x` and `y` must be greater than 1, and `w` must be non-negative, we have that `x^3 + xw + 3xy^2 >= 1`.
Therefore $|e_y| < |e_k|$.
In fact, `e_y` is much lower than `e_k`.

### Spot Price

Spot price for an AMM pool is the derivative of its `CalculateOutAmountGivenIn` equation.
However for the stableswap equation, this is painful: [wolfram alpha link](https://www.wolframalpha.com/input?i=dy%2Fdx+of+y+%3D+%28sqrt%28729+k%5E2+x%5E4+%2B+108+x%5E3+%28w+x+%2B+x%5E3%29%5E3%29+%2B+27+k+x%5E2%29%5E%281%2F3%29%2F%283+2%5E%281%2F3%29+x%29+-+%282%5E%281%2F3%29+%28w+x+%2B+x%5E3%29%29%2F%28sqrt%28729+k%5E2+x%5E4+%2B+108+x%5E3+%28w+x+%2B+x%5E3%29%5E3%29+%2B+27+k+x%5E2%29%5E%281%2F3%29+)

So instead we compute the spot price by approximating the derivative via a small swap.

Let $\epsilon$ be a sentinel very small swap in amount.

Then $\text{spot price} = \frac{\text{CalculateOutAmountGivenIn}(\epsilon)}{\epsilon}$.

### LP equations

We divide this section into two parts, `JoinPoolNoSwap & ExitPool`, and `JoinPool`.

First we recap what are the properties that we'd expect from `JoinPoolNoSwap`, `ExitPool`, and LP shares.
From this, we then derive what we'd expect for `JoinPool`.

#### JoinPoolNoSwap and ExitPool

Both of these methods can be implemented via generic AMM techniques.
(Link to them or describe the idea)

#### JoinPool

The JoinPool API only supports JoinPoolNoSwap if

#### Join pool single asset in

Couple ways to define JoinPool Exit Pool relation

## Code structure

## Testing strategy

- Unit tests for every pool interface method
- Msg tests for custom messages
  - CreatePool
  - SetScalingFactors
- Simulator integrations:
  - Pool creation
  - JoinPool + ExitPool gives a token amount out that is lte input
  - SingleTokenIn + ExitPool + Swap to base token gives a token amount that is less than input
  - CFMM k adjusting in the correct direction after every action
- Fuzz test binary search algorithm, to see that it still works correctly across wide scale ranges
- Fuzz test approximate equality of iterative approximation swap algorithm and direct equation swap.
- Flow testing the entire stableswap scaling factor update process

## Extensions

- The astute observer may notice that the equation we are solving in $\text{solve cfmm}$ is actually a cubic polynomial in $y$, with an always-positive derivative. We should then be able to use newton's root finding algorithm to solve for the solution with quadratic convergence. We do not pursue this today, due to other engineering tradeoffs, and insufficient analysis being done.
