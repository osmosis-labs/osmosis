<!--
order: 4
-->

# Events

The incentives module emits the following events:

## Handlers

### MsgCreatePot

| Type                | Attribute Key       | Attribute Value |
| ------------------- | ------------------- | --------------- |
| create_pot          | pot_id              | {potID}         |
| create_pot          | distribute_to       | {owner}         |
| create_pot          | rewards             | {rewards}       |
| create_pot          | start_time          | {startTime}     |
| create_pot          | num_epochs          | {numEpochs}     |
| message             | action              | create_pot      |
| message             | sender              | {owner}         |
| transfer            | recipient           | {moduleAccount} |
| transfer            | sender              | {owner}         |
| transfer            | amount              | {amount}        |

### MsgAddToPot

| Type                | Attribute Key       | Attribute Value |
| ------------------- | ------------------- | --------------- |
| add_to_pot          | pot_id              | {potID}         |
| create_pot          | rewards             | {rewards}       |
| message             | action              | create_pot      |
| message             | sender              | {owner}         |
| transfer            | recipient           | {moduleAccount} |
| transfer            | sender              | {owner}         |
| transfer            | amount              | {amount}        |

## EndBlockers

### Incentives distribution

| Type          | Attribute Key  | Attribute Value    |
| ------------- | -------------- | ------------------ |
| transfer[]    | recipient      | {receiver}         |
| transfer[]    | sender         | {moduleAccount}    |
| transfer[]    | amount         | {distrAmount}      |
