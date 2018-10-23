**Fantom:**

**Distributed Computing**

**[Introduction](#introduction) 2**

**[Background](#background) 2**

> [Price of Computing](#price-of-computing) 2
>
> [Compute vs Storage](#compute-vs-storage) 3
>
> [Compute Architecture](#compute-architecture) 3
>
> [Tightly Coupled Architecture](#tightly-coupled-architecture) 3

**[Redesigning Stateless Computing](#redesigning-stateless-computing)
3**

**[Time Complexity](#time-complexity) 4**

**[Job Scheduling](#job-scheduling) 4**

**[Verifiable Computing](#verifiable-computing) 5**

**[Deterministic Verifiable
Computing](#deterministic-verifiable-computing) 9**

**[Decoupling Cost](#section-1) 9**

**[Job Scheduling Conflicts](#job-scheduling-conflicts) 10**

**[Job Scheduling](#job-scheduling-1) 10**

> [Job Propagation k/n](#job-propagation-kn) 10
>
> [Job Propagation Leader Selection](#job-propagation-leader-selection)
> 11
>
> [Job Propagation Cryptographic
> Sortition](#job-propagation-cryptographic-sortition) 11

**[Deterministic Verifiable Job Propagation Compute
Marketplace](#section-2) 11**

**[Conclusion](#conclusion) 12**

 
=

Introduction
============

Distributed Computing in a decentralized context is currently only as
efficient as the consensus winning node. The goal is to have the
combined computing power of all participants. Instead we have the
computing power of a single node. We propose a cheap, asynchronous,
fault tolerant, verifiable alternative to current decentralized
solutions.

Background
==========

Price of Computing
------------------

The current computing standard is cloud computing services. AWS, Google,
Azure. We will consider AWS as our benchmark.

t2.medium on-demand: $0.0464 per Hour

t2.medium spot-pricing: $0.026 per Hour

t2.medium reserved instance: $0.027 per Hour

Lambda: $0.20 per 1 million requests, $0.00001667 per GB-second of
compute

Aurora(RDS): $0.1 per GB-month, $0.2 per 1 million requests

Cost Example

Low-traffic Account CRUD API (Lambda)

\~400,000 requests per month

$0.25/mo (@200ms per request)

Low-traffic Account CRUD API (Aurora - RDS)

\~400,000 requests per month

$0.3/mo

Comparing an Ethereum Solution

Low-traffic Account CRUD API (Ethereum)

\~400,000 requests per month (Average between inserts, updates, and
balance requests)

3.3 billion gas @ 50 gwei = 168 ETH = $68,208.00

Cost differential

The principal of a world computer is to be able to leverage all of the
computing capacity of all participating nodes. Given the Proof-of-Work
consensus design, only a single block can be created every epoch. All
nodes must execute the same block to arrive at the same state. Thus,
instead of having the cumulative computing power of all nodes, we have
the computing power of a single node.

Compute vs Storage
------------------

AWS S3: $0.023 / GB

AWS Glacier $0.004 / GB

There is no linear correlation between compute capacity, and storage
capacity. Storage and Computing can not be tightly coupled in a
decentralized environment.

Compute Architecture
--------------------

-   CPU/GPU perform stateless computing.

-   Memory allows for temporary fast access storage.

-   Disk allows for permanent\* slow access storage.

Compute architecture leverages off-of the above 3 principals. Fast
computing loads from slow storage what it needs into fast access memory
to allow for computing to manipulate fast access memory and then write
back to slow storage.

Tightly Coupled Architecture
----------------------------

Storage and Compute in Ethereum VM are tightly coupled. They are
inherent in the OPcodes of the virtual machine.

With no correlation between Computing and Storage, and no correlation in
Compute Architecture, this tight coupling must be removed to be
feasible.

Redesigning Stateless Computing
===============================

Network of *n* participants.

Data source of *ds*.

Execution set of *c*.

*1/n* compute nodes executes *c* with inputs *ds*.

Compute Node Risk

1.  Infinite execution time.

2.  Out of memory.

Compute Provider Risk

1.  Node terminates before execution finishes.

2.  Inaccurate execution results.

Problem Statement \#1

The compute node must be able to assess the compute requirements of the
computation.

Problem Statement \#2

The compute node must be able to assess the memory requirements of the
computation.

Problem Statement \#3

The compute provider must be able to have confirmation of execution
results.

Problem Statement \#4

The compute provider must be able to confirm the validity of the
resulting output.

Time Complexity
===============

Constant O(1)

Logarithmic O(log n)

Polylogarithmic O((log n)<sup>k</sup>)

Sub-linear o(n) or O(n<sup>1/2</sup>)

Linear O(n)

Quasilinear O(n log<sup>k</sup> n)

Sub-quadratic o(n<sup>2</sup>)

Polynomial O(n<sup>k</sup>)

Key qualities of a compute job would need to include

-   Determinism

-   Precise computational steps

The Ethereum Virtual Machine established this by linking Gas costs to
each OPcode execution. Thus calculating execution time\* and memory
requirements\* for a given set of inputs. This cost is tightly coupled
with execution.

A compute node can given a set of inputs (data source) and set of
instructions be able to assess the execution and memory requirements of
the specific job.

Job Scheduling
==============

Related works in centralized scheduling

Cloud task and virtual machine allocation strategy[1]

-   Total execution time of all tasks

Optimal scheduling of computational tasks[2]

-   Execution time - Virtual Machine Tree

A deadline scheduler for jobs[3]

-   Execution time and cost - Cloud Least Laxity First

Job scheduling algorithm based on Berger[4]

-   Execution time

Online optimization for scheduling preemptable tasks on Iaas cloud
systems[5]

-   Computational power, execution time, bandwidth - Dynamic allocation

A priority based job scheduling algorithm[6]

Performance and cost evaluation of Gang Scheduling in a Cloud Computing
system with job migrations and starvation handling[7]

-   Response time and bounded slowdown

Given *n* nodes we can make the following assumptions.

-   Optimized efficiency has exactly *1/n* nodes execute the given
    > compute task.

-   Fault tolerance has at least *3/n* nodes execute the given compute
    > task.

-   Multi Party Computing (MPC) for verifiable results has at least 3/n
    > nodes execute the given compute task.

Given that MPC requires *3/n* and fault tolerance requires *3/n*, we
require *9/n* for fault tolerant multi party computation. This is an
optimized strategy.

Verifiable Computing
====================

Related works in Interactive Proofs

Verifiable Computation with Massively Parallel Interactive Proofs[8]

-   GPU parallel processing for arithmetic circuits of polylogarithmic
    > depth

-   No privacy

-   Publicly Verifiable

-   Efficient

-   Circuits of polylogarithmic depth

Allspice: A Hybrid Architecture for Interactive Verifiable
Computation[9]

-   Verify computations expressed as straight-line programs.

-   No privacy

-   Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Pepper: Making Argument Systems for Outsourced Computation Practical
(Sometimes)[10]

-   Arithmetic constraints instead of circuits, very limited class of
    > functions

-   No privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Ginger: Taking Proof-Based Verified Computation a Few Steps Closer to
Practicality[11]

-   Improvement on Pepper, larger class of computations

-   No privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Zataar: Resolving the Conflict Between Generality and Plausibility in
Verified Computation[12]

-   PCP using algebraic representations as Quadratic Arithmetic Programs

-   No privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Pantry: Verifying Computations with State[13]

-   Zataar improvement which allows verification of stateful
    > computations

-   No privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Stateful

River: Verifiable Computation with Reduced Informational Costs and
Computational Costs[14]

-   Quadratic Arithmetic Program based verifiable computing system.

-   No privacy

-   Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Buffet: Efficient RAM and control flow in verifiable outsourced
computation[15]

-   Support for general loops

Related works in Non-Interactive Proofs

Pinocchio: Nearly Practical Verifiable Computation[16]

-   Arithmetic circuits, no input-output privacy

Gepetto: Versatile Verifiable Computation[17]

-   Public verifiability, no input-output privacy.

-   No Privacy

-   Publicly Verifiable

-   Amortized Efficiency

-   General Loops

SNARKs for C: Verifying Program Executions Succinctly and in Zero
Knowledge[18]

-   Increased overhead. Based on Quadratic Arithmetic Programs. Public
    > verifiability without input-output privacy.

-   No Privacy

-   Publicly Verifiable

-   Amortized Efficiency

-   General Loops

Succinct Non-Interactive Zero Knowledge for a von Neumann
Architecture[19]

-   Quadratic Arithmetic Program based SNARK with more efficient
    > verification and proof generation. Universal circuit generator.

-   No Privacy

-   Publicly Verifiable

-   Amortized Efficiency

-   General Loops

ADSNARK: Nearly Practical and Privacy-Preserving Proofs on Authenticated
Data[20]

-   Straight-line computations on authenticated data. Publicly
    > verifiable and more efficient privately verifiable proof

-   Input Only Privacy

-   Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Block Programs: Improving Efficiency of Verifiable Computation for
Circuits with Repeated Substructures[21]

-   More efficient way to handle loops in a program

Related works in Fully Homomorphic Encryption

Non-Interactive Verifiable Computing: Outsourcing Computation to
Untrusted Workers[22]

-   Yao’s garbled circuits with Fully Homomorphic Encryption, privacy
    > preserving

-   Input & Output Privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Improved Delegation of Computation Using Fully Homomorphic
Encryption[23]

-   Input & Output Privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Arithmetic Circuits

Related works based on Message Authentication Codes

Verifiable Delegation of Computation on Outsourced Data[24]

-   Bilinear maps and closed-form efficiency. Preprocessing stage.

-   No Privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Polynomial of Degree 2

Generalized Homomorphic macs with Efficient Verification[25]

-   *ζ*-linear maps, supports arithmetic circuits of depth *ζ*

Efficiently Verifiable Computation on Encrypted Data[26]

-   Multivariate polynomials of degree 2 with input privacy

-   Input Privacy

-   Not Publicly Verifiable

-   Amortized Efficiency

-   Polynomial of Degree 2

A number of strategies exist for verifiable computing.

-   Interactive Knowledge Proofs

    -   Interactive

    -   Variable attestation

    -   Computation Overhead

    -   Example: Buffet

-   Non-Interactive Knowledge Proofs

    -   Trusted Setup\*

    -   Fixed attestation

    -   Computational Overhead

    -   Example: ADSNARK, zkSNARK

-   Trusted Execution Environments (TEE)

    -   Hardware Requirement

    -   Centralization

    -   High Security

    -   Fast Execution

    -   Example: TPM, TXT, Intel SGX, ARM TrustZone

-   Multi Party Computation (MPC)

    -   Multiple Participants

    -   Unequal reward structures

    -   Example: Ethereum

Viable Options include

-   zk-SNARK wrapped Virtual Machine (Execution overhead, low barrier to
    > entry)

-   TEE wrapped Virtual Machine (Execution optimized, high barrier to
    > entry)

-   MPC with Probabilistic Payments (Execution inefficient, low barrier
    > to entry)

Deterministic Verifiable Computing
==================================

Given;

-   Fixed cost Opcodes

-   Verifiable Computing

-   Multi node fault tolerance redundancy

We can accomplish;

-   Fixed length execution assessment

-   Compute cost execution assessment

-   Verifiable Computing Results

-   Fault Tolerance

Decoupling Cost
===============

A fee based token economy marketplace will be required to ensure fair
value.

Tasks are associated with a bid: the amount of coin offered for
performing the task. Bidders can adjust the bid on their pending tasks
to expedite their execution. Workers will decide if a bid is sufficient
for them to execute upon. Workers will execute work when profitable for
them to do so.

A compute node can decide to execute a given job for less than the
expected fee payment. A compute provider can provide a fee less than the
indicated execution cost.

Job Scheduling Conflicts
========================

Given

-   *n* nodes

-   *k* as our fault tolerance variable

We want *k/n* to execute a given job with the following criteria

-   *k* is willing to accept the job based on the OPcode cost

-   *k* is willing to accept the job based on the execution fee proposed

Risk

-   No nodes willing to accept the job (No profitability attack)

-   *&lt; k* nodes are willing to accept the job

-   All nodes are willing to accept the job (Too profitable attack)

Job scheduling with a scheduler is a fairly well established field of
study. In a decentralized environment, we lack a scheduler.

Job Scheduling 
===============

To achieve decentralized job scheduling we discuss the following
strategies

-   *k/n* selection with a first come leader selection

-   Stake based leader selection

-   Cryptographic Sortition participant selection

Job Propagation k/n
-------------------

Job requests are unique. Each node holds a list of all job IDs they have
received.

Each node that receives a request attempts to act as a job scheduler.

Upon receiving a compute request with an empty job order the compute
node *n<sub>1</sub>* confirms if they are willing to accept the job. If
they are, they sign the request, randomly select *k* nodes, adds each to
the job order and propagates to each node.

Each node receiving the request can accept or deny the request. If they
accept, they sign and propagate to the listed job order nodes. If *y*
nodes have accepted, but *k* has not been reached, *n<sub>1</sub>*
elects *k-y* new nodes and continues this cycle. If
*n<sub>1\ </sub>*knows &lt; *k* nodes *t* it sends the job to *t* nodes
and elects one of *t* nodes as the new leader.

Job Schedulers are allowed to collect the fees.

-   Risk of less than *k* nodes known

-   Denial attacks by *n<sub>1</sub>*

-   Centralized towards *n<sub>1</sub>*

-   Long potential job acceptance time

Job Propagation Leader Selection
--------------------------------

Compute nodes stake tokens for leader election. Each epoch a leader
*n<sub>l</sub>* is elected to be the job scheduler. The elected leader
sends out the job request signed with it as leader. Receiving nodes can
validate that *n<sub>l\ </sub>*is the correct leader for the epoch. If a
node is willing to accept the job they sign the request and send back to
the leader. The leader can then choose a subset of nodes to execute the
given job.

Further optimization: Job scheduling based on cryptographic proofs of
resources available\*

Job Schedulers are allowed to collect the fees.

-   Risk of malicious *n<sub>l</sub>*

-   Denial attacks by *n<sub>l</sub>*

-   Fast job acceptance time

-   Further scheduling optimization possible

Job Propagation Cryptographic Sortition
---------------------------------------

Compute nodes each have an algorithmic lottery chance. The job ID is
used as the seed for the Cryptographic Sortition, all nodes willing to
participate, participate in the sortition. Nodes with a winning match
are allowed to participate.

The highest hash seed matching the winning ticket is awarded the fees if
they complete the job.

-   More than *k* participants

Deterministic Verifiable Job Propagation Compute Marketplace
============================================================

Given;

-   Fixed cost Opcodes

-   Verifiable Computing

-   Multi node fault tolerance redundancy

-   Job Propagation

-   Fee decoupling

We can accomplish;

-   Fixed length execution assessment

-   Compute cost execution assessment

-   Verifiable Computing Results

-   Fault Tolerance

-   k/n selection for efficient computing

-   Cost efficient marketplace

Conclusion
==========

A deterministic, verifiable, market value driven, fault tolerant,
asynchronous world compute engine. We allow for multiple verifiable
compute strategies as well as multiple job selection strategies. The
proposal allows for a compute environment that gravitates towards a
market value compute requirement with low cost resources available for
the ecosystem.

[1] Xu, X., Hu, H., Hu, N., Ying, W.: Cloud task and virtual machine
allocation strategy in cloud computing environment. In: Lei, J., Wang,
F.L., Li, M., Luo, Y. (eds.) NCIS 2012. CCIS, vol. 345, pp. 113–120.
Springer, Heidelberg (2012)

[2] Achar, R., Thilagam, P.S., Shwetha, D., et al.: Optimal scheduling
of computational task in cloud using virtual machine tree. In: 2012
Third International Conference on Emerging Applications of Information
Technology (EAIT), pp. 143–146 (2012)

[3] Perret, Q., Charlemagne, G., Sotiriadis, S., Bessis, N.: A deadline
scheduler for jobs in distributed systems. In: 2013 27th International
Conference on Advanced Information Networking and Applications Workshops
(WAINA), pp. 757–764 (2013)

[4] Baomin, X., Zhao, C., Enzhao, H., Bin, H.: Job scheduling algorithm
based on Berger model in cloud environment. Adv. Eng. Softw. 42, 419–425
(2011)

[5] Li, J., Qiu, M., Ming, Z., Quan, G., Qin, X., Gu, Z.: Online
optimization for scheduling preemptable tasks on IaaS cloud systems. J.
Parallel Distrib. Comput. 72(5), 666–677 (2012)

[6] Ghanbaria, S., Othmana, M.: A priority based job scheduling
algorithm in cloud computing. In: ICASCE 2012, pp. 778–785 (2012)

[7] Moschakis, I.A., Karatza, H.D.: Performance and cost evaluation of
Gang Scheduling in a Cloud Computing system with job migrations and
starvation handling. In: IEEE Symposium on Computers and Communications
(ISCC) (2012) and 2011 IEEE Symposium on Computers and Communications,
pp. 418–423 (2011)

[8] Justin Thaler, Mike Roberts, Michael Mitzenmacher, and Hanspeter
Pfister. Verifiable computation with massively parallel interactive
proofs. In 4th USENIX Workshop on Hot Topics in Cloud Computing,
HotCloud’12, Boston, MA, USA, June 12-13, 2012, 2012.

[9] Victor Vu, Srinath T. V. Setty, Andrew J. Blumberg, and Michael
Walfish. A hybrid architecture for interactive verifiable computation.
In 2013 IEEE Symposium on Security and Privacy, SP 2013, Berkeley, CA,
USA, May 19-22, 2013, pages 223– 237, 2013.

[10] Srinath T. V. Setty, Richard McPherson, Andrew J. Blumberg, and
Michael Walfish. Making argument systems for outsourced computation
practical (sometimes). In 19th Annual Network and Distributed System
Security Symposium, NDSS 2012, San Diego, California, USA, February 5-8,
2012, 2012.

[11] Srinath T. V. Setty, Victor Vu, Nikhil Panpalia, Benjamin Braun,
Andrew J. Blumberg, and Michael Walfish. Taking proof-based verified
computation a few steps closer to practicality. In Proceedings of the
21th USENIX Security Symposium, Bellevue, WA, USA, August 8-10, 2012,
pages 253–268, 2012.

[12] Srinath T. V. Setty, Benjamin Braun, Victor Vu, Andrew J. Blumberg,
Bryan Parno, and Michael Walfish. Resolving the conflict between
generality and plausibility in verified computation. In Eighth Eurosys
Conference 2013, EuroSys ’13, Prague, Czech Republic, April 14-17, 2013,
pages 71–84, 2013.

[13] Benjamin Braun, Ariel J. Feldman, Zuocheng Ren, Srinath T. V.
Setty, Andrew J. Blumberg, and Michael Walfish. Verifying computations
with state. In ACM SIGOPS 24th Symposium on Operating Systems
Principles, SOSP ’13, Farmington, PA, USA, November 3-6, 2013, pages
341–357, 2013.

[14] Gang Xu, George T. Amariucai, and Yong Guan. Verifiable computation
with reduced informational costs and computational costs. In Computer
Security - ESORICS 2014 - 19th European Symposium on Research in
Computer Security, Wroclaw, Poland, September 7-11, 2014. Proceedings,
Part I, pages 292–309, 2014.

[15] Riad S. Wahby, Srinath T. V. Setty, Zuocheng Ren, Andrew J.
Blumberg, and Michael Walfish. Efficient RAM and control flow in
verifiable outsourced computation. In 22nd Annual Network and
Distributed System Security Symposium, NDSS 2015, San Diego, California,
USA, February 8-11, 2014, 2015.

[16] Bryan Parno, Jon Howell, Craig Gentry, and Mariana Raykova.
Pinocchio: Nearly practical verifiable computation. In 2013 IEEE
Symposium on Security and Privacy, SP 2013, Berkeley, CA, USA, May
19-22, 2013, pages 238–252, 2013.

[17] Craig Costello, C´edric Fournet, Jon Howell, Markulf Kohlweiss,
Benjamin Kreuter, Michael Naehrig, Bryan Parno, and Samee Zahur.
Geppetto: Versatile verifiable computation. In 2015 IEEE Symposium on
Security and Privacy, SP 2015, San Jose, CA, USA, May 17-21, 2015, pages
253–270, 2015.

[18] Eli Ben-Sasson, Alessandro Chiesa, Daniel Genkin, Eran Tromer, and
Madars Virza. Snarks for C: verifying program executions succinctly and
in zero knowledge. In Advances in Cryptology - CRYPTO 2013 - 33rd Annual
Cryptology Conference, Santa Barbara, CA, USA, August 18-22, 2013.
Proceedings, Part II, pages 90–108, 2013.

[19] Eli Ben-Sasson, Alessandro Chiesa, Eran Tromer, and Madars Virza.
Succinct noninteractive zero knowledge for a von neumann architecture.
In Proceedings of the 23rd USENIX Security Symposium, San Diego, CA,
USA, August 20-22, 2014., pages 781–796, 2014.

[20] Michael Backes, Manuel Barbosa, Dario Fiore, and Raphael M.
Reischuk. ADSNARK: nearly practical and privacy-preserving proofs on
authenticated data. In 2015 IEEE Symposium on Security and Privacy, SP
2015, San Jose, CA, USA, May 17-21, 2015, pages 271–286, 2015.

[21] Gang Xu, George T. Amariucai, and Yong Guan. Block programs:
Improving efficiency of verifiable computation for circuits with
repeated substructures. In Proceedings of the 10th ACM Symposium on
Information, Computer and Communications Security, ASIA CCS ’15,
Singapore, April 14-17, 2015, pages 405–416, 2015.

[22] Rosario Gennaro, Craig Gentry, and Bryan Parno. Non-interactive
verifiable computing: Outsourcing computation to untrusted workers. In
Advances in Cryptology - CRYPTO 2010, 30th Annual Cryptology Conference,
Santa Barbara, CA, USA, August 15-19, 2010. Proceedings, pages 465–482,
2010.

[23] Kai-Min Chung, Yael Tauman Kalai, and Salil P. Vadhan. Improved
delegation of computation using fully homomorphic encryption. In
Advances in Cryptology - CRYPTO 2010, 30th Annual Cryptology Conference,
Santa Barbara, CA, USA, August 15-19, 2010. Proceedings, pages 483–501,
2010.

[24] Michael Backes, Dario Fiore, and Raphael M. Reischuk. Verifiable
delegation of computation on outsourced data. In 2013 ACM SIGSAC
Conference on Computer and Communications Security, CCS’13, Berlin,
Germany, November 4-8, 2013, pages 863–874, 2013.

[25] Liang Feng Zhang and Reihaneh Safavi-Naini. Generalized homomorphic
macs with efficient verification. In ASIAPKC’14, Proceedings of the 2nd
ACM Wookshop on ASIA Public-Key Cryptography, June 3, 2014, Kyoto,
Japan, pages 3–12, 2014.

[26] Dario Fiore, Rosario Gennaro, and Valerio Pastro. Efficiently
verifiable computation on encrypted data. In Proceedings of the 2014 ACM
SIGSAC Conference on Computer and Communications Security, Scottsdale,
AZ, USA, November 3-7, 2014, pages 844–855, 2014.
