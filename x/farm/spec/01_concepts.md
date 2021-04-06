<!--
order: 1
-->

# Concept
The farm module is a generalized, simplified, and minimal F1 distribution.  
https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/fee_distribution/f1_fee_distr.pdf

Iterating over every allocation record to calculate how much reward each farmer should receive requires excessive computation. Therefore, the farm module uses the below function to calculate the rewards.  

As described in to the document above, the amount of reward that an account that deposited the share can be calculated using the formula below.  
![latex](https://latex.codecogs.com/svg.latex?x%20\sum_{i%20=%20k%20+%201}^{f}%20\frac{T_i}{n_i}%20=%20x\left(\left(\sum_{i=0}^{f}\frac{T_i}{n_i}\right)%20-%20\left(\sum_{i=0}^{k}\frac{T_i}{n_i}\right)\right)%20=%20x\left(Entry_f%20-%20Entry_k\right))

Each farm's `historical record` describes the below up to a certain period.  
![latex](https://latex.codecogs.com/svg.latex?\sum_{i=0}^{f}\frac{T_i}{n_i})  

The Cosmos-SDK distribution module uses the block height as the period. However, the farm module increases the period by 1 for every time reward is allocated to allow it to be used more conveniently.  

Every time the period increases,  
![latex](https://latex.codecogs.com/svg.latex?Entry_f%20=%20\sum_{i=0}^{f}\frac{T_i}{n_i}%20=%20\sum_{i=0}^{f-1}\frac{T_i}{n_i}%20+%20\frac{T_f}{n_f}%20=%20Entry_{f-1}%20+%20\frac{T_f}{n_f})  
a new historical record is calculated then stored using the formula above.  

The reward amount to be received by the farmer can be calculated by finding the difference between the the historical record of when the reward was last withdrawn and the most recent historical record, then multiplying the share that the farmer owns.  

The farm module only provides custody for reward's allocation and distribution. Because shares may not actually be coins, the farm module doesn't custody the shares internally. If coin is a share, the module that's using the farm module must custody the coin.
