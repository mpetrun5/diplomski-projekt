package erc20

import (
	"math/big"
	"strings"

	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/consts"
	"github.com/rs/zerolog/log"
)

type ERC20Contract struct {
	contracts.Contract
}

func NewERC20Contract(
	client calls.ContractCallerDispatcher,
	erc20ContractAddress common.Address,
	transactor transactor.Transactor,
) *ERC20Contract {
	a, _ := abi.JSON(strings.NewReader(consts.ERC20PresetMinterPauserABI))
	b := common.FromHex(consts.ERC20PresetMinterPauserBin)
	return &ERC20Contract{contracts.NewContract(erc20ContractAddress, a, b, client, transactor)}
}

func (c *ERC20Contract) ApproveTokens(
	target common.Address,
	amount *big.Int,
	opts transactor.TransactOptions,
) (*common.Hash, error) {
	log.Debug().Msgf("Approving %s tokens for %s", target.String(), amount.String())
	return c.ExecuteTransaction("approve", opts, target, amount)
}
