# Reward distribution for Gauges

## example 1: 14-day lockup, lock before end of epoch 1, unlock after 15 epochs elapses
| Gauge       | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10| 11| 12| 13| 14| 15| 16| 17| 18| 19| 20| 21| 22| 23| 24| 25| 26| 27| 28| 29|
|----------   |---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 ( 1 day ) | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | - |
| 2 ( 7 days) | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | x | x | x | x | x | x | - |
| 3 (14 days) | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | x | x | x | x | x | x | x | x | x | x | x | x | x | - |
- table entries for end of an epoch
  - 1 = end of epoch 1
  - 2 = end of epoch 2


## example 2: 14-day lockup, lock before end of epoch 1, unlock after 10 epochs elapses
| Gauge       | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10| 11| 12| 13| 14| 15| 16| 17| 18| 19| 20| 21| 22| 23| 24|
|----------   |---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 ( 1 day ) | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | - |
| 2 ( 7 days) | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | x | x | x | x | x | x | - |
| 3 (14 days) | O | O | O | O | O | O | O | O | O | O | x | x | x | x | x | x | x | x | x | x | x | x | x | - |
