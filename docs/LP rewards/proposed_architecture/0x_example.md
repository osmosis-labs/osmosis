# Reward distribution for Gauges

## example 1: lock on day 0, unlock on day 15
| Gauge       | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10| 11| 12| 13| 14| 15| 16| 17| 18| 19| 20| 21| 22| 23| 24| 25| 26| 27| 28| 29|
|----------   |---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 ( 1 day ) | - | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | - |
| 2 ( 7 days) | - | x | x | x | x | x | x | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | x | x | x | x | x | x | - |
| 3 (14 days) | - | x | x | x | x | x | x | x | x | x | x | x | x | x | O | O | x | x | x | x | x | x | x | x | x | x | x | x | x | - |


## example 2: lock on day 0, unlock on day 10
| Gauge       | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10| 11| 12| 13| 14| 15| 16| 17| 18| 19| 20| 21| 22| 23| 24|
|----------   |---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 ( 1 day ) | - | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | O | - |
| 2 ( 7 days) | - | x | x | x | x | x | x | O | O | O | O | O | O | O | O | O | O | O | x | x | x | x | x | x | - |
| 3 (14 days) | - | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | x | - |
- Gauge 3 does not get any rewards as it would never be locked for 14 days; as soon as user transitions to unlock, counting days for locked gauges stops
