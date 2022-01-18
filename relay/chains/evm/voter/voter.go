// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package voter

import (
	"fmt"
	"time"

	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/voter/proposal"
	"github.com/mpetrun5/diplomski-projekt/relayer/message"
	"github.com/rs/zerolog/log"
)

const (
	maxSimulateVoteChecks = 5
	maxShouldVoteChecks   = 40
	shouldVoteCheckPeriod = 15
)

var (
	Sleep = time.Sleep
)

type MessageHandler interface {
	HandleMessage(m *message.Message) (*proposal.Proposal, error)
}

type BridgeContract interface {
	VoteProposal(proposal *proposal.Proposal, opts transactor.TransactOptions) (*common.Hash, error)
}

type EVMVoter struct {
	mh                   MessageHandler
	bridgeContract       BridgeContract
	pendingProposalVotes map[common.Hash]uint8
}

func NewVoter(mh MessageHandler, bridgeContract BridgeContract) *EVMVoter {
	return &EVMVoter{
		mh:                   mh,
		bridgeContract:       bridgeContract,
		pendingProposalVotes: make(map[common.Hash]uint8),
	}
}

func (v *EVMVoter) VoteProposal(m *message.Message) error {
	prop, err := v.mh.HandleMessage(m)
	if err != nil {
		return err
	}

	hash, err := v.bridgeContract.VoteProposal(prop, transactor.TransactOptions{})
	if err != nil {
		return fmt.Errorf("voting failed. Err: %w", err)
	}

	log.Debug().Str("hash", hash.String()).Uint64("nonce", prop.DepositNonce).Msgf("Voted")
	return nil
}
