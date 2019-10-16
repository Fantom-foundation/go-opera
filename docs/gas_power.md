
## Gas power
```GasPower``` limits the ```GasLimit``` of transactions which validator may originate.

This mechanics is based on ```MedianTime```. The general idea is simple -
the available gas power is proportional to event’s ```MedianTime```, i.e.
more time passed since self-parent - the more gas power validator has in this event.
Due to ```MedianTime```, it isn't possible to bias the ```MedianTime```,
unless more than 1/2W collude about the shift of time.

#### Exact formula
Constants for each validator:
- ```TotalPerH``` = determines the gas power allocation per hour in the whole network.
- ```validator’s gas per hour``` = ```TotalPerH``` * ```validator’s stake``` / ```total stake```.
- ```max gas power``` = ```validator’s gas per hour``` * ```MaxStashedPeriod```.
- ```startup gas``` = max(```validator’s gas per hour``` * ```StartupPeriod```, ```MinStartupGasPower```).

```go
if {e.SelfParent} != nil
    {prev gas left} = e.SelfParent.GasPowerLeft
    {prev median time} = e.SelfParent.MedianTime
else if prevEpoch.LastConfirmedEvent[validator] exists
    {prev gas left} = max(prevEpoch.LastConfirmedEvent[validator].GasPowerLeft, {startup gas})
    {prev median time} = prevEpoch.LastConfirmedEvent[validator].MedianTime
else
    {gas stashed} = {startup gas}
    {prev median time} = prevEpoch.Time
{gas power allocated} = ({e.MedianTime} - {prev median time}) * {validator’s gas per hour} / hour
{GasPower} = {prev gas left} + {gas power allocated}
if {GasPower} > {max gas power}
    {GasPower} = {max gas power}
{e.GasPowerLeft} = {GasPower} - {e.GasPowerUsed}

return {GasPower}, {e.GasPowerLeft}
```
