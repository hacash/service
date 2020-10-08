package rpc

import "net/http"

func (api *RpcService) lastBlockHeight(r *http.Request, w http.ResponseWriter) {

	state := api.backend.BlockChain().State()

	// get
	lastblk, e1 := state.ReadLastestBlockHeadAndMeta()
	if e1 != nil {
		ResponseError(w, e1)
		return
	}

	// return
	ResponseData(w, ResponseCreateData("height", lastblk.GetHeight()))
}
