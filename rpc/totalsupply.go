package rpc

import (
	rpc "github.com/hacash/service/server"
	"net/http"
)

func (api *RpcService) totalSupply(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	state := api.backend.BlockChain().GetChainEngineKernel().StateRead()

	// get
	ttspl, e1 := state.ReadTotalSupply()
	if e1 != nil {
		ResponseError(w, e1)
		return
	}

	var data = make(map[string]interface{})

	// Read Circulation Statistics
	data, _ = rpc.RenderTotalSupplyObject(state, ttspl, false)

	// ok
	ResponseData(w, data)
}
