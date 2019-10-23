# Lachesis API

Lachesis does support web3.js API, with subtle differences,
due to the graph structure of the ledger.

Due to the support of web3.js, current doc describes only the differences from web3 API.

-   web3 RPC calls reference: <https://github.com/ethereum/wiki/wiki/JSON-RPC>

## Differences from web3.js API

#### Pending blocks

Retrieval of pending blocks isn't supported, because Lachesis doesn't have this
entity.

#### Block Header fields

-   "nonce" is always 0
-   "mixHash" is always 0
-   "sha3Uncles" is always 0
-   "logsBloom" is always 0
    (Lachesis uses the DB index for efficient logs search, instead of Bloom filters)
-   "mixHash" is always 0
-   "miner" field is an undefined address
-   "difficulty" is always 0
-   "extraData" is always empty
-   "gasLimit" is always 0xFFFFFFFFFFFFFFFF
    (Lachesis has a different transaction authorization mechanism)
-   "transactionsRoot" is always 0
-   "receiptsRoot" is always 0
-   "timestampNano" additional field, returns block's consensus time in UnixNano

#### Transactions Receipt fields

-   "logsBloom" is always 0
    (Lachesis uses the DB index for efficient logs search, instead of Bloom filters)

#### Not supported namespaces

-   shh
-   db
-   bzz

## New API calls (comparing to web3 API)

#### debug_getEvent

returns Lachesis event by hash or short ID

##### Parameters

1.  `String`, - full event ID (hex-encoded 32 bytes) or short event ID.
2.  `Boolean` - If true it returns the full transaction objects, if false only the hashes of the transactions.

##### Returns

`Object` - An event object, or null when no event was found:

-   `version`: `QUANTITY` - the event version.
-   `epoch`: `QUANTITY` - the event epoch number.
-   `seq`: `QUANTITY` - the event sequence number.
-   `hash`: `DATA`, 32 Bytes - full event ID.
-   `frame`: `QUANTITY` - event's frame number.
-   `isRoot`: `Boolean` - true if event is root.
-   `creator`: `DATA`, 20 Bytes - the address of the event creator (validator).
-   `prevEpochHash`: `DATA`, 32 Bytes - the hash of the state of previous epoch
-   `parents`: `Array`, - array of event IDs
-   `gasPowerLeft`: `QUANTITY` - event's not spent gas power.
-   `gasPowerUsed`: `QUANTITY` - event's spent gas power.
-   `lamport`: `QUANTITY` - event's Lamport index.
-   `claimedTime`: `QUANTITY` - the UnixNano timestamp of creator's local creation time.
-   `medianTime`: `QUANTITY` - the UnixNano timestamp of the secure median time.
-   `extraData`: `DATA` - the "extra data" field of this event.
-   `transactionsRoot`: `DATA`, 32 Bytes - the root of the transaction trie of the event.
-   `transactions`: `Array` - Array of transaction objects, or 32 Bytes transaction hashes depending on the last given parameter.

##### Example with short ID

    curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_getEvent","params":["1:3:a2395846", true],"id":1}' localhost:18545

##### Example with full ID

    curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_getEvent","params":["0x00000001000000039bcda184cc9e2b20386dcee5f39fe3c4f36f7b47c297ff2b", true],"id":1}' localhost:18545

#### debug_getEventHeader

returns the Lachesis event header by hash or short ID.

##### Parameters

1.  `String`, - full event ID (hex-encoded 32 bytes) or short event ID.

##### Returns

`Object` - An event object, or null when no event was found:

-   `version`: `QUANTITY` - the event version.
-   `epoch`: `QUANTITY` - the event epoch number.
-   `seq`: `QUANTITY` - the event sequence number.
-   `hash`: `DATA`, 32 Bytes - full event ID.
-   `seq`: `QUANTITY` - the event sequence number.
-   `frame`: `QUANTITY` - event's frame number.
-   `isRoot`: `Boolean` - true if event is root.
-   `creator`: `DATA`, 20 Bytes - the address of the event creator (validator).
-   `prevEpochHash`: `DATA`, 32 Bytes - the hash of the state of previous epoch
-   `parents`: `Array`, - array of event IDs
-   `gasPowerLeft`: `QUANTITY` - event's not spent gas power.
-   `gasPowerUsed`: `QUANTITY` - event's spent gas power.
-   `lamport`: `QUANTITY` - event's Lamport index.
-   `claimedTime`: `QUANTITY` - the UnixNano timestamp of creator's local creation time.
-   `medianTime`: `QUANTITY` - the UnixNano timestamp of the secure median time.
-   `extraData`: `DATA` - the "extra data" field of this event.
-   `transactionsRoot`: `DATA`, 32 Bytes - the root of the transaction trie of the event.

##### Example with short ID

    curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_getEventHeader","params":["1:3:a2395846"],"id":1}' localhost:18545

##### Example with full ID

    curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_getEventHeader","params":["0x00000001000000039bcda184cc9e2b20386dcee5f39fe3c4f36f7b47c297ff2b"],"id":1}' localhost:18545

#### debug_getHeads

returns IDs of all the events with no descendants in current epoch

##### Parameters

none

##### Returns

`Array` - Array of event IDs

##### Example

    curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_getHeads","params":[],"id":1}' localhost:18545

#### debug_getConsensusTime

returns event's consensus time, if event is confirmed (more accurate than `MedianTime`)

##### Parameters

1.  `String`, - full event ID (hex-encoded 32 bytes) or short event ID.

##### Returns

`QUANTITY` - the UnixNano timestamp of the secure and accurate consensus time.

##### Example with short ID

    curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_getConsensusTime","params":["1:3:a2395846"],"id":1}' localhost:18545

###### Example with full ID

    curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_getConsensusTime","params":["0x00000001000000039bcda184cc9e2b20386dcee5f39fe3c4f36f7b47c297ff2b"],"id":1}' localhost:18545
