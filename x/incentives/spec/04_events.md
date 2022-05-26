# Events

The incentives module emits the following events:

## Handlers

### MsgCreateGauge

  Type            Attribute Key             Attribute Value
  ---------------; -------------------------; ---------------------;
  create\_gauge   gauge\_id                 {gaugeID}
  create\_gauge   distribute\_to            {owner}
  create\_gauge   rewards                   {rewards}
  create\_gauge   start\_time               {startTime}
  create\_gauge   num\_epochs\_paid\_over   {numEpochsPaidOver}
  message         action                    create\_gauge
  message         sender                    {owner}
  transfer        recipient                 {moduleAccount}
  transfer        sender                    {owner}
  transfer        amount                    {amount}

### MsgAddToGauge

  Type             Attribute Key   Attribute Value
  ----------------; ---------------; -----------------;
  add\_to\_gauge   gauge\_id       {gaugeID}
  create\_gauge    rewards         {rewards}
  message          action          create\_gauge
  message          sender          {owner}
  transfer         recipient       {moduleAccount}
  transfer         sender          {owner}
  transfer         amount          {amount}

## EndBlockers

### Incentives distribution

  Type           Attribute Key   Attribute Value
  --------------; ---------------; -----------------;
  transfer\[\]   recipient       {receiver}
  transfer\[\]   sender          {moduleAccount}
  transfer\[\]   amount          {distrAmount}
