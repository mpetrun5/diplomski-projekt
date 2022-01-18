package calls

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmclient"
)

type TxFabric func(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrices []*big.Int, data []byte) (evmclient.CommonTransaction, error)

type ContractChecker interface {
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
}

type ContractCaller interface {
	CallContract(ctx context.Context, callArgs map[string]interface{}, blockNumber *big.Int) ([]byte, error)
}

type GasPricer interface {
	GasPrice() ([]*big.Int, error)
}

type ClientDispatcher interface {
	WaitAndReturnTxReceipt(h common.Hash) (*types.Receipt, error)
	SignAndSendTransaction(ctx context.Context, tx evmclient.CommonTransaction) (common.Hash, error)
	GetTransactionByHash(h common.Hash) (tx *types.Transaction, isPending bool, err error)
	UnsafeNonce() (*big.Int, error)
	LockNonce()
	UnlockNonce()
	UnsafeIncreaseNonce() error
	From() common.Address
}

type ContractCallerDispatcher interface {
	ContractCaller
	ClientDispatcher
	ContractChecker
}
