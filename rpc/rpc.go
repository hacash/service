package rpc

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
	"net/http"
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

	// routes
	queryRoutes   map[string]func(*http.Request, http.ResponseWriter, []byte)
	createRoutes  map[string]func(*http.Request, http.ResponseWriter, []byte)
	submitRoutes  map[string]func(*http.Request, http.ResponseWriter, []byte)
	operateRoutes map[string]func(*http.Request, http.ResponseWriter, []byte)
}

func NewRpcService(cnf *RpcConfig) *RpcService {
	return &RpcService{
		config:        cnf,
		backend:       nil,
		txpool:        nil,
		queryRoutes:   make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		createRoutes:  make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		submitRoutes:  make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		operateRoutes: make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
	}
}

func (api *RpcService) Start() error {
	if api.backend == nil {
		return fmt.Errorf("api.backend not be set.")
	}
	if api.txpool == nil {
		return fmt.Errorf("api.txpool not be set.")
	}
	// start
	api.RunHttpRpcService(api.config.HttpListenPort)
	return nil
}

func (api *RpcService) SetTxPool(txpool interfaces.TxPool) {
	api.txpool = txpool
}

func (api *RpcService) SetBackend(backend interfaces.Backend) {
	api.backend = backend
}
