# temporary package for comparison

## RLP vs ProtoBuf serialization 

```go generate .```


### Result:

```
BenchmarkRlp-4            300000              4973 ns/op            1825 B/op         33 allocs/op
BenchmarkProto-4         1000000              2096 ns/op            1360 B/op         20 allocs/op
```

![CPU prof output](./cpu_prof.svg)

### Conclusion:

ProtoBuf serialization is more then 2 times faster then RLP.
