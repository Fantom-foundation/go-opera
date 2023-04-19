// SPDX-License-Identifier: UNLICENSED

pragma solidity >=0.8.0;
pragma experimental AccountAbstraction;

account contract AccountAbstractionTester {
    uint256 public nonce;
    uint256 public gasPrice;
    uint256 public gasLeft;
    address public sender;
    address public origin;
    uint256 public balance;

    constructor() payable {}

    function revertBeforePaygas() external {
        nonce = tx.nonce;
        revert();
    }

    function setNonceBeforePaygasAndRevert() external {
        nonce = tx.nonce;
        paygas();
        revert();
    }

    function gasLeftBeforePaygas() external {
        gasLeft = gasleft();
        paygas();
    }

    function gasLeftAfterPaygas() external {
        paygas();
        gasLeft = gasleft();
    }

    function setSenderBeforePaygas() external {
        sender = msg.sender;
        paygas();
    }

    function setSenderAfterPaygas() external {
        paygas();
        sender = msg.sender;
    }

    function setOriginBeforePaygas() external {
        origin = tx.origin;
        paygas();
    }

    function setOriginAfterPaygas() external {
        paygas();
        origin = tx.origin;
    }

    function setNonce() external {
        nonce = tx.nonce;
        paygas();
    }

    function setBalanceBeforePaygas() external {
        balance = address(this).balance;
        paygas();
    }

    function setBalanceAfterPaygas() external {
        paygas();
        balance = address(this).balance;
    }

    function setGasPrice() external {
        paygas();
        gasPrice = tx.gasprice;
    }

    function call(address _contract, string calldata _method) external {
        paygas();
        (bool success,) = _contract.call(abi.encodeWithSignature(_method));
        require(success);
    }

    function reset() external {
        paygas();
        nonce = 0;
        gasPrice = 0;
        gasLeft = 0;
        sender = address(0);
        origin = address(0);
        balance = 0;
    }
}
