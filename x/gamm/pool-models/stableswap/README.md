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

### Swaps

First notice that for all swaps, we only deal with two assets at a time, as swaps are given one asset in, and one asset out. For exposition, lets call the input asset $x$, the output asset $y$, and we can compute $v$ and $w$ given the other assets, whose reserves are untouched.

First we note the direct way of solving this, its limitation, and then the binary search equations.

#### Direct swap solution

Suppose the existence of a function $\text{solve\_cfmm}(x, v, w, k) = y\text{ s.t. }g(x, y, v, w) = k$.
Then we can solve swaps by first computing $k = g(x_0, y_0, v, w)$.
Then suppose I want to swap $a$ units of $x$, and then want to find how many units $b$ of $y$ that we get out.
We do this by computing $y_f = \text{solve\_cfmm}(x_0 + a, v, w, k)$, and then $b = y_0 - y_f$

So all we need is an equation for $\text{solve\_cfmm}$! Its essentially inverting a multi-variate polynomial, and in this case is solvable: <https://www.wolframalpha.com/input?i=solve+for+y+in+x+*+y+*+v+*+%28x%5E2+%2B+y%5E2+%2B+w%29+%3D+k>

Or if were clever with simplification in the two asset case, we can reduce it to: <https://www.desmos.com/calculator/ktdvu7tdxv>.

These functions are a bit complex, which is fine as they are easy to prove correct. However, they are relatively expensive to compute, the latter needs precision on the order of x^4, and requires computing multiple cubic roots.

Instead there is a more generic way to compute these, which we detail in the next subsection.

#### Binary search solution

### Spot Price

### LP equations

#### Single-asset

#### Non-perfect multi-ratio

### Scaling factor handling

Throughout

## Code structure

## Testing strategy

## Extensions