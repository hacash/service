package rpc

import (
	rpc "github.com/hacash/service/server"
)

func (api *DeprecatedApiService) totalSupply(params map[string]string) map[string]string {
	result := make(map[string]string)
	state := api.blockchain.State()

	// get
	ttspl, e1 := state.ReadTotalSupply()
	if e1 != nil {
		result["err"] = e1.Error()
		return result
	}

	//fmt.Println(ttspl.Serialize())
	_, result = rpc.RenderTotalSupplyObject(ttspl, true)

	// ok
	return result

}
