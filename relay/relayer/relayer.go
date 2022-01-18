// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"github.com/mpetrun5/diplomski-projekt/relayer/message"
	"github.com/rs/zerolog/log"
)

type Metrics interface {
	TrackDepositMessage(m *message.Message)
}

type RelayedChain interface {
	PollEvents(stop <-chan struct{}, sysErr chan<- error, eventsChan chan *message.Message)
	Write(message *message.Message) error
	DomainID() uint8
}

func NewRelayer(chains []RelayedChain) *Relayer {
	return &Relayer{relayedChains: chains}
}

type Relayer struct {
	metrics       Metrics
	relayedChains []RelayedChain
	registry      map[uint8]RelayedChain
}

func (r *Relayer) Start(stop <-chan struct{}, sysErr chan error) {
	log.Debug().Msgf("Starting relayer")

	messagesChannel := make(chan *message.Message)
	for _, c := range r.relayedChains {
		log.Debug().Msgf("Starting chain %v", c.DomainID())
		r.addRelayedChain(c)
		go c.PollEvents(stop, sysErr, messagesChannel)
	}

	for {
		select {
		case m := <-messagesChannel:
			go r.route(m)
			continue
		case <-stop:
			return
		}
	}
}

func (r *Relayer) route(m *message.Message) {
	r.metrics.TrackDepositMessage(m)

	destChain, ok := r.registry[m.Destination]
	if !ok {
		log.Error().Msgf("no resolver for destID %v to send message registered", m.Destination)
		return
	}

	log.Debug().Msgf("Sending message %+v to destination %v", m, m.Destination)

	if err := destChain.Write(m); err != nil {
		log.Error().Err(err).Msgf("writing message %+v", m)
		return
	}
}

func (r *Relayer) addRelayedChain(c RelayedChain) {
	if r.registry == nil {
		r.registry = make(map[uint8]RelayedChain)
	}
	domainID := c.DomainID()
	r.registry[domainID] = c
}
