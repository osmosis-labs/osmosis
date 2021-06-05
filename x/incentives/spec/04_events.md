<!--
order: 4
-->

# Events

The incentives module emits the following events:

## Handlers

### MsgCreateGauge

| Type         | Attribute Key        | Attribute Value     |
| ------------ | -------------------- | ------------------- |
| create_gauge | gauge_id             | {gaugeID}           |
| create_gauge | distribute_to        | {owner}             |
| create_gauge | rewards              | {rewards}           |
| create_gauge | start_time           | {startTime}         |
| create_gauge | num_epochs_paid_over | {numEpochsPaidOver} |
| message      | action               | create_gauge        |
| message      | sender               | {owner}             |
| transfer     | recipient            | {moduleAccount}     |
| transfer     | sender               | {owner}             |
| transfer     | amount               | {amount}            |

### MsgAddToGauge

| Type         | Attribute Key | Attribute Value |
| ------------ | ------------- | --------------- |
| add_to_gauge | gauge_id      | {gaugeID}       |
| create_gauge | rewards       | {rewards}       |
| message      | action        | create_gauge    |
| message      | sender        | {owner}         |
| transfer     | recipient     | {moduleAccount} |
| transfer     | sender        | {owner}         |
| transfer     | amount        | {amount}        |

## EndBlockers

### Incentives distribution

| Type       | Attribute Key | Attribute Value |
| ---------- | ------------- | --------------- |
| transfer[] | recipient     | {receiver}      |
| transfer[] | sender        | {moduleAccount} |
| transfer[] | amount        | {distrAmount}   |
