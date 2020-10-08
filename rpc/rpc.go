package rpc

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
	"net/http"
	"os"
)

type RpcConfig struct {
	HttpListenPort int
}

func NewEmptyRpcConfig() *RpcConfig {
	return &RpcConfig{
		HttpListenPort: 8083,
	}
}

func NewRpcConfig(inicnf *sys.Inicnf) *RpcConfig {
	cnf := NewEmptyRpcConfig()

	section := inicnf.Section("service")

	cnf.HttpListenPort = section.Key("rpc_listen_port").MustInt(0)

	return cnf
}

/////////////////////////////////////

type RpcService struct {
	config *RpcConfig

	backend interfaces.Backend
	txpool  interfaces.TxPool

	queryRoutes map[string]func(*http.Request, http.ResponseWriter)
}

func NewRpcService(cnf *RpcConfig) *RpcService {
	return &RpcService{
		config:      cnf,
		backend:     nil,
		txpool:      nil,
		queryRoutes: make(map[string]func(*http.Request, http.ResponseWriter)),
	}
}

func (api *RpcService) Start() {
	if api.backend == nil {
		fmt.Println("api.backend not be set.")
		os.Exit(0)
	}
	if api.txpool == nil {
		fmt.Println("api.txpool not be set.")
		os.Exit(0)
	}
	// start
	api.RunHttpRpcService(api.config.HttpListenPort)
}

func (api *RpcService) SetTxPool(txpool interfaces.TxPool) {
	api.txpool = txpool
}

func (api *RpcService) SetBackend(backend interfaces.Backend) {
	api.backend = backend
}
