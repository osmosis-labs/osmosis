# Gov

The ```gov``` module enables on-chain governance which allows Osmosis token holders to participate in a community led decision-making process. For example, users can:

- Form an idea and seek feedback
- Create a proposal and adjust according to feedback as needed
- Submit a proposal along with an initial deposit
- Deposit tokens and fund an active proposal
- Vote for an active proposal

[comment]: <> (Add Proposal Process page)

</br>
</br>

## Overview

### Network parameters

The network parameters for the gov module are:

- **```deposit_params```** - Deposit related parameters
  - **```min_deposit```**: Minimum deposit (in uOSMO) for a proposal to enter voting period
  - **```max_deposit_period```**: Maximum period (in nanoseconds) for OSMO holders to deposit on a proposal.

- **```voting_params```** - Voting related parameters
  - **```voting_period```**: The length of the voting period (in nanoseconds)

- **```tally_params```** - Tally related parameters
  - **```quorum```**: The minimum percentage (in decimal form) of voting power that needs to be casted on a proposal for the result to be valid
  - **```threshold```**: Minimum proportion (in decimal form) of Yes votes (excluding Abstain votes) for the proposal to be accepted
  - **```veto```**: Minimum value of Veto votes to total votes ratio (in decimal form) for proposal to be vetoed.

</br>
</br>

### The Governance Procedure

**Phase 0 - Submit a proposal along with an initial deposit**

Users submits a proposal with an initial deposit. The proposal will then become "active" and enters the deposit period.

**Phase 1 - Deposit period**

During the deposit period, users can deposit and support an active proposal. Once the deposit of the proposal reaches the ```min_deposit```, it will enter the voting period. Otherwise, if the proposal is not successfully funded within ```max_deposit_period```, It will become inactive and **all the deposits will be burned**.

**Phase 2 - Voting period**

During the voting period, staked (bonded) tokens will be able to participate in the voting process. Users can choose one of the following options: ```yes```, ```no```, ```no_with_veto``` and ```abstain```.

After the ```voting_period``` has passed, the proposal will be considered "Rejected" and **the funds deposited in the deposit period will be burned if**:

- Votes do not reach the ```quorum```
- Enough vote ```no_with_veto``` when compared with total votes to meet the veto to total votes ratio specified in ```tally_params```

The proposal will be considered "Rejected" and **the funds deposited in the deposit period will be returned if**

- No one votes (or everyone votes to ```abstain```)
- More than ```threshold``` of non-abstaining voters vote ```no```

Otherwise, the proposal will be accepted and changes will be implemented according to the proposal.

</br>
</br>

## Transactions

### submit-proposal

Submit a proposal along with an initial deposit

```
tx gov submit-proposal [flags]
```

There are different types of proposal submission types, of them include `text`, `param-change`, `community-pool-spend`, `software-upgrade`, and `cancel-software-upgrade`. We will go over each of these submission types in detail now:

</br>
</br>

### submit-proposal (text)

Submit a proposal in text form

```bash
tx gov submit-proposal --title --description --type="Text" --from --chain-id
```

Text proposals differ from other proposal submission types in that after it passes, no logic is automatically executed. This is good for proposing changes to Osmosis that are not linked to a specific daemon parameter.

#### Example

Create a text signaling proposals to match external incentives for a `DOGE/OSMO` and `DOGE/ATOM` pair.

```bash
osmosisd tx gov submit-proposal --title="Match External Incentives for DOGE/OSMO and DOGE/ATOM pairs" --description="Input description" --type="Text" --from=WALLET_NAME --chain-id=CHAIN_ID
```

</br>
</br>

### submit-proposal (param change)

Submit a proposal to modify network parameters during run time

```
tx gov submit-proposal param-change [proposal-file] --from --chain-id
```

#### Example

Change the parameter MaxValidators (maximum number of validator) in the staking module:
 
```bash
osmosisd tx gov submit-proposal param-change proposal.json --from WALLET_NAME --chain-id CHAIN_ID
```
The proposal.json file would look as follows:

```json
{
  "title": "Staking Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 150
    }
  ]
}
```

</br>
</br>

### submit-proposal (community pool spend)

Submit a proposal and request funds from the community pool to support projects or other activities

```bash
tx gov submit-proposal community-pool-spend [proposal-file] --from --chain-id
```

#### Example

Submit a proposal to use community funds to fund a DAO:

```
osmosisd tx gov submit-proposal community-pool-spend proposal.json --from WALLET_NAME --chain-id CHAIN_ID
```

