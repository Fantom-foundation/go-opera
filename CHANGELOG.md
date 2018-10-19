## UNRELEASED

SECURITY:

FEATURES:

IMPROVEMENTS:

BUG FIXES:

## v0.4.0 (October 14, 2018)

SECURITY:

* keygen: write keys to files instead of tty. 

FEATURES:

* proxy: Introduced in-memory proxy.
* cmd: Enable reading config from file (lachesis.toml, .json, or .yaml)

IMPROVEMENTS:

* node: major refactoring of configuration and initialization of Lachesis node.
* node: Node ID is calculated from public key rather than from sorting the 
peers.json file.

## v0.3.0 (September 4, 2018)

FEATURES:

* poset: Replaced Leemon Baird's original "Fair" ordering method with 
Lamport timestamps.
* poset: Introduced the concept of Frames and Roots to enable initializing a
poset from a "non-zero" state.
* node: Added FastSync protocol to enable nodes to catch up with other nodes 
without downloading the entire poset. 
* proxy: Introduce Snapshot/Restore functionality.

IMPROVEMENTS:

* poset: Refactored the consensus methods around the concept of Frames.
* poset: Removed special case for "initial" Events, and make use of Roots 
instead. 
* docs: Added sections on Lachesis and FastSync.