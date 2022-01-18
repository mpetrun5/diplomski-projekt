// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"math/big"

	"github.com/mpetrun5/diplomski-projekt/types"
)

type TransferType string

const (
	FungibleTransfer    TransferType = "FungibleTransfer"
	NonFungibleTransfer TransferType = "NonFungibleTransfer"
	GenericTransfer     TransferType = "GenericTransfer"
)

type ProposalStatus struct {
	Status        uint8
	YesVotes      *big.Int
	YesVotesTotal uint8
	ProposedBlock *big.Int
}

const (
	ProposalStatusInactive uint8 = iota
	ProposalStatusActive
	ProposalStatusPassed
	ProposalStatusExecuted
	ProposalStatusCanceled
)

type Message struct {
	Source       uint8
	Destination  uint8
	DepositNonce uint64
	ResourceId   types.ResourceID
	Payload      []interface{}
	Type         TransferType
}
