package bridge

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts/bridge"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmtransaction"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/cli/flags"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/cli/initialize"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var registerResourceCmd = &cobra.Command{
	Use:   "register-resource",
	Short: "Register a resource ID",
	Long:  "Register a resource ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := initialize.InitializeClient(url, senderKeyPair)
		if err != nil {
			return err
		}
		t, err := initialize.InitializeTransactor(gasPrice, evmtransaction.NewTransaction, c)
		if err != nil {
			return err
		}
		return RegisterResourceCmd(cmd, args, bridge.NewBridgeContract(c, BridgeAddr, t))
	},
	Args: func(cmd *cobra.Command, args []string) error {
		err := ValidateRegisterResourceFlags(cmd, args)
		if err != nil {
			return err
		}

		err = ProcessRegisterResourceFlags(cmd, args)
		return err
	},
}

func BindRegisterResourceFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&Handler, "handler", "", "Handler contract address")
	cmd.Flags().StringVar(&Bridge, "bridge", "", "Bridge contract address")
	cmd.Flags().StringVar(&Target, "target", "", "Contract address to be registered")
	cmd.Flags().StringVar(&ResourceID, "resource", "", "Resource ID to be registered")
	flags.MarkFlagsAsRequired(cmd, "handler", "bridge", "target", "resource")
}

func init() {
	BindRegisterResourceFlags(registerResourceCmd)
}

func ValidateRegisterResourceFlags(cmd *cobra.Command, args []string) error {
	if !common.IsHexAddress(Handler) {
		return fmt.Errorf("invalid handler address %s", Handler)
	}
	if !common.IsHexAddress(Target) {
		return fmt.Errorf("invalid target address %s", Target)
	}
	if !common.IsHexAddress(Bridge) {
		return fmt.Errorf("invalid bridge address %s", Bridge)
	}
	return nil
}

func ProcessRegisterResourceFlags(cmd *cobra.Command, args []string) error {
	var err error
	HandlerAddr = common.HexToAddress(Handler)
	TargetContractAddr = common.HexToAddress(Target)
	BridgeAddr = common.HexToAddress(Bridge)

	ResourceIdBytesArr, err = flags.ProcessResourceID(ResourceID)
	return err
}

func RegisterResourceCmd(cmd *cobra.Command, args []string, contract *bridge.BridgeContract) error {
	h, err := contract.AdminSetResource(
		HandlerAddr, ResourceIdBytesArr, TargetContractAddr, transactor.TransactOptions{},
	)
	if err != nil {
		log.Error().Err(err)
		return err
	}

	fmt.Printf("Resource registered with hash: %s", h.Hex())
	return nil
}
