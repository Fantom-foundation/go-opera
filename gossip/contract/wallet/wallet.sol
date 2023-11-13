// SPDX-License-Identifier: UNLICENSED

pragma solidity >=0.4.22;
pragma experimental AccountAbstraction;

account contract Wallet {
    address public owner;

    struct Signature {
        uint8 v;
        bytes32 r;
        bytes32 s;
    }

    constructor(address _owner) payable {
        owner = _owner;
    }

    modifier unlock(Signature calldata signature) {
        bytes memory hash = abi.encodePacked(
                this,
                tx.nonce
        );
        bytes32 digest = keccak256(hash);
        address signer = recover(digest, signature);
        require(owner == signer, "Not owner");
        paygas();
        _;
    }

    function transfer(
        address payable to,
        uint256 amount,
        bytes calldata payload,
        Signature calldata signature
    ) external unlock(signature) {
        (bool success,) = to.call{value: amount}(payload);
        require(success);
    }

    function changeOwner(
        address newOwner,
        Signature calldata signature
    ) external unlock(signature) {
        owner = newOwner;
    }

    function recover(
        bytes32 digest,
        Signature calldata signature
    ) private pure returns (address) {
        return ecrecover(
            digest,
            signature.v,
            signature.r,
            signature.s
        );
    }
}
