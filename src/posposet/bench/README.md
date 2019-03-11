# temporary package for comparison

## RLP vs ProtoBuf serialization 

```go generate .```


### Result:

```
BenchmarkRlp-4            200000              5571 ns/op            1888 B/op         35 allocs/op
BenchmarkProto-4          500000              3636 ns/op            1424 B/op         22 allocs/op
```

![CPU prof output](./cpu_prof.svg)

### Conclusion:

ProtoBuf serialization is more then 2 times faster then RLP.
