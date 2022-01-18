// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package evm

import (
	"fmt"
	"math/big"

	"github.com/mpetrun5/diplomski-projekt/config/chain"
	"github.com/mpetrun5/diplomski-projekt/relayer/message"
	"github.com/mpetrun5/diplomski-projekt/store"
	"github.com/rs/zerolog/log"
)

type EventListener interface {
	ListenToEvents(startBlock *big.Int, domainID uint8, blockstore *store.BlockStore, stopChn <-chan struct{}, errChn chan<- error) <-chan *message.Message
}

type ProposalVoter interface {
	VoteProposal(message *message.Message) error
}

type EVMChain struct {
	listener   EventListener
	writer     ProposalVoter
	blockstore *store.BlockStore
	config     *chain.EVMConfig
}

func NewEVMChain(listener EventListener, writer ProposalVoter, blockstore *store.BlockStore, config *chain.EVMConfig) *EVMChain {
	return &EVMChain{listener: listener, writer: writer, blockstore: blockstore, config: config}
}

func (c *EVMChain) PollEvents(stop <-chan struct{}, sysErr chan<- error, eventsChan chan *message.Message) {
	log.Info().Msg("Polling Blocks...")

	startBlock, err := c.blockstore.GetStartBlock(
		*c.config.GeneralChainConfig.Id,
		c.config.StartBlock,
		c.config.GeneralChainConfig.LatestBlock,
		c.config.GeneralChainConfig.FreshStart,
	)
	if err != nil {
		sysErr <- fmt.Errorf("error %w on getting last stored block", err)
		return
	}

	ech := c.listener.ListenToEvents(startBlock, *c.config.GeneralChainConfig.Id, c.blockstore, stop, sysErr)
	for {
		select {
		case <-stop:
			return
		case newEvent := <-ech:
			eventsChan <- newEvent
			continue
		}
	}
}

func (c *EVMChain) Write(msg *message.Message) error {
	return c.writer.VoteProposal(msg)
}

func (c *EVMChain) DomainID() uint8 {
	return *c.config.GeneralChainConfig.Id
}
