**How will Fantom take care of on chain data storage?**

**Background**

We identify the following areas on on-chain storage

1.  Smart Contract State

2.  DLT State

3.  Transactions

4.  Immutable Data Storage

5.  Mutable Data Storage

**Universal State (Smart Contract State & DLT State)**

State is stored on an account level. An account has ownership of its
storage. This storage is captured in a trie.

**Understanding Smart Contract State**

Using the EVM as an example. State is a mapping between addresses
(160-bit identifiers) and account states. Account states are Recursive
Length Prefix (RLP) encoded data structures. This mapping is maintained
in a modified Merkle Patricia tree. The root node of this structure is
cryptographically dependent on all internal data.

The account state consists of the following fields

1.  Nonce

2.  Balance

3.  StorageRoot

4.  CodeHash

StorageRoot is a 256-bit hash of the root node of a Merkle Patricia tree
which is an encoded mapping between 256-bit integer values.

**Understanding DLT State**

The current design of blockchains, requires replaying all data from
transaction index 0 to deterministically arrive at the current state.
This requires full chain data to be stored. This is how to achieve the
current accurate UTXOs or State. Every transaction created needs to be
stored, shared, and computed. The current distribution mechanism for
this is blocks.

This design has no built in archiving strategy, the ledger will grow
infinitely and more storage must be consequently added.

This further increases the barrier to entry leading to less
decentralization. A mobile device would currently not be able to
participate in the Ethereum network unless it can store 1TB worth of
data.

Append only ledgers grow infinitely.

**Defining Immutable Data Storage**

Storing data in a smart contract requires 32,000 ETH per 1GB of data (at
50 gwei). This is a fixed ratio and ignores the price of the token. We
will show that the value of a token is correlated to the production
value of the ecosystem. You can therefore not directly correlate
production value as a fixed value towards storage.

**On-Chain Storage**

Storage areas identified include

-   Full transaction history for full nodes

-   Light hash history for light clients

-   Snapshot state storage for

-   Accounts

-   Smart Contract Code

-   Smart Contract State

**Legacy Design**

For each node;

-   Store all blocks

-   Store all transactions

-   Store each account state

-   Store each smart contract state

Purpose of storing all blocks

Blocks allow for anchored security, we need the sequence of blocks to be
able to accurately define the sequence of transactions. Blocks allow for
transaction ordering.

Purpose of storing all transactions

Each transaction is required to be able to arrive at the current state.
We accomplish this by taking transaction index 0 and applying each
transaction until index n to arrive at the current state answer.

Purpose of storing account & smart contract state

Each node stores each account state to be able to be 99.99% fault
tolerant. Even if only 1 node remains the system will function.

-   Blocks are append only lists.

-   Transactions are append only lists.

-   Account size is correlated to amount of accounts on the network.

-   Smart contract state is variable.

**Data Strategies**

The following strategies have been identified

-   Signature snapshots

-   Mimblewimble (UTXO only)

-   State transition proofs

-   Data sharding

**Signature snapshots**

Each block contains the root signature of the account state and a list
of all participating validators. Blocks are signed by all participating
validators. This is considered the signed state.

This can be used as the proof of consensus state to third parties.

This requires a secondary structure that lists each address and their
stake for each round. Cryptographically anchored to each previous round
until the genesis round.

The hash of the genesis round is a unique identifier of the ledger.

**Mimblewimble**

Bitcoin is categorized as a UTXO based system. UTXO define Unspent
Transaction Outputs. An account has Inputs described as;

type Transaction struct {

ID \[\]byte

Vin \[\]TXInput

Vout \[\]TXOutput

}

type TXInput struct {

Txid \[\]byte

Vout int

Signature \[\]byte

PubKey \[\]byte

}

type TXOutput struct {

Value int

PubKeyHash \[\]byte

}

A transaction consists of Inputs (value received by the account), and
outputs (values sent from the account).

Alice has input 1 BTC from Bob, then Bob has output 1 BTC to Alice. If
Alice then wishes to send 0.5 BTC back to Bob, and 0.5 BTC to Charlie,
Alice would create a transaction with 1 BTC input from Bob, 0.5 output
to Bob, and 0.5 output to Charlie.

