# Pool-incentives

The `pool-incentives` module is separate but related to the `incentives` module. When a pool is created using the `GAMM` module, the `pool-incentives` module creates individual gauges for every lock duration that exists in that pool.


distributes incentives to the bonded LP tokens within the `GAMM` module.





</br>
</br>





## Transactions

### replace-pool-incentives 

Submit a full replacement to the records for pool incentives




</br>
</br>




### update-pool-incentives  

Submit an update to the records for pool incentives





</br>
</br>

## Queries

### distr-info                   

Query distribution info



</br>
</br>


### external-incentivized-gauges 

Query external incentivized gauges



</br>
</br>



### gauge-ids                    

Query the matching gauge ids and durations by pool id




</br>
</br>





### incentivized-pools           

Query incentivized pools




</br>
</br>




### lockable-durations           

Query lockable durations





</br>
</br>



### params                       

Query module params