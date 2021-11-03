package rpc

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
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

	cnf.HttpListenPort = section.Key("deprecated_http_port").MustInt(0)

	return cnf
}

/////////////////////////////////////

type DeprecatedApiService struct {
	config *DeprecatedApiServiceConfig

	backend interfaces.Backend

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

func (api *DeprecatedApiService) Start() error {
	if api.blockchain == nil {
		return fmt.Errorf("api.blockchain not be set.")
	}
	if api.txpool == nil {
		return fmt.Errorf("api.txpool not be set.")
	}
	// start
	api.RunHttpRpcService(api.config.HttpListenPort)
	return nil
}

func (api *DeprecatedApiService) SetBlockChain(blockchain interfaces.BlockChain) {
	api.blockchain = blockchain
}

func (api *DeprecatedApiService) SetTxPool(txpool interfaces.TxPool) {
	api.txpool = txpool
}

func (api *DeprecatedApiService) SetBackend(backend interfaces.Backend) {
	api.backend = backend
}
