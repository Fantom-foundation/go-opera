# nodes internal state diff tool


## Goals:

- simplicity and completeness of tests;

- common check approach for all test cases (including nodes switch on/off);

- debug of consensus;


## Roadmap:

- dedicated interface provides node's internal data structs (disabled by default): inmemory interface for unit-tests, RPC-API for integration tests;

- cause of inconsistencies seeking;

- diff result showing a struct (Flag Table, Clotho Check List and others) and row;

