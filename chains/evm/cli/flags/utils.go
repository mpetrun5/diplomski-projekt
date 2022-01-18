package flags

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls"

	"github.com/mpetrun5/diplomski-projekt/types"

	"github.com/mpetrun5/diplomski-projekt/crypto/secp256k1"
	"github.com/spf13/cobra"
)

var DefaultGasLimit = uint64(200000)

func GlobalFlagValues(cmd *cobra.Command) (string, uint64, *big.Int, *secp256k1.Keypair, error) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return "", DefaultGasLimit, nil, nil, err
	}

	gasLimitInt, err := cmd.Flags().GetUint64("gas-limit")
	if err != nil {
		return "", DefaultGasLimit, nil, nil, err
	}

	gasPriceInt, err := cmd.Flags().GetUint64("gas-price")
	if err != nil {
		return "", DefaultGasLimit, nil, nil, err
	}
	var gasPrice *big.Int = nil
	if gasPriceInt != 0 {
		gasPrice = big.NewInt(0).SetUint64(gasPriceInt)
	}

	senderKeyPair, err := defineSender(cmd)
	if err != nil {
		return "", DefaultGasLimit, nil, nil, err
	}
	return url, gasLimitInt, gasPrice, senderKeyPair, nil
}

func defineSender(cmd *cobra.Command) (*secp256k1.Keypair, error) {
	privateKey, err := cmd.Flags().GetString("private-key")
	if err != nil {
		return nil, err
	}
	if privateKey != "" {
		kp, err := secp256k1.NewKeypairFromString(privateKey)
		if err != nil {
			return nil, err
		}
		return kp, nil
	}

	return nil, nil
}

func ProcessResourceID(resourceID string) (types.ResourceID, error) {
	if resourceID[0:2] == "0x" {
		resourceID = resourceID[2:]
	}
	resourceIdBytes, err := hex.DecodeString(resourceID)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed decoding resourceID hex string: %s", err)
	}
	return calls.SliceTo32Bytes(resourceIdBytes), nil
}

func MarkFlagsAsRequired(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
