package rpc

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
	"os"
)

type DeprecatedApiServiceConfig struct {
	HttpListenPort int
}

func NewEmptyDeprecatedApiServiceConfig() *DeprecatedApiServiceConfig {
	return &DeprecatedApiServiceConfig{
		HttpListenPort: 3338,
	}
}

func NewDeprecatedApiServiceConfig(inicnf *sys.Inicnf) *DeprecatedApiServiceConfig {
	cnf := NewEmptyDeprecatedApiServiceConfig()

	section := inicnf.Section("service")

	cnf.HttpListenPort = section.Key("deprecated_http_port").MustInt(3338)

	return cnf
}

/////////////////////////////////////

type DeprecatedApiService struct {
	config *DeprecatedApiServiceConfig

	blockchain interfaces.BlockChain

	txpool interfaces.TxPool
}

func NewDeprecatedApiService(cnf *DeprecatedApiServiceConfig) *DeprecatedApiService {
	return &DeprecatedApiService{
		config:     cnf,
		blockchain: nil,
		txpool:     nil,
	}
}

func (api *DeprecatedApiService) Start() {
	if api.blockchain == nil {
		fmt.Println("api.blockchain not be set.")
		os.Exit(0)
	}
	if api.txpool == nil {
		fmt.Println("api.txpool not be set.")
		os.Exit(0)
	}
	// start

	api.RunHttpRpcService(api.config.HttpListenPort)
}

func (api *DeprecatedApiService) SetBlockChain(blockchain interfaces.BlockChain) {
	api.blockchain = blockchain
}

func (api *DeprecatedApiService) SetTxPool(txpool interfaces.TxPool) {
	api.txpool = txpool
}
