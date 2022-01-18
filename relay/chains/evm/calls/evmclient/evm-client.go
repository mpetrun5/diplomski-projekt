package evmclient

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/consts"
	"github.com/mpetrun5/diplomski-projekt/crypto/secp256k1"
	"github.com/mpetrun5/diplomski-projekt/util"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	bridgeTypes "github.com/mpetrun5/diplomski-projekt/types"
	"github.com/rs/zerolog/log"
)

type EVMClient struct {
	*ethclient.Client
	kp        *secp256k1.Keypair
	rpClient  *rpc.Client
	nonce     *big.Int
	nonceLock sync.Mutex
}

type DepositLogs struct {
	DestinationDomainID uint8
	ResourceID          bridgeTypes.ResourceID
	DepositNonce        uint64
	SenderAddress       common.Address
	Data                []byte
	HandlerResponse     []byte
}

type CommonTransaction interface {
	Hash() common.Hash
	RawWithSignature(key *ecdsa.PrivateKey, domainID *big.Int) ([]byte, error)
}

func NewEVMClientFromParams(url string, privateKey *ecdsa.PrivateKey) (*EVMClient, error) {
	rpcClient, err := rpc.DialContext(context.TODO(), url)
	if err != nil {
		return nil, err
	}
	c := &EVMClient{}
	c.Client = ethclient.NewClient(rpcClient)
	c.rpClient = rpcClient
	c.kp = secp256k1.NewKeypair(*privateKey)
	return c, nil
}

// LatestBlock returns the latest block from the current chain
func (c *EVMClient) LatestBlock() (*big.Int, error) {
	var head *headerNumber
	err := c.rpClient.CallContext(context.Background(), &head, "eth_getBlockByNumber", toBlockNumArg(nil), false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}
	if err != nil {
		return nil, err
	}
	return head.Number, nil
}

type headerNumber struct {
	Number *big.Int `json:"number"           gencodec:"required"`
}

func (h *headerNumber) UnmarshalJSON(input []byte) error {
	type headerNumber struct {
		Number *hexutil.Big `json:"number" gencodec:"required"`
	}
	var dec headerNumber
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Number == nil {
		return errors.New("missing required field 'number' for Header")
	}
	h.Number = (*big.Int)(dec.Number)
	return nil
}

func (c *EVMClient) WaitAndReturnTxReceipt(h common.Hash) (*types.Receipt, error) {
	retry := 50
	for retry > 0 {
		receipt, err := c.Client.TransactionReceipt(context.Background(), h)
		if err != nil {
			retry--
			time.Sleep(5 * time.Second)
			continue
		}
		if receipt.Status != 1 {
			return receipt, fmt.Errorf("transaction failed on chain. Receipt status %v", receipt.Status)
		}
		return receipt, nil
	}
	return nil, errors.New("tx did not appear")
}

func (c *EVMClient) GetTransactionByHash(h common.Hash) (tx *types.Transaction, isPending bool, err error) {
	return c.Client.TransactionByHash(context.Background(), h)
}

func (c *EVMClient) FetchDepositLogs(ctx context.Context, contractAddress common.Address, startBlock *big.Int, endBlock *big.Int) ([]*DepositLogs, error) {
	logs, err := c.FilterLogs(ctx, buildQuery(contractAddress, string(util.Deposit), startBlock, endBlock))
	if err != nil {
		return nil, err
	}
	depositLogs := make([]*DepositLogs, 0)

	abi, err := abi.JSON(strings.NewReader(consts.BridgeABI))
	if err != nil {
		return nil, err
	}

	for _, l := range logs {
		dl, err := c.UnpackDepositEventLog(abi, l.Data)
		if err != nil {
			log.Error().Msgf("failed unpacking deposit event log: %v", err)
			continue
		}
		log.Debug().Msgf("Found deposit log in block: %d, TxHash: %s, contractAddress: %s, sender: %s", l.BlockNumber, l.TxHash, l.Address, dl.SenderAddress)

		depositLogs = append(depositLogs, dl)
	}

	return depositLogs, nil
}

func (c *EVMClient) UnpackDepositEventLog(abi abi.ABI, data []byte) (*DepositLogs, error) {
	var dl DepositLogs

	err := abi.UnpackIntoInterface(&dl, "Deposit", data)
	if err != nil {
		return &DepositLogs{}, err
	}

	return &dl, nil
}

func (c *EVMClient) FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error) {
	return c.FilterLogs(ctx, buildQuery(contractAddress, event, startBlock, endBlock))
}

func (c *EVMClient) SendRawTransaction(ctx context.Context, tx []byte) error {
	return c.rpClient.CallContext(ctx, nil, "eth_sendRawTransaction", hexutil.Encode(tx))
}

func (c *EVMClient) CallContract(ctx context.Context, callArgs map[string]interface{}, blockNumber *big.Int) ([]byte, error) {
	var hex hexutil.Bytes
	err := c.rpClient.CallContext(ctx, &hex, "eth_call", callArgs, toBlockNumArg(blockNumber))
	if err != nil {
		return nil, err
	}
	return hex, nil
}

func (c *EVMClient) CallContext(ctx context.Context, target interface{}, rpcMethod string, args ...interface{}) error {
	err := c.rpClient.CallContext(ctx, target, rpcMethod, args...)
	if err != nil {
		return err
	}
	return nil
}

func (c *EVMClient) PendingCallContract(ctx context.Context, callArgs map[string]interface{}) ([]byte, error) {
	var hex hexutil.Bytes
	err := c.rpClient.CallContext(ctx, &hex, "eth_call", callArgs, "pending")
	if err != nil {
		return nil, err
	}
	return hex, nil
}

func (c *EVMClient) From() common.Address {
	return c.kp.CommonAddress()
}

func (c *EVMClient) SignAndSendTransaction(ctx context.Context, tx CommonTransaction) (common.Hash, error) {
	id, err := c.ChainID(ctx)
	if err != nil {
		//panic(err)
		// Probably chain does not support chainID eg. CELO
		id = nil
	}
	rawTx, err := tx.RawWithSignature(c.kp.PrivateKey(), id)
	if err != nil {
		return common.Hash{}, err
	}
	err = c.SendRawTransaction(ctx, rawTx)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

func (c *EVMClient) RelayerAddress() common.Address {
	return c.kp.CommonAddress()
}

func (c *EVMClient) LockNonce() {
	c.nonceLock.Lock()
}

func (c *EVMClient) UnlockNonce() {
	c.nonceLock.Unlock()
}

func (c *EVMClient) UnsafeNonce() (*big.Int, error) {
	var err error
	for i := 0; i <= 10; i++ {
		if c.nonce == nil {
			nonce, err := c.PendingNonceAt(context.Background(), c.kp.CommonAddress())
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			c.nonce = big.NewInt(0).SetUint64(nonce)
			return c.nonce, nil
		}
		return c.nonce, nil
	}
	return nil, err
}

func (c *EVMClient) UnsafeIncreaseNonce() error {
	nonce, err := c.UnsafeNonce()
	if err != nil {
		return err
	}
	c.nonce = nonce.Add(nonce, big.NewInt(1))
	return nil
}

func (c *EVMClient) BaseFee() (*big.Int, error) {
	head, err := c.HeaderByNumber(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return head.BaseFee, nil
}

func (c *EVMClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1000000000000), nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

func buildQuery(contract common.Address, sig string, startBlock *big.Int, endBlock *big.Int) ethereum.FilterQuery {
	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{contract},
		Topics: [][]common.Hash{
			{crypto.Keccak256Hash([]byte(sig))},
		},
	}
	return query
}
