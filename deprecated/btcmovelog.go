package rpc

import (
	"strconv"
	"strings"
)

// 查询线性锁仓信息
func (api *DeprecatedApiService) getBtcMoveLogPageData(params map[string]string) map[string]string {
	result := make(map[string]string)
	page, e0 := strconv.ParseUint(params["page"], 10, 0)
	if e0 != nil {
		result["err"] = "param page must."
		return result
	}

	// 查询
	datas, e2 := api.blockchain.State().BlockStore().GetBTCMoveLogPageData(int(page))
	if e2 != nil {
		result["err"] = "not find."
		return result
	}

	// 返回信息

	result["jsondata"] = "[\"" + strings.Join(datas, "\",\"") + "\"]"
	return result
}
