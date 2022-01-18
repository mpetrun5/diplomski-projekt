pragma solidity 0.6.12;
pragma experimental ABIEncoderV2;

import "./utils/AccessControl.sol";
import "./utils/Pausable.sol";
import "./utils/SafeMath.sol";
import "./utils/SafeCast.sol";
import "./interfaces/IDepositExecute.sol";
import "./interfaces/IERCHandler.sol";
import "./interfaces/IGenericHandler.sol";

contract Bridge is Pausable, AccessControl, SafeMath {
    using SafeCast for *;
    uint256 constant public MAX_RELAYERS = 200;

    uint8   public _domainID;
    uint8   public _relayerThreshold;
    uint128 public _fee;
    uint40  public _expiry;

    enum ProposalStatus {Inactive, Active, Passed, Executed, Cancelled}

    struct Proposal {
        ProposalStatus _status;
        uint200 _yesVotes;s
        uint8   _yesVotesTotal;
        uint40  _proposedBlock;
    }

    mapping(bytes32 => address) public _resourceIDToHandlerAddress;
    mapping(uint72 => mapping(bytes32 => Proposal)) private _proposals;

    event Deposit(
        uint8   destinationDomainID,
        bytes32 resourceID,
        uint64  depositNonce,
        address indexed user,
        bytes data,
        bytes handlerResponse
    );
    event ProposalEvent(
        uint8          originDomainID,
        uint64         depositNonce,
        ProposalStatus status,
        bytes32 dataHash
    );
    event ProposalVote(
        uint8   originDomainID,
        uint64  depositNonce,
        ProposalStatus status,
        bytes32 dataHash
    );

    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");

    modifier onlyRelayers() {
        _onlyRelayers();
        _;
    }

    function _onlyRelayers() private view {
        require(hasRole(RELAYER_ROLE, _msgSender()), "sender doesn't have relayer role");
    }

    function _relayerBit(address relayer) private view returns(uint) {
        return uint(1) << sub(AccessControl.getRoleMemberIndex(RELAYER_ROLE, relayer), 1);
    }

    function _hasVoted(Proposal memory proposal, address relayer) private view returns(bool) {
        return (_relayerBit(relayer) & uint(proposal._yesVotes)) > 0;
    }

    function _msgSender() internal override view returns (address payable) {
        return msg.sender
    }

    constructor (uint8 domainID, address[] memory initialRelayers, uint256 initialRelayerThreshold, uint256 fee, uint256 expiry) public {
        _domainID = domainID;
        _relayerThreshold = initialRelayerThreshold.toUint8();
        _fee = fee.toUint128();
        _expiry = expiry.toUint40();

        _setupRole(DEFAULT_ADMIN_ROLE, _msgSender());

        for (uint256 i; i < initialRelayers.length; i++) {
            grantRole(RELAYER_ROLE, initialRelayers[i]);
        }
    }

    function _hasVotedOnProposal(uint72 destNonce, bytes32 dataHash, address relayer) public view returns(bool) {
        return _hasVoted(_proposals[destNonce][dataHash], relayer);
    }

    function isRelayer(address relayer) external view returns (bool) {
        return hasRole(RELAYER_ROLE, relayer);
    }

    function adminSetResource(address handlerAddress, bytes32 resourceID, address tokenAddress) external onlyAdmin {
        _resourceIDToHandlerAddress[resourceID] = handlerAddress;
        IERCHandler handler = IERCHandler(handlerAddress);
        handler.setResource(resourceID, tokenAddress);
    }

    function getProposal(uint8 originDomainID, uint64 depositNonce, bytes32 dataHash) external view returns (Proposal memory) {
        uint72 nonceAndID = (uint72(depositNonce) << 8) | uint72(originDomainID);
        return _proposals[nonceAndID][dataHash];
    }

    function deposit(uint8 destinationDomainID, bytes32 resourceID, bytes calldata data) external payable {
        require(msg.value == _fee, "Incorrect fee supplied");

        address handler = _resourceIDToHandlerAddress[resourceID];
        require(handler != address(0), "resourceID not mapped to handler");

        uint64 depositNonce = ++_depositCounts[destinationDomainID];
        address sender = _msgSender();

        IDepositExecute depositHandler = IDepositExecute(handler);
        bytes memory handlerResponse = depositHandler.deposit(resourceID, sender, data);

        emit Deposit(destinationDomainID, resourceID, depositNonce, sender, data, handlerResponse);
    }

    function voteProposal(uint8 domainID, uint64 depositNonce, bytes32 resourceID, bytes calldata data) external onlyRelayers {
        address handler = _resourceIDToHandlerAddress[resourceID];
        uint72 nonceAndID = (uint72(depositNonce) << 8) | uint72(domainID);
        bytes32 dataHash = keccak256(abi.encodePacked(handler, data));
        Proposal memory proposal = _proposals[nonceAndID][dataHash];

        require(_resourceIDToHandlerAddress[resourceID] != address(0), "no handler for resourceID");

        if (proposal._status == ProposalStatus.Passed) {
            executeProposal(domainID, depositNonce, data, resourceID, true);
            return;
        }

        address sender = _msgSender();

        require(uint(proposal._status) <= 1, "proposal already executed/cancelled");
        require(!_hasVoted(proposal, sender), "relayer already voted");

        if (proposal._status == ProposalStatus.Inactive) {
            proposal = Proposal({
                _status : ProposalStatus.Active,
                _yesVotes : 0,
                _yesVotesTotal : 0,
                _proposedBlock : uint40(block.number)
            });

            emit ProposalEvent(domainID, depositNonce, ProposalStatus.Active, dataHash);
        } else if (uint40(sub(block.number, proposal._proposedBlock)) > _expiry) {
            proposal._status = ProposalStatus.Cancelled;

            emit ProposalEvent(domainID, depositNonce, ProposalStatus.Cancelled, dataHash);
        }

        if (proposal._status != ProposalStatus.Cancelled) {
            proposal._yesVotes = (proposal._yesVotes | _relayerBit(sender)).toUint200();
            proposal._yesVotesTotal++;

            emit ProposalVote(domainID, depositNonce, proposal._status, dataHash);
            if (proposal._yesVotesTotal >= _relayerThreshold) {
                proposal._status = ProposalStatus.Passed;
                emit ProposalEvent(domainID, depositNonce, ProposalStatus.Passed, dataHash);
            }
        }
        _proposals[nonceAndID][dataHash] = proposal;

        if (proposal._status == ProposalStatus.Passed) {
            executeProposal(domainID, depositNonce, data, resourceID, false);
        }
    }

    function executeProposal(uint8 domainID, uint64 depositNonce, bytes calldata data, bytes32 resourceID, bool revertOnFail) public onlyRelayers {
        address handler = _resourceIDToHandlerAddress[resourceID];
        uint72 nonceAndID = (uint72(depositNonce) << 8) | uint72(domainID);
        bytes32 dataHash = keccak256(abi.encodePacked(handler, data));
        Proposal storage proposal = _proposals[nonceAndID][dataHash];

        require(proposal._status == ProposalStatus.Passed, "Proposal must have Passed status");

        proposal._status = ProposalStatus.Executed;
        IDepositExecute depositHandler = IDepositExecute(handler);

        if (revertOnFail) {
            depositHandler.executeProposal(resourceID, data);
        } else {
            try depositHandler.executeProposal(resourceID, data) {
            } catch (bytes memory lowLevelData) {
                proposal._status = ProposalStatus.Passed;
                emit FailedHandlerExecution(lowLevelData);
                return;
            }
        }

        emit ProposalEvent(domainID, depositNonce, ProposalStatus.Executed, dataHash);
    }
}
