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

{include details of scaling factors and how they imply price peg, scaling factor governors}

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
We wish to have a simpler CFMM function to work with in these cases.
Due to the CFMM equation $f$ being a symmetric function, we can wlog reorder the arguments to the function. Thus we put the assets of relevance at the beginning of the function. So if two assets $x, y$, we write: $f(x,y, a_3, ... a_n) = xy * a_3 * ... a_n (x^2 + y^2 + a_3^2 + ... + a_n^2)$.

We then take a more convenient expression to work with, via variable substition.
$$ \begin{equation}
    v =
    \begin{cases}
      1, & \text{if}\ n=2 \\
      \prod_{i=3}^n a_i, & \text{otherwise}
    \end{cases}
  \end{equation} \newline 
  
  \begin{equation}
    w =
    \begin{cases}
      0, & \text{if}\ n=2 \\
      \sum_{i=3}^n a_i^2, & \text{otherwise}
    \end{cases}
  \end{equation} \newline
  \text{then } g(x,y,v,w) = xyv(x^2 + y^2 + w) = f(x,y, a_3, ... a_n) 
$$

As a corollary, notice that $g(x,y,v,w) = v * g(x,y,1,w)$, which will be useful when we have to compare before and after quantities. We will use $h(x,y,w) := g(x,y,1,w)$ as short-hand for this.

### Swaps

First notice that for all swaps, we only deal with two assets at a time, as swaps are given one asset in, and one asset out. For exposition, lets call the input asset $x$, the output asset $y$, and we can compute $v$ and $w$ given the other assets, whose reserves are untouched.

First we note the direct way of solving this, its limitation, and then the binary search equations.

#### Direct swap solution

The question we need to answer for a swap is "suppose I want to swap $a$ units of $x$, and then want to find how many units $b$ of $y$ that we get out".

The method to compute this under 0 swap fee is implied by the CFMM equation itself, since the constant refers to:
$g(x_0, y_0, v, w) = k = g(x_0 + a, y_0 - b, v, w)$. As $k$ is linearly related to $v$, and $v$ is unchanged throughout the swap, we can simplify the equation to be reasoning about $k' = \frac{k}{v}$ as the constant, and $h$ instead of $g$

We then model the solution by finding a function $\text{solve\_cfmm}(x, w, k') = y\text{ s.t. }h(x, y, w) = k'$.
Then we can solve the swap amount out by first computing $k'$ as $k' = h(x_0, y_0, w)$, and 
computing $y_f := \text{solve\_cfmm}(x_0 + a, w, k')$. We then get that $b = y_0 - y_f$.

So all we need is an equation for $\text{solve\_cfmm}$! Its essentially inverting a multi-variate polynomial, and in this case is solvable: [wolfram alpha link](https://www.wolframalpha.com/input?i=solve+for+y+in+x+*+y+*+%28x%5E2+%2B+y%5E2+%2B+w%29+%3D+k)

Or if were clever with simplification in the two asset case, we can reduce it to: [desmos link](https://www.desmos.com/calculator/ktdvu7tdxv).

These functions are a bit complex, which is fine as they are easy to prove correct. However, they are relatively expensive to compute, the latter needs precision on the order of x^4, and requires computing multiple cubic roots.

Instead there is a more generic way to compute these, which we detail in the next subsection.

#### Iterative search solution

Instead of using the direct solution for $\text{solve\_cfmm}(x, w, k')$, instead notice that $h(x, y, w)$ is an increasing function in $y$. 
So we can simply binary search for $y$ such that $h(x, y, w) = k'$, and we are guaranteed convergence within some error bound. 

In order to do a binary search, we need bounds on $y$. The lowest lowerbound is $0$, and the largest upperbound is $\infty$. The maximal upperbound is obviously unworkable, and in general binary searching around wide ranges is unfortunate, as we expect most trades to be centered around $y_0$. This would suggest that we should do something smarter to iteratively approach the right value. Notice that $h$ is super-linearly related in $y$, and at most cubically related to $y$. So $2 * h(x,y,w) <= h(x,2*y,w) <= 8 * h(x,y,w)$. We can use this fact to get a pretty-good initial upperbound and lowerbound guess for $y$. As these bounds are relatively tight (only 3 binary search steps away from one another), we do not aim to optimize further.

We detail how we use this below:
<!--Let $k_{guess,0} := h(x_f, y_0, w)$. If $k_{guess,0} > k$, then $y_0$ is our upper bound --->s

```python
def iterative_search(x_f, y_0, w, k, err_tolerance):
  k_0 = h(x_f, y_0, w)
  lowerbound, upperbound = y_0, y_0
  if k_0 > k:
    # need to find a lowerbound, assume cubic relationship
  elif k_0 < k:
    # need to find an upperbound, assume linear relationship

def binary_search(x_f, y_0, w, k, err_tolerance):
  iter_count = 0
  y_est = y_0
  k_est = h(x_f, y_0, w)
  lowerbound = 0
  upperbound = 
```

### Spot Price

### LP equations

#### Single-asset

#### Non-perfect multi-ratio

### Scaling factor handling

Throughout

## Code structure

## Testing strategy

## Extensions