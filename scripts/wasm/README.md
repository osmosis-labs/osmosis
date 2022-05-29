# Wasm Test Scripts

These are a few basic scripts to test that wasm integration is working.

They are a bit fragile, as we cannot easily parse out responses (like
proposal id, code id, contract address) using bash scripts. Since
ids/addresses are deterministic, these will run properly the first time
on any chain, and fail after that, unless you update the variables.

It may be easier to just edit the variables and possibly cut-and-paste
this. But these show a proper workflow and commands to test it.

## Assumptions

This is designed to run on dev-focused testnets / CI.

It assumes you have key called `validator` with enough voting power to
pass a proposal by itself. It assumes we are using
`--backend-keyring=test` for these keys. It also assumes the staking and
fee token is called `stake`.

This obviously will not work on any production-like network.

## Usage

First, download the contracts using `./get_contracts.sh`. This will
create some Wasm files in this directory.

Second, run a few proposals to add these codes to the blockchain, via
`./install_contracts.sh`.

After they are uploaded, wait for the proposals to pass and the code is
activated. Then, you can run `./test_cw20.sh` to create a cw20 token
owned by the validator, and sending some tokens to a new account.
