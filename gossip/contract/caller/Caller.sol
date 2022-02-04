// SPDX-License-Identifier: AGPL-3.0-only
pragma solidity ^0.8.4;


interface ICalled {
	function PutVal(uint8 key, uint8 val)
	  external;

	function GetVal(uint8 key)
	  external
	  view
	  returns (uint8 val);
}


contract Caller
{
	ICalled public immutable calledContract;

	constructor(ICalled _called) {
		calledContract = _called;
	}

	function Inc(uint8 _key) external returns (uint8 curr) {
		try calledContract.GetVal(_key) returns (
			uint8 was
		) {
			curr = was + 1;
		}
		catch {
			curr = 1;
		}

		calledContract.PutVal(_key, curr);
	}

}
