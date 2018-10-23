**How will Fantom prevent attacks?**

**Introduction**

In a decentralized environment one of the key design considerations is
the presence of adversarial attackers. This is further compounded by the
financial nature of blockchains.

We will define common attacks and how they are circumvented as well as
speculate on potential new attack vectors that might arise due to our
design

**Attack Vectors**

**Transaction Flooding**

A malicious participant may run a large number of valid transactions
from their account under their control with the purpose of overloading
the network. In order to prevent such a case, the chain intends to
impose a minimal transaction fee. Since there is a transaction fee, the
malicious user cannot continue to perform such attacks. Participants who
participate in nodes are rewarded, and those who contribute to the
ecosystem, such as by running transactions, are continuously rewarded.
Such rewards are expected to be adequate in running transactions for
appropriate purposes. However, since it would require tremendous cost to
perform abnormal attacks, it would be difficult for a malicious attacker
to create transaction flooding.

**Parasite Chain**

In a DAG-based protocol, a parasite chain can be made with a malicious
purpose, attempting connection by making it look like a legitimate event
block. When the Main Chain is created, verification for each event block
is performed. In the verification process, any block that is not
connected to the Main Chain is deemed to be invalid and is ignored, as
in the case of double spending.

We suppose that less than one-third of nodes are malicious. The
malicious nodes create a parasite chain. By the witness definition,
witnesses are nominated by 2n/3 node awareness. A parasite chain is only
shared with malicious nodes that are less than one-third of
participating nodes. A parasite chain is unable to generate witnesses
and have a shared consensus time.

**Double Spending**

A double spend attack is when a malicious entity attempts to spend their
funds twice. Entity A has 10 tokens, they send 10 tokens to B via node A
and 10 tokens to C via node Z. Both node A and node Z agree that the
transaction is valid, since A has the funds to send to B (according to
A) and C (according to Z).

Consensus is a mechanism whereby multiple distributed parties can reach
agreement on the order and state of a sequence of events. Let’s consider
the following 3 transactions

*txA*: A (starting balance of 10) transfers 10 to B

*txB*: B (starting balance of 0) transfers 10 to C

*txC*: C (starting balance of 0) transfers 10 to D

In a centralized environment the above doesn’t present any problems, we
simply process the results asynchronously. In a distributed environment
however we cannot be assured of the ordering.

We consider Node A that received the order *txA txB txC*

The state of Node A is A:0, B:0, C:0, D:10

Now, we consider Node B that receives the order *txC txB txA*

The state of Node B is A:0, B:10, C:0, D:0

Consensus ordering gives us a sequence of events.

*If the pair of event blocks (x, y) has a double spending transaction,
the chain can structurally detect the double spend and delay action for
the event blocks until the event blocks assign time ordering.*

Suppose that the pair of event blocks (x, y) has same generation g.
Then, all nodes must detect two event blocks before generation g+2. By
the witness definition, each witness supra-shares more than 2n/3
previous witnesses. For this reason, when two witnesses in g + 1 are
selected, they must supra-share same the witnesses which are more than
one-thirds of witnesses in g. This means that more than n/3 witnesses in
g + 1 share both two witnesses which include the pair respectively. With
the witness definition and previous explanation, all witness in g + 2
share both the pairs. Thus, all nodes detect the double spending event
blocks at g+2 or earlier.

**Long-range attack**

In blockchains an adversary can create another chain. If this chain is
longer than the original, the network will accept the longer chain. This
mechanism exists to identify which chain has had more work (or stake)
involved in its creation.

2n/3 participating nodes are required to create a new chain. To
accomplish a long-range attack you would first need to create &gt; 2n/3
participating malicious nodes to create the new chain.

**Bribery attack**

An adversary could bribe nodes to validate conflicting transactions.
Since 2n/3 participating nodes are required, this would require the
adversary to bribe &gt; 1n/3 of all validators to begin a bribery
attack.

**Denial of Service**

We are a leaderless system requiring 2n/3 participation. An adversary
would have to deny &gt; 1n/3 participants to be able to successfully
mount a DDoS attack.

**Sybil**

Each participating node must stake a minimum amount of FTM to
participate in the network. Being able to stake 2n/3 total stake would
be prohibitively expensive

**Quantum Attacks**

Cryptographic protocols are susceptible to attack by the development of
a sufficiently large quantum computer. Of specific interest to
cryptocurrencies is how this relates to proof-of-work and more
specifically, the elliptic curve signature scheme. Optimistic estimates
state that this can be broken by a quantum computer as early as 2027, it
is therefore important to adopt a post-quantum signature scheme.

Signatures are often based on the Elliptic Curve Digital Signature
Algorithm secp256k1 curve. The security of this system is based on the
hardness of the Elliptic Curve Discrete Log Problem (ECDLP).

How quickly can a quantum computer compute the Elliptic Curve Discrete
Log Problem? An instance with a *n* bit prime field, can be solved using
9n + 2 \[log2(n)\]+10 logical qubits and (448log2(n)+4090)n3 Toffoli
gates. Bitcoin uses n=256 bit signatures.

For 10GHz clock speed and error rate of 10−5 , the signature is cracked
in 30 minutes using 485550 qubits.

So if all Elliptic Curve Digital Signature Algorithms are susceptible,
then how can you implement a quantum proof solution?

<img src="./attacks/image1.png" style="width:6.27083in;height:3.22222in" />

In blockchain context we care about signature and public key lengths
(since these have to be stored to fully verify).

Hash based schemes like XMSS have provable security. Grover’s algorithm
can still be used to attack. DILITHIUM at 138 bits require time 2125

A truly quantum proof cryptographic algorithm does not currently exist.
Instead, the our architecture allows for multiple cryptographic
implementations to be plug and play, given the modular architecture
design. Since we aren’t focusing on tightly coupled architecture, it
means we could implement ECC, XMSS, and DILITHIUM (and something new we
haven’t announced yet).
