
#### Fork definitions

Fork pair is a pair of fork events {a, b}, such that:
```
a != b AND a.Seq == b.Seq && a.Epoch == b.Epoch && a.Creator == b.Creator
```

The creator of a fork pair is called ```cheater```.
Cheater is always a Byzantine node, but not vice versa.

Simply said, the fork is created when validator doesn't use his last event as self-parent.

#### No "longest chain rule"

Unlike blockchains (especially with consensuses like PoW, round-robin PoS or coinage-based PoS),
Lachesis stores and connects all the fork events,
because they are valid events, which contain information
about misbehaving. When a fork event is connected,
Lachesis node doesn't "choose" only one "branch", it continues to
normally store and process the whole graph.

The Lachesis consensus will result into the exact same blocks regardless of fork events,
unless more than 1/3W are Byzantine.

#### Protections

To protect the consensus from forks, the stricter ```forklessCause```
relation is used in Atropos election instead of ```happened-before``` relation.

When a fork pair gets confirmed, the validator gets a harsh penalty.

See [attack](attack.md) page for more detailed information on these and other
protections related to forks.
