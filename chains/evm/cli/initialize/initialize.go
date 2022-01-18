package initialize

import (
	"math/big"

	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmclient"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmgaspricer"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor/signAndSend"
	"github.com/mpetrun5/diplomski-projekt/crypto/secp256k1"
)

func InitializeClient(
	url string,
	senderKeyPair *secp256k1.Keypair,
) (*evmclient.EVMClient, error) {
	ethClient, err := evmclient.NewEVMClientFromParams(
		url, senderKeyPair.PrivateKey())
	if err != nil {
		return nil, err
	}
	return ethClient, nil
}

func InitializeTransactor(
	gasPrice *big.Int,
	txFabric calls.TxFabric,
	client *evmclient.EVMClient,
) (transactor.Transactor, error) {
	gasPricer := evmgaspricer.NewStaticGasPriceDeterminant(
		client,
		&evmgaspricer.GasPricerOpts{UpperLimitFeePerGas: gasPrice},
	)

	trans := signAndSend.NewSignAndSendTransactor(txFabric, gasPricer, client)
	return trans, nil
}
