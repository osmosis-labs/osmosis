# Generalized Solidly Stableswap

Stableswaps are pools that offer low slippage for two assets that are intended to be tightly correlated.
There is a price ratio they are expected to be at, and the AMM offers low slippage around this price.
There is still price impact for each trade, and as the liquidity becomes more lop-sided, the slippage drastically increases.

This package implements the Solidly stableswap curve, namely a CFMM with
invariant: $f(x, y) = xy(x^2 + y^2) = k$

It is generalized to the multi-asset setting as $f(a_1, ..., a_n) = a_1 * ... * a_n (a_1^2 + ... + a_n^2)$

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

##### Altering binary search equations due to error tolerance

Great, we have a binary search to finding an input `new_y_reserve`, such that we get a value `k` within some error bound close to the true desired `k`! We can prove that an error by a factor of `e` in `k`, implies an error of a factor less than `e` in `new_y_reserve`. So we could set `e` to be close to some correctness bound we want. Except... `new_y_reserve >> y_in`, so we'd need an extremely high error tolerance for this to work. So we actually want to adapt the equations, to reduce the "common terms" in `k` that we need to binary search over, to help us search. To do this, we open up what are we doing again, and re-expose `y_out` as a variable we explicitly search over (and therefore get error terms in `k` implying error in `y_out`)

What we are doing above in the binary search is setting `k_target` and searching over `y_f` until we get `k_iter` {within tolerance} to `k_target`. Sine we want to change to iterating over $y_{out}$, we unroll that $y_f = y_0 - y_{out}$ where they are defined as:
$$k_{target} = x_0 y_0 (x_0^2 + y_0^2 + w)$$
$$k_{iter}(y_0 - y_{out}) = h(x_f, y_0 - y_{out}, w) = x_f (y_0 - y_{out}) (x_f^2 + (y_0 - y_{out})^2 + w)$$

But we can remove many of these terms! First notice that `x_f` is a constant factor in `k_iter`, so we can just divide `k_target` by `x_f` to remove that. Then we switch what we search over, from `y_f` to `y_out`, by fixing `y_0`, so were at:

$$k_{target} = x_0 y_0 (x_0^2 + y_0^2 + w) / x_f$$

$$k_{iter}(y_{out}) = (y_0 - y_{out}) (x_f^2 + (y_0 - y_{out})^2 + w) = (y_0 - y_{out}) (x_f^2 + w) + (y_0 - y_{out})^3$$

So $k_{iter}(y_{out})$ is a cubic polynomial in $y_{out}$. Next we remove the terms that have no dependence on `y_{delta}` (the constant term in the polynomial). To do this first we rewrite this to make the polynomial clearer:

$$k_{iter}(y_{out}) = (y_0 - y_{out}) (x_f^2 + w) + y_0^3 - 3y_0^2 y_{out} + 3 y_0 y_{out}^2 - y_{out}^3$$

$$k_{iter}(y_{out}) = y_0 (x_f^2 + w) - y_{out}(x_f^2 + w) + y_0^3 - 3y_0^2 y_{out} + 3 y_0 y_{out}^2 - y_{out}^3$$

$$k_{iter}(y_{out}) = -y_{out}^3 + 3 y_0 y_{out}^2 - (x_f^2 + w + 3y_0^2)y_{out} + (y_0 (x_f^2 + w) + y_0^3)$$

So we can subtract this constant term `y_0 (x_f^2 + w) + y_0^3`, which for `y_out < y_0` is the dominant term in the expression!

So lets define this as:

$$k_{target} = \frac{x_0 y_0 (x_0^2 + y_0^2 + w)}{x_f} - (y_0 (x_f^2 + w) + y_0^3)$$

$$k_{iter}(y_{out}) = -y_{out}^3 + 3 y_0 y_{out}^2 - (x_f^2 + w + 3y_0^2)y_{out}$$

