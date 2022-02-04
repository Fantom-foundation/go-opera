// SPDX-License-Identifier: AGPL-3.0-only
pragma solidity ^0.8.4;


contract Called
{

	constructor() {
	}

	mapping(uint8 => uint8) private mem;

	function PutVal(uint8 _key, uint8 _val) external {
		require(_val > 0, 'Unexpected value');
		mem[_key] = _val;
	}

	function GetVal(uint8 _key) external view returns (uint8 val) {
		val = mem[_key];
		require(val > 0, 'Undefined value');
	}

}
