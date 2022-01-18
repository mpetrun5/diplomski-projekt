package bridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/crypto/secp256k1"
)

//flag vars
var (
	Bridge     string
	Handler    string
	ResourceID string
	Target     string
)

//processed flag vars
var (
	BridgeAddr         common.Address
	ResourceIdBytesArr [32]byte
	HandlerAddr        common.Address
	TargetContractAddr common.Address
)

// global flags
var (
	url           string
	gasLimit      uint64
	gasPrice      *big.Int
	senderKeyPair *secp256k1.Keypair
)
