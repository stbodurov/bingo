// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title HumanRegistry — public commitments for the Singleton proof-of-human network.
/// @notice Each epoch the registry operator anchors the Merkle root of every predicate set
///         (human, human_doc, age_over_16, …). Verifiers only accept presentations against
///         anchored roots, so the operator cannot silently add or remove holders.
///         No personal data — not even public keys — is stored on-chain: only 32-byte roots.
contract HumanRegistry {
    struct SetCommitment {
        bytes32 root;
        uint64 size;
    }

    address public operator;
    address public pendingOperator;
    uint64 public latestEpoch;

    /// epoch => setId (keccak256(setName)) => commitment
    mapping(uint64 => mapping(bytes32 => SetCommitment)) public sets;
    /// epoch => keccak256(concat(sorted revoked public keys))
    mapping(uint64 => bytes32) public revocations;

    event EpochAnchored(uint64 indexed epoch, bytes32[] setIds, bytes32[] roots, uint64[] sizes);
    event RevocationsPublished(uint64 indexed epoch, bytes32 commitment);
    event OperatorTransferStarted(address indexed from, address indexed to);
    event OperatorTransferred(address indexed from, address indexed to);

    error NotOperator();
    error EpochNotIncreasing(uint64 latest, uint64 given);
    error LengthMismatch();

    modifier onlyOperator() {
        if (msg.sender != operator) revert NotOperator();
        _;
    }

    constructor(address initialOperator) {
        operator = initialOperator;
    }

    /// @notice Anchor all set roots for a newly sealed epoch.
    /// @dev Selector: keccak256("anchorEpoch(uint64,bytes32[],bytes32[],uint64[])")[0:4]
    function anchorEpoch(uint64 epoch, bytes32[] calldata setIds, bytes32[] calldata roots, uint64[] calldata sizes) external onlyOperator {
        if (epoch <= latestEpoch) revert EpochNotIncreasing(latestEpoch, epoch);
        if (setIds.length != roots.length || roots.length != sizes.length) revert LengthMismatch();
        for (uint256 i = 0; i < setIds.length; i++) {
            sets[epoch][setIds[i]] = SetCommitment({root: roots[i], size: sizes[i]});
        }
        latestEpoch = epoch;
        emit EpochAnchored(epoch, setIds, roots, sizes);
    }

    /// @notice Commit to the revocation list as of an epoch.
    function publishRevocations(uint64 epoch, bytes32 commitment) external onlyOperator {
        revocations[epoch] = commitment;
        emit RevocationsPublished(epoch, commitment);
    }

    /// @notice Convenience view for verifiers: is `root` the anchored root of `setName` at `epoch`?
    function isAnchored(uint64 epoch, string calldata setName, bytes32 root) external view returns (bool) {
        return sets[epoch][keccak256(bytes(setName))].root == root;
    }

    /// @notice Two-step operator rotation (e.g. to a threshold-signature multisig).
    function transferOperator(address next) external onlyOperator {
        pendingOperator = next;
        emit OperatorTransferStarted(operator, next);
    }

    function acceptOperator() external {
        if (msg.sender != pendingOperator) revert NotOperator();
        emit OperatorTransferred(operator, pendingOperator);
        operator = pendingOperator;
        pendingOperator = address(0);
    }
}
