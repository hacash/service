package rpc

import "net/http"

func (api *RpcService) lastBlock(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	state := api.backend.BlockChain().State()

	// get
	lastblk, e1 := state.ReadLastestBlockHeadAndMeta()
	if e1 != nil {
		ResponseError(w, e1)
		return
	}

	retHead := CheckParamBool(r, "head", false)
	retInfo := CheckParamBool(r, "info", false)

	// return
	data := ResponseCreateData("height", lastblk.GetHeight())

	if retHead || retInfo {
		data["hash"] = lastblk.HashFresh().ToHex()
		data["timestamp"] = lastblk.GetTimestamp()
		data["prev_hash"] = lastblk.GetPrevHash().ToHex()
		data["mrkl_root"] = lastblk.GetMrklRoot().ToHex()
		data["transaction_count"] = lastblk.GetCustomerTransactionCount()
	}

	ResponseData(w, data)
}
