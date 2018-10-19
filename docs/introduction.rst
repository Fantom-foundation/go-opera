.. _introduction:

Introduction
============

What is Lachesis?
---------------

Lachesis is an open-source software component intended for developers who want to 
build peer-to-peer (p2p) applications, mobile or other, without having to 
implement their own p2p networking layer from scratch. Under the hood, it 
enables many computers to behave as one; a technique known as state machine 
replication. 

Lachesis is designed to easily plug into applications written in any programming 
language. Developers can focus on building the application logic and simply 
integrate with Lachesis to handle the replication aspect. Basically, Lachesis will 
connect to other Lachesis nodes and guarantee that everyone processes the same 
commands in the same order. To do this, it uses p2p networking and a Byzantine 
Fault Tolerant (BFT) consensus algorithm.

.. image:: assets/lachesis_network.png
   :height: 453px
   :width: 640px
   :align: left

Lachesis is:

- **Asynchronous**: 
    Participants have the freedom to process commands at different times.
- **Leaderless**: 
    No participant plays a 'special' role.
- **Byzantine Fault-Tolerant**: 
    Supports one third of faulty nodes, including malicious behavior.
- **Final**: 
    Lachesis's output can be used immediately, no need for block confirmations, 
    etc.