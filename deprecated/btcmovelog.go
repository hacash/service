package rpc

import (
	"github.com/hacash/core/stores"
	"strconv"
	"strings"
)

// Query linear lock information
func (api *DeprecatedApiService) getBtcMoveLogPageData(params map[string]string) map[string]string {
	result := make(map[string]string)
	page, e0 := strconv.ParseUint(params["page"], 10, 0)
	if e0 != nil {
		result["err"] = "param page must."
		return result
	}

	// query
	datas, e2 := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead().GetBTCMoveLogPageData(int(page))
	if e2 != nil {
		result["err"] = "not find."
		return result
	}

	// Return information
	result["jsondata"] = "[\"" + strings.Join(stores.SatoshiGenesisPageSerializeForShow(datas), "\",\"") + "\"]"

	return result
}
