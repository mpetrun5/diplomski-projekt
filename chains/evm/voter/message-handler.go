package voter

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts/bridge"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/voter/proposal"
	"github.com/mpetrun5/diplomski-projekt/relayer/message"
	"github.com/rs/zerolog/log"
)

type MessageHandlerFunc func(m *message.Message, handlerAddr, bridgeAddress common.Address) (*proposal.Proposal, error)

func NewEVMMessageHandler(bridgeContract bridge.BridgeContract) *EVMMessageHandler {
	return &EVMMessageHandler{
		bridgeContract: bridgeContract,
	}
}

type EVMMessageHandler struct {
	bridgeContract bridge.BridgeContract
	handlers       map[common.Address]MessageHandlerFunc
}

func (mh *EVMMessageHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	addr, err := mh.bridgeContract.GetHandlerAddressForResourceID(m.ResourceId)
	if err != nil {
		return nil, err
	}
	handleMessage, err := mh.MatchAddressWithHandlerFunc(addr)
	if err != nil {
		return nil, err
	}
	prop, err := handleMessage(m, addr, *mh.bridgeContract.ContractAddress())
	if err != nil {
		return nil, err
	}
	return prop, nil
}

func (mh *EVMMessageHandler) MatchAddressWithHandlerFunc(addr common.Address) (MessageHandlerFunc, error) {
	h, ok := mh.handlers[addr]
	if !ok {
		return nil, fmt.Errorf("no corresponding message handler for this address %s exists", addr.Hex())
	}
	return h, nil
}

func (mh *EVMMessageHandler) RegisterMessageHandler(address string, handler MessageHandlerFunc) {
	if address == "" {
		return
	}
	if mh.handlers == nil {
		mh.handlers = make(map[common.Address]MessageHandlerFunc)
	}

	log.Info().Msgf("Registered message handler for address %s", address)

	mh.handlers[common.HexToAddress(address)] = handler
}

func ERC20MessageHandler(m *message.Message, handlerAddr, bridgeAddress common.Address) (*proposal.Proposal, error) {
	amount, ok := m.Payload[0].([]byte)
	if !ok {
		return nil, errors.New("wrong payloads amount format")
	}
	recipient, ok := m.Payload[1].([]byte)
	if !ok {
		return nil, errors.New("wrong payloads recipient format")
	}
	var data []byte
	data = append(data, common.LeftPadBytes(amount, 32)...)
	recipientLen := big.NewInt(int64(len(recipient))).Bytes()
	data = append(data, common.LeftPadBytes(recipientLen, 32)...)
	data = append(data, recipient...)
	return proposal.NewProposal(m.Source, m.DepositNonce, m.ResourceId, data, handlerAddr, bridgeAddress), nil
}