The proposal.json would look as follows:

```json
{
  "title": "Osmosis DAO",
  "description": "Establish a DAO for Osmosis. Potentially add external links for more information or allow discussion",
  "recipient": "osmo1r9pjvsuahxwkxg8cnhacd6alkmxq330fl9pqqt",
  "amount": [
    {
      "denom": "uosmo",
      "amount": "60000000000"
    }
  ]
}
```
If passed, the requested community funds would be sent to the recipient address provided in the json file.

</br>
</br>

### submit-proposal (software upgrade)

Submit an upgrade proposal and suggest a software upgrade at a specific block height

```
tx gov submit-proposal software-upgrade [proposal-file] --from --chain-id
```

#### Example

Update osmosis to V4:

```bash
osmosisd tx gov submit-proposal software-upgrade proposal.json --from WALLET_NAME --chain-id CHAIN_ID
```

The proposal.json would look as follows:

```json
{
  "name": "v4",
  "time": "0001-01-01T00:00:00Z",
  "height": "1314500",
  "info": "https://raw.githubusercontent.com/osmosis-labs/networks/main/osmosis-1/upgrades/v4/mainnet/upgrade_4_binaries.json",
},
```

</br>
</br>

### submit-proposal (cancel upgrade)

Cancel the planned software upgrade before the upgrade height is reached

```
tx gov submit-proposal cancel-software-upgrade --title= --description
```

The software upgrade does not have to be specified, as this will cancel the currently active software upgrade proposal. 

#### Example

If the above software upgrade proposal in the previous example was active, to propose its cancellation, run the following:

```bash
osmosisd tx gov submit-proposal cancel-software-upgrade --title="cancel v4" --description="cancel v4 upgrade" --from=WALLET_NAME --chain-id=CHAIN_ID
```

</br>
</br>

### submit-proposal (update pool incentives)

Update the weight of specified pool gauges in regards to their share of incentives

```
tx gov submit-proposal update-pool-incentives [proposal-file] --from --chain-id
```

#### Example

Update the pool incentives for `gauge_id` 0 and 1:

```bash
osmosisd tx gov submit-proposal update-pool-incentives proposal.json --from WALLET_NAME --chain-id CHAIN_ID
```

The proposal.json would look as follows:

```json
{
  "title": "Pool Incentive Adjustment",
  "description": "Adjust pool incentives",
  "records": [
    {
      "gauge_id": "0",
      "weight": "100000"
    },
    {
      "gauge_id": "1",
      "weight": "1766249"
    },
  ]
}
```

</br>
</br>

### deposit

Deposit tokens for an active proposal

```
tx gov deposit [proposal-id] [deposit] --from --chain-id
``` 

#### Example

If proposal number 12 is in the deposit period and you would like to help bring it to a vote, you could deposit 500 OSMO to that proposal as follows:

```bash
osmosisd tx gov deposit 12 500000000uosmo --from WALLET_NAME --chain-id CHAIN_ID
```

</br>
</br>

### vote

Vote for an active proposal

```
tx gov vote [proposal-id] [option] --from --chain-id
```

Valid value of ```option``` field is ```yes```, ```no```, ```no_with_veto``` and ```abstain```. 

#### Example

To vote yes for proposal 12:

```bash
osmosisd tx gov vote 12 yes --from WALLET_NAME --chain-id CHAIN_ID
```

</br>
</br>

## Queries

### proposals

Query all proposals

```
query gov proposals [proposal-id]
``` 

#### Example

We can list all proposals in json format by:

```bash
osmosisd query gov proposals -o json | jq
```

An example of the output:

```json
  {
    "proposals": [
      {
        "proposal_id": "1",
        "content": {
          "@type": "/cosmos.params.v1beta1.ParameterChangeProposal",
          "title": "Staking Param Change",
          "description": "Update max validators",
          "changes": [
            {
              "subspace": "staking",
              "key": "MaxValidators",
              "value": "150"
            }
          ]
        },
        "status": "PROPOSAL_STATUS_PASSED",
        "final_tally_result": {
          "yes": "50040000000000",
          "abstain": "0",
          "no": "0",
          "no_with_veto": "0"
        },
        "submit_time": "2021-10-15T10:05:49.996956080Z",
        "deposit_end_time": "2021-10-15T22:05:49.996956080Z",
        "total_deposit": [
          {
            "denom": "uosmo",
            "amount": "100000000"
          }
        ],
        "voting_start_time": "2021-10-15T10:14:56.958963929Z",
        "voting_end_time": "2021-10-15T22:14:56.958963929Z"
      }
    ],
    "pagination": {
      "next_key": null,
      "total": "0"
    }
  }
...
```

