package rpc

import (
	rpc "github.com/hacash/service/server"
	"net/http"
)

func (api *RpcService) totalSupply(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	state := api.backend.BlockChain().State()

	// get
	ttspl, e1 := state.ReadTotalSupply()
	if e1 != nil {
		ResponseError(w, e1)
		return
	}

	//fmt.Println(ttspl.Serialize())

	var data = make(map[string]interface{})

	// 读取流通量统计
	data, _ = rpc.RenderTotalSupplyObject(ttspl, false)

	// ok
	ResponseData(w, data)
}
