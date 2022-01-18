package bridge

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/consts"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts/deposit"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/voter/proposal"
	"github.com/rs/zerolog/log"
)

type BridgeContract struct {
	contracts.Contract
}

func NewBridgeContract(
	client calls.ContractCallerDispatcher,
	bridgeContractAddress common.Address,
	transactor transactor.Transactor,
) *BridgeContract {
	a, _ := abi.JSON(strings.NewReader(consts.BridgeABI))
	b := common.FromHex(consts.BridgeBin)
	return &BridgeContract{contracts.NewContract(bridgeContractAddress, a, b, client, transactor)}
}

func (c *BridgeContract) AdminSetResource(
	handlerAddr common.Address,
	rID [32]byte,
	targetContractAddr common.Address,
	opts transactor.TransactOptions,
) (*common.Hash, error) {
	log.Debug().Msgf("Setting resource %s", hexutil.Encode(rID[:]))
	return c.ExecuteTransaction(
		"adminSetResource",
		opts,
		handlerAddr, rID, targetContractAddr,
	)
}

func (c *BridgeContract) deposit(
	resourceID [32]byte,
	destDomainID uint8,
	data []byte,
	opts transactor.TransactOptions,
) (*common.Hash, error) {
	return c.ExecuteTransaction(
		"deposit",
		opts,
		destDomainID, resourceID, data,
	)
}

func (c *BridgeContract) Erc20Deposit(
	recipient common.Address,
	amount *big.Int,
	resourceID [32]byte,
	destDomainID uint8,
	opts transactor.TransactOptions,
) (*common.Hash, error) {
	data := deposit.ConstructErc20DepositData(recipient.Bytes(), amount)
	txHash, err := c.deposit(resourceID, destDomainID, data, opts)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}
	return txHash, err
}

func (c *BridgeContract) VoteProposal(
	proposal *proposal.Proposal,
	opts transactor.TransactOptions,
) (*common.Hash, error) {
	return c.ExecuteTransaction(
		"voteProposal",
		opts,
		proposal.Source, proposal.DepositNonce, proposal.ResourceId, proposal.Data,
	)
}

func (c *BridgeContract) GetHandlerAddressForResourceID(
	resourceID [32]byte,
) (common.Address, error) {
	res, err := c.CallContract("_resourceIDToHandlerAddress", resourceID)
	if err != nil {
		return common.Address{}, err
	}
	out := *abi.ConvertType(res[0], new(common.Address)).(*common.Address)
	return out, nil
}
