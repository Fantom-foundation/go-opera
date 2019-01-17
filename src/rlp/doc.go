/*
Package rlp implements the RLP serialization format.

The purpose of RLP (Recursive Linear Prefix) is to encode arbitrarily
nested arrays of binary data, and RLP is the main encoding method used
to serialize objects. The only purpose of RLP is to encode structure;
encoding specific atomic data types (eg. strings, ints, floats)
is left up to higher-order protocols; integers must be represented in
big endian binary form with no leading zeroes (thus making the integer
value zero equivalent to the empty byte array).

RLP values are distinguished by a type tag. The type tag precedes the
value in the input stream and defines the size and kind of the bytes
that follow.
*/
package rlp
