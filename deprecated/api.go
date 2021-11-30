package rpc

import (
	"fmt"
	"github.com/hacash/core/interfacev2"
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

	backend interfacev2.Backend

	blockchain interfacev2.BlockChain

	txpool interfacev2.TxPool
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

func (api *DeprecatedApiService) SetBlockChain(blockchain interfacev2.BlockChain) {
	api.blockchain = blockchain
}

func (api *DeprecatedApiService) SetTxPool(txpool interfacev2.TxPool) {
	api.txpool = txpool
}

func (api *DeprecatedApiService) SetBackend(backend interfacev2.Backend) {
	api.backend = backend
}