In the above example, there is only one proposal with ```"proposal_id": "1"```, with the title: ```"Staking Param Change"``` that change the ```MaxValidators``` parameter of the ```staking``` module to ```150```. We can also see that the status of the proposal is ```"PROPOSAL_STATUS_PASSED"```, which means that this proposal has been passed. In reality, the output would be much longer with all proposals listed.

</br>
</br>

### proposal

Query details of a single proposal

```
query gov proposal [proposal-id]
```

#### Example

To check proposal 13 and list in json format:

```bash
osmosisd query gov proposal 13 -o json | jq
```

</br>
</br>


### tally

Get the tally of a proposal vote

```
query gov tally [proposal-id]
```

#### Example

To check the tally of proposal 13 and output in json:

```bash
osmosisd query gov tally 13 -o json | jq
```

Which outputs:

```json
{
  "yes": "11126523145952",
  "abstain": "58623193556",
  "no": "44915148922",
  "no_with_veto": "5194297427"
}
```

This shows how the community voted on a specific proposal.

</br>
</br>

### params

Query the current gov parameters

```
query gov params
```

#### Example

To check the current gov parameters and output in json:

```bash
osmosisd query gov params --output json | jq
```

Which outputs:

```json
{
  "voting_params": {
    "voting_period": "259200000000000"
  },
  "tally_params": {
    "quorum": "0.200000000000000000",
    "threshold": "0.500000000000000000",
    "veto_threshold": "0.334000000000000000"
  },
  "deposit_params": {
    "min_deposit": [
      {
        "denom": "uosmo",
        "amount": "500000000"
      }
    ],
    "max_deposit_period": "1209600000000000"
  }
}
```

See the network parameters section for a detailed explanation of the above parameters.

</br>
</br>

## Appendix

### Current Configuration

```gov``` **module: Network Parameter effects and current configuration**

The following tables show overall effects on different configurations of the ```gov``` related network parameters:

<table><thead><tr><th></th> 
<th><code>min_deposit</code></th> 
<th><code>max_deposit_period</code></th> 
<th><code>voting_period</code></th></tr></thead> <tbody>
<tr><td>Type</td> 
<td>array (coins)</td> 
<td>string (time ns)</td> 
<td>string (time ns)</td></tr> 
<tr><td>Higher</td> 
<td>Larger window for calculating the downtime</td> 
<td>More time to solicit funds to reach <code>min_deposit</code> </td> 
<td>Longer voting period</td></tr> 
<tr><td>Lower</td> 
<td>Smaller window for calculating the downtime</td> 
<td>Less time to solicit funds to reach <code>min_deposit</code></td> 
<td>Shorter voting period</td></tr> 
<tr><td>Constraints</td> 
<td>Value has to be a positive integer</td> 
<td>Value has to be positive</td> 
<td>Value has to be positive</td></tr> 
<tr><td>Current configuration</td> 
<td><code>500000000</code> (500 OSMO)</td> <td><code>1209600000000000</code> (2 weeks)</td> <td><code>259200000000000</code> (3 days)</td></tr>
</tbody></table>

<table><thead><tr><th></th> 
<th><code>quorum</code></th> 
<th><code>threshold</code></th> 
<th><code>veto</code></th></tr></thead> 
<tbody><tr><td>Type</td> 
<td>string (dec)</td> 
<td>string (dec)</td> 
<td>string (dec)</td></tr> 
<tr><td>Higher</td> 
<td>Easier for a proposal to be passed</td> 
<td>Easier for a proposal to be passed</td> 
<td>Easier for a proposal to be passed</td></tr> 
<tr><td>Lower</td> 
<td>Harder for a proposal to be passed</td> 
<td>Harder for a proposal to be passed</td> 
<td>Harder for a proposal to be passed</td></tr> 
<tr><td>Constraints</td> 
<td>Value has to be less or equal to <code>1</code></td> 
<td>Value has to be less or equal to <code>1</code></td> 
<td>Value has to be less or equal to <code>1</code></td></tr> 
<tr><td>Current configuration</td> 
<td><code>0.2</code> (20%)</td> 
<td><code>0.5</code> (50%)</td> 
<td><code>0.334</code> (33.4%)</td></tr>
</tbody></table>