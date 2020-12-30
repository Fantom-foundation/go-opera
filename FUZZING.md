## Fuzzing

Fuzzing is mainly applicable to packages that parse complex inputs (both text and binary), and is especially useful for hardening of systems that parse inputs from potentially malicious users (e.g. anything accepted over a network).
To run a p2p protocol fuzzer locally, use the command:

```
make fuzz
```
That command should generate a `gossip-fuzz.zip` in the `fuzzing/` directory and runs fuzzer:

```
2020/12/30 22:50:51 workers: 0, corpus: 1 (3s ago), crashers: 0, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 3s
2020/12/30 22:50:54 workers: 0, corpus: 1 (6s ago), crashers: 0, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 6s
2020/12/30 22:50:57 workers: 3, corpus: 1 (9s ago), crashers: 0, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 9s
2020/12/30 22:51:00 workers: 3, corpus: 1 (12s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 12s
2020/12/30 22:51:03 workers: 3, corpus: 1 (15s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 15s
2020/12/30 22:51:06 workers: 3, corpus: 1 (18s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 18s
2020/12/30 22:51:09 workers: 3, corpus: 1 (21s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 21s
2020/12/30 22:51:12 workers: 3, corpus: 1 (24s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 24s
2020/12/30 22:51:15 workers: 3, corpus: 1 (27s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 27s
2020/12/30 22:51:18 workers: 3, corpus: 1 (30s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 30s
2020/12/30 22:51:21 workers: 3, corpus: 1 (33s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 33s
2020/12/30 22:51:24 workers: 3, corpus: 1 (36s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 36s
2020/12/30 22:51:27 workers: 3, corpus: 1 (39s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 39s
2020/12/30 22:51:30 workers: 3, corpus: 1 (42s ago), crashers: 1, restarts: 1/0, execs: 0 (0/sec), cover: 0, uptime: 42s
```

See [github.com/dvyukov/go-fuzz](https://github.com/dvyukov/go-fuzz) to get more details.


### Notes

Once a 'crasher' is found, the fuzzer tries to avoid reporting the same vector twice, so stores the fault in the `suppressions` folder. Thus, if you 
e.g. make changes to fix a bug, you should _remove_ all data from the `suppressions`-folder, to verify that the issue is indeed resolved. 
