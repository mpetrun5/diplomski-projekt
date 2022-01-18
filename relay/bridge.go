package bridge

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mpetrun5/diplomski-projekt/chains/evm"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/contracts/bridge"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmclient"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmgaspricer"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/evmtransaction"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/calls/transactor/signAndSend"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/listener"
	"github.com/mpetrun5/diplomski-projekt/chains/evm/voter"
	"github.com/mpetrun5/diplomski-projekt/config"
	"github.com/mpetrun5/diplomski-projekt/config/chain"
	"github.com/mpetrun5/diplomski-projekt/crypto/secp256k1"
	"github.com/mpetrun5/diplomski-projekt/flags"
	"github.com/mpetrun5/diplomski-projekt/lvldb"
	"github.com/mpetrun5/diplomski-projekt/relayer"
	"github.com/mpetrun5/diplomski-projekt/store"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Run() error {
	errChn := make(chan error)
	stopChn := make(chan struct{})

	configuration, err := config.GetConfig(viper.GetString(flags.ConfigFlagName))
	db, err := lvldb.NewLvlDB(viper.GetString(flags.BlockstoreFlagName))
	if err != nil {
		panic(err)
	}
	blockstore := store.NewBlockStore(db)

	chains := []relayer.RelayedChain{}
	for _, chainConfig := range configuration.ChainConfigs {
		config, err := chain.NewEVMConfig(chainConfig)
		if err != nil {
			panic(err)
		}

		kp, err := secp256k1.GenerateKeypair()
		if err != nil {
			panic(err)
		}

		client, err := evmclient.NewEVMClientFromParams(config.GeneralChainConfig.Endpoint, kp.PrivateKey())
		if err != nil {
			panic(err)
		}
		gasPricer := evmgaspricer.NewStaticGasPriceDeterminant(client)
		t := signAndSend.NewSignAndSendTransactor(evmtransaction.NewTransaction, gasPricer, client)
		bridgeContract := bridge.NewBridgeContract(client, common.HexToAddress(config.Bridge), t)

		eventHandler := listener.NewETHEventHandler(*bridgeContract)
		eventHandler.RegisterEventHandler(config.Erc20Handler, listener.Erc20EventHandler)
		evmListener := listener.NewEVMListener(client, eventHandler, common.HexToAddress(config.Bridge))

		mh := voter.NewEVMMessageHandler(*bridgeContract)
		mh.RegisterMessageHandler(config.Erc20Handler, voter.ERC20MessageHandler)

		var evmVoter *voter.EVMVoter
		evmVoter = voter.NewVoter(mh, bridgeContract)

		chains = append(chains, evm.NewEVMChain(evmListener, evmVoter, blockstore, config))
	}

	r := relayer.NewRelayer(chains)
	go r.Start(stopChn, errChn)

	sysErr := make(chan os.Signal, 1)
	signal.Notify(
		sysErr,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT)

	select {
	case err := <-errChn:
		log.Error().Err(err).Msg("failed to listen and serve")
		close(stopChn)
		return err
	case sig := <-sysErr:
		log.Info().Msgf("terminating got [%v] signal", sig)
		return nil
	}
}