The Mimblewimble specification has the concept of UTXO compacting.
Called Cut-through, it is explained as follows

Blocks let miners assemble multiple transactions into a single set
that’s added to the chain. In the following block representations,
containing 3 transactions, we only show inputs and outputs of
transactions. Inputs reference outputs they spend. An output included in
a previous block is marked with a lower-case x.

I1(x1) — — O1

\|- O2

I2(x2) — — O3

I3(O2) -\|

I4(O3) — — O4

\|- O5

We notice the two following properties:

Within this block, some outputs are directly spent by included inputs
(I3 spends O2 and I4 spends O3).

The structure of each transaction does not actually matter. As all
transactions individually sum to zero, the sum of all transaction inputs
and outputs must be zero.

Similarly to a transaction, all that needs to be checked in a block is
that ownership has been proven (which comes from transaction kernels)
and that the whole block did not add any money supply (other than what’s
allowed by the coinbase). Therefore, matching inputs and outputs can be
eliminated, as their contribution to the overall sum cancels out. Which
leads to the following, much more compact block:

I1(x1) \| O1

I2(x2) \| O4

\| O5

Through compacting, Mimblewimble can reduce all UTXO to single pairs.
This allows the chain to be drastically compacted.

The same can be achieved with finalized transactions and state output.

We redefine a transaction as the sum of its inputs and outputs and
derive a current state. We can group all similar transactional outputs
into single inputs for state transitions.

If Alice sent Bob 0.5 BTC, and Charlie 0.5 BTC, followed by Bob sending
0.25BTC to Charlie, this can be represented as Alice sends 0.25 BTC to
Bob and 0.75 BTC to Charlie. This is the concept behind compacting.

Applied over the entire blockchain, we can considerably reduce the
amount of transactions that need to be replayed to arrive at the current
state, reducing size and sync speed.

**State Proofs**

We have computation *c(a,b)* to execute. We give *c(a,b)* to untrusted
parties. We assume standard 2n/3 fault tolerance, so we send *c(a,b)* to
3 parties. 2 parties return with the same results for *c(a,b)*. We
assume the result for *c(a,b)* is correct. This is verifiable computing.

VM execution occurs on chain because we use multi party consensus to
verify computing. We provide proof with zero knowledge that execution
occurred in a trusted manner. Execution no longer needs to occur
on-chain.

A zero knowledge proof VM, can ensure verified computing.

We have secured execution, but we still have outputs, for example ERC20
addresses and balances.

We knew state *s1*, and we can prove that transition occurred, we can
prove *s2*. We need *s1* and transactions *tn* to prove *s2*. Given
atomic state output *o* and transition proof , we can prove that
participant *p* has balance *b*.

To have verified data, all the chain needs to save is the atomic state
output and the proof .

t = (s1,tn)=s2 m

setup(t)

t with witness w

prove(,tn,w)

verify(,tn,)

Consider a block, a block is our proof of transactions in it. A block is
proof of a state transition. State *s1* when applied with block *bn*
gives us state *s2*. So *bn* is a state transition.

**Data Sharding**

Goals

-   Reduction in data size

-   Sustainable data growth

-   Secure storage

-   Byzantine fault tolerant

-   Proof of Storage

Technology

-   Erasure coding

-   Reed-Solomon codes

-   100–200 shards Large-Scale Reed-Solomon Codes

Erasure coding allows us to set our fault tolerance, these allow us to
expand our data set with data redundancy, at an increase of storage
requirement. This may seem counter intuitive, but this allows us to have
a greater amount of shards in a given network.

We could have erasure codes as high as 99% tolerance, which allows for
99% of all nodes in the network to disappear and we will still be able
to recover our data.

The erasure codes along with large scale sharding allows us to store any
of the above entities across multiple participants at a fraction of the
storage requirement.

**Conclusion**

At this point we have identified a few areas and addressed a few
strategies. But there isn’t a once size fits all strategy.

-   To address transaction history, we adopt signed state and state
    > proofs.

-   To address smart contract state, we adopt data sharding techniques.

-   To address DLT state, we adopt data sharding techniques and state
    > proofs.

-   To address immutable and mutable secondary storage, we adopt a new
    > marketplace driven storage solution.

There are more nuances here, that we will discuss in further detailed
articles.
