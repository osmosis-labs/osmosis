<!--
order: 5
-->

# Events

The market module emits the following events:

## Handlers

### MsgSwap

| Type    | Attribute Key | Attribute Value    |
|---------|---------------|--------------------|
| swap    | offer         | {offerCoin}        |
| swap    | trader        | {traderAddress}    |
| swap    | recipient     | {recipientAddress} |
| swap    | swap_coin     | {swapCoin}         |
| swap    | swap_fee      | {swapFee}          |
| message | module        | market             |
| message | action        | swap               |
| message | sender        | {senderAddress}    |

### MsgSwapSend

| Type    | Attribute Key | Attribute Value    |
|---------|---------------|--------------------|
| swap    | offer         | {offerCoin}        |
| swap    | trader        | {traderAddress}    |
| swap    | recipient     | {recipientAddress} |
| swap    | swap_coin     | {swapCoin}         |
| swap    | swap_fee      | {swapFee}          |
| message | module        | market             |
| message | action        | swapsend           |
| message | sender        | {senderAddress}    |
