# Epochs

Often in the SDK, we would like to run certain code every-so often. The purpose of the `epochs` module is to allow other modules to set that they would like to be signaled once every period, and then execute code after this period has taken place. In other words, `epochs` creates a generalized epoch interface to other modules so that they can easily be signalled after set periods of time.

</br>
</br>

## Queries

### epoch-infos

Query the currently running epochInfos

```
osmosisd query epochs epoch-infos
```

#### Example

An example output:

```sh
epochs:
- current_epoch: "183"
  current_epoch_start_height: "2438409"
  current_epoch_start_time: "2021-12-18T17:16:09.898160996Z"
  duration: 86400s
  epoch_counting_started: true
  identifier: day
  start_time: "2021-06-18T17:00:00Z"
- current_epoch: "26"
  current_epoch_start_height: "2424854"
  current_epoch_start_time: "2021-12-17T17:02:07.229632445Z"
  duration: 604800s
  epoch_counting_started: true
  identifier: week
  start_time: "2021-06-18T17:00:00Z"
```

</br>
</br>

### current-epoch

Query the current epoch by the specified identifier

```
osmosisd query epochs current-epoch [identifier]
```

#### Example

Query the current `day` epoch:

```sh
osmosisd query epochs current-epoch day
```

Which in this example outputs:

```sh
current_epoch: "183"
```