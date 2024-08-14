
<!--
order: 7
-->

# Parameters

The treasury module contains the following parameters:

| Key                     | Type              | Example                |
|-------------------------|-------------------|------------------------|
| taxpolicy               | PolicyConstraints | {"rate_min": "0.0005", "rate_max": "0.01", "cap": {"denom": "usdr", "amount": "1000000"}, "change_rate_max": "0.00025"} |
| rewardpolicy            | PolicyConstraints | {"rate_min": "0.05", "rate_max": "0.5", "cap": {"denom": "unused", "amount": "0"}, "change_rate_max": "0.025"} |
| seigniorageburdentarget | string (dec)      | "0.670000000000000000" |
| miningincrement         | string (dec)      | "1.070000000000000000" |
| windowshort             | string (int)      | "4"                    |
| windowlong              | string (int)      | "52"                   |
| windowprobation         | string (int)      | "12"                   |
| burntaxsplit            | string (dec)      | "0.500000000000000000" |