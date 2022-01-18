package evmgaspricer

import (
	"context"
	"math/big"
)

type GasPriceClient interface {
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}

type StaticGasPriceDeterminant struct {
	client GasPriceClient
}

func NewStaticGasPriceDeterminant(client GasPriceClient) *StaticGasPriceDeterminant {
	return &StaticGasPriceDeterminant{client: client}
}

func (gasPricer *StaticGasPriceDeterminant) GasPrice() ([]*big.Int, error) {
	gp, err := gasPricer.client.SuggestGasPrice(context.TODO())
	if err != nil {
		return nil, err
	}

	gasPrices := make([]*big.Int, 1)
	gasPrices[0] = gp
	return gasPrices, nil
}
