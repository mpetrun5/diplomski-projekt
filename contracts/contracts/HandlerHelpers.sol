// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity 0.6.12;

contract HandlerHelpers {
    address public immutable _bridgeAddress;

    mapping (bytes32 => address) public _resourceIDToTokenContractAddress;
    mapping (address => bytes32) public _tokenContractAddressToResourceID;
    mapping (address => bool) public _contractWhitelist;
    mapping (address => bool) public _burnList;

    modifier onlyBridge() {
        _onlyBridge();
        _;
    }

    constructor(
        address          bridgeAddress
    ) public {
        _bridgeAddress = bridgeAddress;
    }

    function _onlyBridge() private view {
        require(msg.sender == _bridgeAddress, "sender must be bridge contract");
    }

    function setResource(bytes32 resourceID, address contractAddress) external override onlyBridge {
        _setResource(resourceID, contractAddress);
    }

    function _setResource(bytes32 resourceID, address contractAddress) internal {
        _resourceIDToTokenContractAddress[resourceID] = contractAddress;
        _tokenContractAddressToResourceID[contractAddress] = resourceID;

        _contractWhitelist[contractAddress] = true;
    }
}