We prove [here](#err_proof) that an error of a multiplicative `e` between `target_k` and `iter_k`, implies an error of less than a factor of `10e` in `y_{out}`, as long as `|y_{out}| < y_0`. (The proven bounds are actually better)

We target an error of less than `10^{-8}` in `y_{out}`, so we conservatively set a bound of `10^{-12}` for `e_k`.

##### Combined pseudocode

Now we want to wrap this binary search into `solve_cfmm`. We changed the API slightly, from what was previously denoted, to have this "y_0" term, in order to derive initial bounds.

One complexity is that in the iterative search, we iterate over $y_f$, but then translate to $y_0$ in the internal equations.
So we also use the 

```python
# solve_y returns y_out s.t. CFMM_eq(x_f, y_f, w) = k = CFMM_eq(x_0, y_0, w)
# for x_f = x_0 + x_in.
def solve_y(x_0, y_0, w, x_in):
  x_f = x_0 + x_in
  err_tolerance = {"within factor of 10^-12", RoundUp}
  y_f = iterative_search(x_0, x_f, y_0, w, err_tolerance)
  y_out = y_0 - y_f
  return y_out

def iter_k_fn(x_f, y_0, w):
  def f(y_f):
    y_out = y_0 - y_f
    return -(y_out)**3 + 3 y_0 * y_out^2 - (x_f**2 + w + 3*y_0**2) * y_out

def iterative_search(x_0, x_f, y_0, w, err_tolerance):
  target_k = target_k_fn(x_0, y_0, w, x_f)
  iter_k_calculator = iter_k_fn(x_f, y_0, w)

  # use original CFMM to get y_f reserve bounds
  bound_estimation_target_k = cfmm(x_0, y_0, w)
  bound_estimation_k0 = cfmm(x_f, y_0, w)
  lowerbound, upperbound = y_0, y_0
  k_ratio = bound_estimation_k0 / bound_estimation_target_k
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
  max_iteration_count = 100
  return binary_search(lowerbound, upperbound, k_calculator, target_k, err_tolerance)

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

##### Setting the error tolerance

What remains is setting the error tolerance. We need two properties:

- The returned value to be within some correctness threshold of the true value
- The returned value to be rounded correctly (always ending with the user having fewer funds to avoid pool drain attacks). Mitigated by swap fees for normal swaps, but needed for 0-fee to be safe.

The error tolerance we set is defined in terms of error in `k`, which itself implies some error in `y`.
An error of `e_k` in `k`, implies an error `e_y` in `y` that is less than `e_k`. We prove this [here](#err_proof) (and show that `e_y` is actually much less than the error in `e_k`, but for simplicity ignore this fact). We want `y` to be within a factor of `10^(-12)` of its true value.
To ensure the returned value is always rounded correctly, we define the rounding behavior expected.

- If `x_in` is positive, then we take `y_out` units of `y` out of the pool. `y_out` should be rounded down. Note that `y_f < y_0` here. Therefore to round `y_out = y_0 - y_f` down, given fixed `y_0`, we want to round `y_f` up.
- If `x_in` is negative, then `y_out` is also negative. The reason is that this is called in CalcInAmtGivenOut, so confusingly `x_in` is the known amount out, as a negative quantity. `y_out` is negative as well, to express that we get that many tokens out. (Since negative, `-y_out` is how many we add into the pool). We want `y_out` to be a larger negative, which means we want to round it down. Note that `y_f > y_0` here. Therefore `y_out = y_0 - y_f` is more negative, the higher `y_f` is. Thus we want to round `y_f` up.

And therefore we round up in both cases.

##### Further optimization

- The astute observer may notice that the equation we are solving in $\text{solve cfmm}$ is actually a cubic polynomial in $y$, with an always-positive derivative.
We should then be able to use newton's root finding algorithm to solve for the solution with quadratic convergence.
We do not pursue this today, due to other engineering tradeoffs, and insufficient analysis being done.

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
  out_amt_scaled = solve_y(in_reserve, out_reserve, remReserves, amm_in)
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
  out_amt_scaled = pool.ScaleToken(out_coin, RoundingMode.RoundUp)

  amm_in_scaled = solve_y(out_reserve, in_reserve, remReserves, -out_amt_scaled)
  swap_in_scaled = ceil(amm_in_scaled / (1 - swapfee))
  in_amt = pool.DescaleToken(swap_in_scaled, in_denom)
  return in_amt
```

We see correctness of the swap fee, by imagining what happens if we took this resultant input amount, and ran `SwapExactAmountIn (seai)`. Namely, that `seai_amm_in = amm_in * (1 - swapfee) = amm_in`, as desired!

#### Precision handling

{Something we have to be careful of is precision handling, notes on why and how we deal with it.}

<a name="err_proof">

#### Proof that |e_y| < 100|e_k|

</a>

The function $f(y_{out}) = -y_{out}^3 + 3 y_0 y_{out}^2 - (x_f^2 + w + 3y_0^2)y_{out}$ is monotonically increasing over the reals.
You can prove this, by seeing that its [derivative's](https://www.wolframalpha.com/input?i=d%2Fdx+-x%5E3+%2B+3a+x%5E2+-+%28b+%2B+3a%5E2%29+x+) 0 values are both imaginary, and therefore has no local minima or maxima in the reals.
Therefore, there exists exactly one real $y_{out}$ s.t. $f(y_{out}) = k$.
Via binary search, we solve for a value $y_{out}^{\*}$ such that $\left|\frac{ k - k^{\*} }{k}\right| < e_k$, where $k^{\*} = f(y_{out}^{\*})$. We seek to then derive bounds on $e_y = \left|\frac{ y_{out} - y_{out}^{\*} }{y_{out}}\right|$ in relation to $e_k$.

**Theorem**: $e_y < 100 e_k$ as long as $|y_{out}| <= .9y_0$.
**Informal**, we claim that for $.9y_0 < |y_{out}| < y_0$, `e_y` is "close" to `e_k` under expected parameterizations. And for $y_{out}$ significantly less than $.9y_0$, the error bounds are much better. (Often better than $e_k$)


Let $y_{out} - y_{out}^* = a_y$, we are going to assume that $a_y << y_{out}$, and will justify this later. But due to this, we treat $a_y^c = 0$ for $c > 1$. This then implies that $y_{out}^2 - y_{out}^{*2} = y_{out}^2 - (y_{out} - a_y)^2 \approx 2y_{out}a_y$, and similarly $y_{out}^3 - y_{out}^{*3} \approx 3y_{out}^2 a_y$

Now we are prepared to start bounding this.
$$k - k^{\*} = -(y_{out}^3 - y_{out}^{3\*}) + 3y_0(y_{out}^2 - y_{out}^{2\*}) - (x_f^2 + w + 3y_0^2)(y_{out} - y_{out}^{\*})$$

$$k - k^{\*} \approx -(3y_{out}^2 a_y) + 3y_0 (2y_{out}a_y) - (x_f^2 + w + 3y_0^2)a_y$$

$$k - k^{\*} \approx a_y(-3y_{out}^2 + 6y_0y_{out} - (x_f^2 + w + 3y_0^2))$$

Rewrite $k = y_{out}(-y_{out}^2 + 3y_0y_{out} - (x_f^2 + w + 3y_0^2))$

$$e_k > \left|\frac{ k - k^{\*} }{k}\right| = \left|\frac{a_y}{y_{out}} \frac{(-3y_{out}^2 + 6y_0y_{out} - (x_f^2 + w + 3y_0^2))}{(-y_{out}^2 + 3y_0y_{out} - (x_f^2 + w + 3y_0^2))}\right|$$

Notice that $\left|\frac{a_y}{y_{out}}\right| = e_y$! Therefore

$$e_k > e_y\left|\frac{(-3y_{out}^2 + 6y_0y_{out} - (x_f^2 + w + 3y_0^2))}{(-y_{out}^2 + 3y_0y_{out} - (x_f^2 + w + 3y_0^2))}\right|$$

We bound the right hand side, with the assistance of wolfram alpha. Let $a = y_{out}, b = y_0, c = x_f^2 + w$. Then we see from [wolfram alpha here](https://www.wolframalpha.com/input?i=%7C%28-3a%5E2+%2B+6ab+-+%28c+%2B+3b%5E2%29%29+%2F+%28-a%5E2+%2B+3ab+-+%28c+%2B+3b%5E2%29%29+%7C+%3E+.01), that this right hand expression is provably greater than `.01` if some set of decisions hold. We describe the solution set that satisfies our use case here:

* When $y_{out} > 0$
  * Use solution set: $a > 0, b > \frac{2}{3} a, c > \frac{1}{99} (-299a^2 + 597ab - 297b^2)$
    * $a > 0$ by definition.
    * $b > \frac{2}{3} a$, as thats equivalent to $y_0 > \frac{2}{3} y_{out}$. We already assume that $y_0 >= y_{out}$.
    * Set $y_{out} = .9y_0$, per our theorem assumption. So $b = .9a$. Take $c = x^2 + w = 0$. Then [we can show that](https://www.wolframalpha.com/input?i=0+%3E+-299a%5E2+%2B+597ab+-+297b%5E2%2C+when+b%3D+.90a) $(-299a^2 + 597ab - 297b^2) < 0$ for all $a$. This completes the constraint set.
* When $y_{out} < 0$
  * Use solution set: $a < 0, b > \frac{2}{3} a, c > -a^2 + 3ab - 3b^2$
    * $a < 0$ by definition.
    * $b > \frac{2}{3} a$, as $y_0$ is positive.
    * $c > 0$ is by definition, so we just need to bound when $-a^2 + 3ab - 3b^2 < 0$. This is always the case as long as one of $a$ or $b$ is non-zero, per [here](https://www.wolframalpha.com/input?i=-a%5E2+%2B+3ab+-+3b%5E2+%3C+0).

Tieing this all together, we have that $e_k > .01e_y$. Therefore $e_y < 100 e_k$, satisfying our theoerem!

To show the informal claims, the constraint that led to this 100x error blowup was trying to accomodate high $y_{out}$. When $y_{out}$ is smaller, the error is far lower. (Often to the case that $e_y < e_k$, you can convince yourself of this by setting the ratio to being greater than 1 in wolfram alpha) When $y_{out}$ is bigger than $.9y_0$, we can rely on x_f^2 + w being much larger to lower this error. In these cases, the $x_f$ term must be large relative to $y_0$, which would yield a far better error bound.

TODO: Justify a_y << y_out. (This should be easy, assume its not, that leads to e_k being high. Ratio test probably easiest. Maybe just add a sentence to that effect)

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

There are a couple ways to define `JoinPoolSingleAssetIn`. The simplest way is to define it from its intended relation from the CFMM, with Exit pool. We describe this below under the zero swap fee case.

Let `pool_{L, S}` represent a pool with liquidity `L`, and `S` total LP shares.
If we call `pool_{L, S}.JoinPoolSingleAssetIn(tokensIn) -> (N, pool_{L + tokensIn, S + N})`, or in others we get out `N` new LP shares, and a pool with with tokensIn added to liquidity. 
It must then be the case that `pool_{L+tokensIn, S+N}.ExitPool(N) -> (tokensExited, pool_{L + tokensIn - tokensExited, S})`.
Then if we swap all of `tokensExited` back to tokensIn, under 0 swap fee, we should get back to `pool_{L, S}` under the CFMM property.

In other words, if we single asset join pool, and then exit pool, we should return back to the same CFMM `k` value we started with. Then if we swap back to go entirely back into our input asset, we should have exactly many tokens as we started with, under 0 swap fee.

We can solve this relation with a binary search over the amount of LP shares to give!

Thus we are left with how to account swap fee. We currently account for swap fee, by considering the asset ratio in the pool. If post scaling factors, the pool liquidity is say 60:20:20, where 60 is the asset were bringing in, then we consider "only (1 - 60%) = 40%" of the input as getting swapped. So we charge the swap fee on 40% of our single asset join in input. So the pseudocode for this is roughly:

```python
def JoinPoolSingleAssetIn(pool, tokenIn):
  swapFeeApplicableFraction = 1 - (pool.ScaledLiquidityOf(tokenIn.Denom) / pool.SumOfAllScaledLiquidity())
  effectiveSwapFee = pool.SwapFee * swapFeeApplicableFraction
  effectiveTokenIn = RoundDown(tokenIn * (1 - effectiveSwapFee))
  return BinarySearchSingleJoinLpShares(pool, effectiveTokenIn)
```

We leave the rounding mode for the scaling factor division unspecified.
This is because its expected to be tiny (as the denominator is larger than the numerator, and we are operating in BigDec),
and it should be dominated by the later step of rounding down.

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
