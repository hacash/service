package rpc

import "net/http"

func (api *RpcService) lastBlock(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	kernel := api.backend.BlockChain().GetChainEngineKernel()

	// get
	lastblk, _, e1 := kernel.LatestBlock()
	if e1 != nil {
		ResponseError(w, e1)
		return
	}

	retHead := CheckParamBool(r, "head", false)
	retInfo := CheckParamBool(r, "info", false)

	// return
	data := ResponseCreateData("height", lastblk.GetHeight())

	if retHead || retInfo {
		data["hash"] = lastblk.Hash().ToHex()
		data["timestamp"] = lastblk.GetTimestamp()
		data["prev_hash"] = lastblk.GetPrevHash().ToHex()
		data["mrkl_root"] = lastblk.GetMrklRoot().ToHex()
		data["transaction_count"] = lastblk.GetCustomerTransactionCount()
	}

	ResponseData(w, data)
}
