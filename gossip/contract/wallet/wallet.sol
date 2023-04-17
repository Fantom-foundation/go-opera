// SPDX-License-Identifier: UNLICENSED

pragma solidity >=0.4.22;

contract Wallet {
    bytes32 passwordHash;

    constructor(bytes32 _passwordHash) payable {
        passwordHash = _passwordHash;
    }

    function transfer(
        bytes calldata password,
        bytes32 newPasswordHash,
        address payable to,
        uint256 amount,
        bytes calldata payload
    ) external {
        bytes32 hash = keccak256(password);
        require(hash == passwordHash);

        passwordHash = newPasswordHash;
        paygas();

        (bool success,) = to.call{value: amount}(payload);
        require(success);
    }
}
