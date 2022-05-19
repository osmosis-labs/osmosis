# Events

The lockup module emits the following events:

## Handlers

### MsgLockTokens

  Type           Attribute Key      Attribute Value
  -------------- ------------------ -----------------
  lock\_tokens   period\_lock\_id   {periodLockID}
  lock\_tokens   owner              {owner}
  lock\_tokens   amount             {amount}
  lock\_tokens   duration           {duration}
  lock\_tokens   unlock\_time       {unlockTime}
  message        action             lock\_tokens
  message        sender             {owner}
  transfer       recipient          {moduleAccount}
  transfer       sender             {owner}
  transfer       amount             {amount}

### MsgBeginUnlocking

  Type            Attribute Key      Attribute Value
  ---------------; ------------------; ------------------;
  begin\_unlock   period\_lock\_id   {periodLockID}
  begin\_unlock   owner              {owner}
  begin\_unlock   amount             {amount}
  begin\_unlock   duration           {duration}
  begin\_unlock   unlock\_time       {unlockTime}
  message         action             begin\_unlocking
  message         sender             {owner}

### MsgBeginUnlockingAll

  Type                 Attribute Key      Attribute Value
  --------------------; ------------------; -----------------------;
  begin\_unlock\_all   owner              {owner}
  begin\_unlock\_all   unlocked\_coins    {unlockedCoins}
  begin\_unlock        period\_lock\_id   {periodLockID}
  begin\_unlock        owner              {owner}
  begin\_unlock        amount             {amount}
  begin\_unlock        duration           {duration}
  begin\_unlock        unlock\_time       {unlockTime}
  message              action             begin\_unlocking\_all
  message              sender             {owner}

## Endblocker

### Automatic withdraw when unlock time mature

  Type             Attribute Key      Attribute Value
  ----------------; ------------------; -----------------;
  message          action             unlock\_tokens
  message          sender             {owner}
  transfer\[\]     recipient          {owner}
  transfer\[\]     sender             {moduleAccount}
  transfer\[\]     amount             {unlockAmount}
  unlock\[\]       period\_lock\_id   {owner}
  unlock\[\]       owner              {lockID}
  unlock\[\]       duration           {lockDuration}
  unlock\[\]       unlock\_time       {unlockTime}
  unlock\_tokens   owner              {owner}
  unlock\_tokens   unlocked\_coins    {totalAmount}
