// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/relayer/message"
	"github.com/mpetrun5/diplomski-projekt/store"
	"github.com/mpetrun5/diplomski-projekt/types"

	"github.com/rs/zerolog/log"
)

var (
	blockRetryInterval = 10 * time.Second
	blockDelay         = big.NewInt(3)
)

type EventHandler interface {
	HandleEvent(sourceID, destID uint8, nonce uint64, resourceID types.ResourceID, calldata, handlerResponse []byte) (*message.Message, error)
}
type ChainClient interface {
	LatestBlock() (*big.Int, error)
	FetchDepositLogs(ctx context.Context, address common.Address, startBlock *big.Int, endBlock *big.Int) ([]*evmclient.DepositLogs, error)
	CallContract(ctx context.Context, callArgs map[string]interface{}, blockNumber *big.Int) ([]byte, error)
}

type EVMListener struct {
	chainReader   ChainClient
	eventHandler  EventHandler
	bridgeAddress common.Address
}

func NewEVMListener(chainReader ChainClient, handler EventHandler, bridgeAddress common.Address) *EVMListener {
	return &EVMListener{chainReader: chainReader, eventHandler: handler, bridgeAddress: bridgeAddress}
}

func (l *EVMListener) ListenToEvents(
	startBlock *big.Int,
	domainID uint8,
	blockstore *store.BlockStore,
	stopChn <-chan struct{},
	errChn chan<- error,
) <-chan *message.Message {
	ch := make(chan *message.Message)
	go func() {
		for {
			select {
			case <-stopChn:
				return
			default:
				head, err := l.chainReader.LatestBlock()
				if err != nil {
					time.Sleep(blockRetryInterval)
					continue
				}
				if startBlock == nil {
					startBlock = head
				}
				if big.NewInt(0).Sub(head, startBlock).Cmp(blockDelay) == -1 {
					time.Sleep(blockRetryInterval)
					continue
				}
				logs, err := l.chainReader.FetchDepositLogs(context.Background(), l.bridgeAddress, startBlock, startBlock)
				if err != nil {
					continue
				}
				for _, eventLog := range logs {
					log.Debug().Msgf("Deposit log found from sender: %s in block: %s with  destinationDomainId: %v, resourceID: %s, depositNonce: %v", eventLog.SenderAddress, startBlock.String(), eventLog.DestinationDomainID, eventLog.ResourceID, eventLog.DepositNonce)
					m, err := l.eventHandler.HandleEvent(domainID, eventLog.DestinationDomainID, eventLog.DepositNonce, eventLog.ResourceID, eventLog.Data, eventLog.HandlerResponse)
					if err != nil {
						continue
					} else {
						log.Debug().Msgf("Resolved message %+v in block %s", m, startBlock.String())
						ch <- m
					}
				}
				err = blockstore.StoreBlock(startBlock, domainID)
				if err != nil {
					log.Error().Str("block", startBlock.String()).Err(err).Msg("Failed to write latest block to blockstore")
				}
				startBlock.Add(startBlock, big.NewInt(1))
			}
		}
	}()
	return ch
}
