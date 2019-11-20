# Step by step example:

1. Build docker images:

```sh
git checkout scope3
git pull origin scope3
cd docker
make
```

2. Run 3 lachesis-node containers:

```sh
N=3 ./start.sh
```

> b4eb25c9857162e24fa6a80086672d6058b0e61e7a5f2e0faacaeefede3d0d3b
>
> Start 3 nodes:
>
> 3a300b1e24d50e3a946baded2deff78d70e0f84008c674dc92ec72b486979988
> 891f50d0402bf38472117f54793ba6bdcd3fea782345d93d8c2b802157bf0390
> 632c994950ac973ae5a664fd4f62fa7e155b1672f865de814fee3c5345f1d8a6
> Connect nodes to ring:
> 
>  getting node-2 address:
>     p2p address = "enode://e453fc451bd9fbea12014db2147ce73139ad43c87a716038edae27fc0402fc54fa365d35fa16e6730403fd94cadbe898f7215839a983d66747e352957def1793@172.21.0.3:5050"
>  connecting node-1 to node-2:
>     result = true
> ...
>  getting node-1 address:
>     p2p address = "enode://9ff17f1f84f6be3d788f0f889c99d9ccedc41740535839f0be6a74645f4de658c2308f90f75b9f1daa50f215f870552995784c23a3b3322e57f3ebf06d3cec56@172.21.0.2:5050"
>  connecting node-3 to node-1:
>    result = true


3. Run 3 tx-storm containers:

```sh
N=3 ./txstorm-on.sh
```

> Start 3 tx generators:
> 
> cb203c57b2da0d8581678eb8a8fe3cb90ee790b2cba96011a4bfe17523e37872
> 36784122107e9e4d801700419ed996b8e78dafbf3a08a9c8ded57d8b22fbead9
> 8e037e6538896850097b6ba53fd62e10ae80090224aeda031943f4d22ae4fff7


4. Run prometheus container:

```sh
./prometheus-on.sh 
```

> Start Prometheus:
> 
> 99ce607579fa2511fec3eb275b51fd1a7272d672a13ed7ca662948ca6c3b78d3

5. See metrics visualization at:

  [TPS](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_tx_count_got&g0.tab=0)

  [Tx latency](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_tx_latency&g0.tab=0)

  [Sent to lachesis node transactions](http://localhost:9090/graph?g0.range_input=5m&g0.expr=lachesis_tx_count_sent&g0.tab=0)

Links will be available if prometheus container is running.


6. See logs:

```sh
docker logs --tail 100 -f node1
docker logs --tail 100 -f txgen1
```

7. Stops the all:
```sh
./stop.sh
```
