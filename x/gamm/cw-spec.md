# CWX Spec: AMM Customization Contracts

CWX is a specification for AMM custom data provider contracts. Instead of a
module handling swap execution, contracts(which could be provided by either
users or other modules), with implementing CWX specification, serve as a dynamic
curve provider.

The customization is split into the following parts:
- Paremeterization: Contracts provide numeric paremeters without actual
execution.
- Customization: Contracts are turing complete script that executes swap
operations.
- Governance: Contracts updates internal parameter and/or code.

In most cases, simple parameterization with governance module would be enough 
for constructing and managing swap pools. However, users could utilize
CosmWasm's high degree of freedom, including cross-contract call and turing
completeness to provide complex AMM functionalities. For example,

- Parameter calculation would require external oracle feed.
- Non-traditional curve expression outside of the limitation of AMM module.
- Delegated governance using a CW4 group contract.

CWX swap execution works as fallback style. Each custom swap operation 
requires a set of parameter set. For a specific operation message, if there is a
corresponding custom message handler implemented, it is executed in priority,
and if there is not, the set of parameter methods are queried and the module
executes the swap using those parameters. Swap operations are always called 
through an AMM module.

Governance messages could be called externally(invoked by users).

## Messages

Creating a pool is done by deploying a new CWX compatible contract.

`Swap{TokenIn, TokenInMaxAmount, TokenOut, TokenOutMaxAmount, MaxSpotPrice}`
makes a swap operation. It returns `SwapPair{TokenInAmount, TokenOutAmount}` and
the token transfer is handled by the caller. This operation can have additional
side effects, including pool size adjustment.

Either `TokenIn` xor `TokenOut` should be provided and the amount of the other
size is calculated relatively, where it does not violates the conditional
parameters.

`JoinPool{}` 
