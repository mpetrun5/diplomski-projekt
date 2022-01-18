package erc20

import (
	"errors"
	"fmt"
	"math/big"

	callsUtil "github.com/mpetrun5/diplomski-projekt/chains/evm/calls"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts/erc20"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmtransaction"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/cli/initialize"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/cli/flags"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve an ERC20 tokens",
	Long:  "Approve an ERC20 tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := initialize.InitializeClient(url, senderKeyPair)
		if err != nil {
			return err
		}
		t, err := initialize.InitializeTransactor(gasPrice, evmtransaction.NewTransaction, c)
		if err != nil {
			return err
		}
		return ApproveCmd(cmd, args, erc20.NewERC20Contract(c, Erc20Addr, t))
	},
	Args: func(cmd *cobra.Command, args []string) error {
		err := ValidateApproveFlags(cmd, args)
		if err != nil {
			return err
		}

		err = ProcessApproveFlags(cmd, args)
		return err
	},
}

func BindApproveFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&Erc20Address, "contract", "", "ERC20 contract address")
	cmd.Flags().StringVar(&Amount, "amount", "", "Amount to grant allowance")
	cmd.Flags().StringVar(&Recipient, "recipient", "", "Recipient address")
	cmd.Flags().Uint64Var(&Decimals, "decimals", 0, "ERC20 token decimals")
	flags.MarkFlagsAsRequired(cmd, "contract", "amount", "recipient", "decimals")
}

func init() {
	BindApproveFlags(approveCmd)
}

func ValidateApproveFlags(cmd *cobra.Command, args []string) error {
	if !common.IsHexAddress(Erc20Address) {
		return errors.New("invalid erc20Address address")
	}
	if !common.IsHexAddress(Recipient) {
		return errors.New("invalid minter address")
	}
	return nil
}

func ProcessApproveFlags(cmd *cobra.Command, args []string) error {
	var err error

	decimals := big.NewInt(int64(Decimals))
	Erc20Addr = common.HexToAddress(Erc20Address)
	RecipientAddress = common.HexToAddress(Recipient)
	RealAmount, err = callsUtil.UserAmountToWei(Amount, decimals)
	if err != nil {
		return err
	}

	return nil
}

func ApproveCmd(cmd *cobra.Command, args []string, contract *erc20.ERC20Contract) error {
	_, err := contract.ApproveTokens(RecipientAddress, RealAmount, transactor.TransactOptions{GasLimit: gasLimit})
	if err != nil {
		log.Fatal().Err(err)
		return err
	}

	fmt.Printf(
		"%s account granted allowance on %v tokens of %s",
		RecipientAddress.String(), Amount, RecipientAddress.String(),
	)
	return nil
}
