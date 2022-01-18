package listener

import (
	"errors"

	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts/bridge"
	"github.com/mpetrun5/diplomski-projekt/relayer/message"
	"github.com/mpetrun5/diplomski-projekt/types"
	"github.com/rs/zerolog/log"

	"github.com/ethereum/go-ethereum/common"
)

type EventHandlers map[common.Address]EventHandlerFunc
type EventHandlerFunc func(sourceID, destId uint8, nonce uint64, resourceID types.ResourceID, calldata, handlerResponse []byte) (*message.Message, error)

type ETHEventHandler struct {
	bridgeContract bridge.BridgeContract
	eventHandlers  EventHandlers
}

func NewETHEventHandler(bridgeContract bridge.BridgeContract) *ETHEventHandler {
	return &ETHEventHandler{
		bridgeContract: bridgeContract,
	}
}

func (e *ETHEventHandler) HandleEvent(sourceID, destID uint8, depositNonce uint64, resourceID types.ResourceID, calldata, handlerResponse []byte) (*message.Message, error) {
	handlerAddr, err := e.bridgeContract.GetHandlerAddressForResourceID(resourceID)
	if err != nil {
		return nil, err
	}

	eventHandler, err := e.matchAddressWithHandlerFunc(handlerAddr)
	if err != nil {
		return nil, err
	}

	return eventHandler(sourceID, destID, depositNonce, resourceID, calldata, handlerResponse)
}

func (e *ETHEventHandler) matchAddressWithHandlerFunc(handlerAddress common.Address) (EventHandlerFunc, error) {
	hf, ok := e.eventHandlers[handlerAddress]
	if !ok {
		return nil, errors.New("no corresponding event handler for this address exists")
	}
	return hf, nil
}

func (e *ETHEventHandler) RegisterEventHandler(handlerAddress string, handler EventHandlerFunc) {
	if handlerAddress == "" {
		return
	}

	if e.eventHandlers == nil {
		e.eventHandlers = make(map[common.Address]EventHandlerFunc)
	}

	log.Info().Msgf("Registered event handler for address %s", handlerAddress)

	e.eventHandlers[common.HexToAddress(handlerAddress)] = handler
}

func Erc20EventHandler(sourceID, destId uint8, nonce uint64, resourceID types.ResourceID, calldata, handlerResponse []byte) (*message.Message, error) {
	if len(calldata) < 84 {
		err := errors.New("invalid calldata length: less than 84 bytes")
		return nil, err
	}

	amount := calldata[:32]
	recipientAddress := calldata[64:]

	return &message.Message{
		Source:       sourceID,
		Destination:  destId,
		DepositNonce: nonce,
		ResourceId:   resourceID,
		Type:         message.FungibleTransfer,
		Payload: []interface{}{
			amount,
			recipientAddress,
		},
	}, nil
}
