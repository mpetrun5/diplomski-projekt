pragma solidity 0.6.12;

interface IDepositExecute {
    function deposit(bytes32 resourceID, address depositer, bytes calldata data) external returns (bytes memory);
    function executeProposal(bytes32 resourceID, bytes calldata data) external;
}
