package erc20

import (
	"math/big"

	"github.com/mpetrun5/diplomski-projekt/crypto/secp256k1"

	"github.com/ethereum/go-ethereum/common"
)

//flag vars
var (
	Amount         string
	Decimals       uint64
	DstAddress     string
	Erc20Address   string
	Recipient      string
	Bridge         string
	DomainID       uint8
	ResourceID     string
	AccountAddress string
	OwnerAddress   string
	SpenderAddress string
	Minter         string
)

//processed flag vars
var (
	RecipientAddress   common.Address
	RealAmount         *big.Int
	Erc20Addr          common.Address
	MinterAddr         common.Address
	BridgeAddr         common.Address
	ResourceIdBytesArr [32]byte
)

// global flags
var (
	dstAddress    common.Address
	url           string
	gasLimit      uint64
	gasPrice      *big.Int
	senderKeyPair *secp256k1.Keypair
)
