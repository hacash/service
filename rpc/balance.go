package rpc

import (
	"github.com/hacash/core/fields"
	"net/http"
	"strings"
)

// 查看账户余额
func (api *RpcService) balance(r *http.Request, w http.ResponseWriter) {

	// 地址列表
	addresslistStr := strings.Trim(CheckParamString(r, "addresslist", ""), " ")
	if len(addresslistStr) == 0 {
		ResponseErrorString(w, "param addresslist must give")
		return
	}
	addresslists := strings.Split(addresslistStr, ",")
	addresses := make([]*fields.Address, len(addresslists))
	for i, v := range addresslists {
		addr, e := fields.CheckReadableAddress(v)
		if e != nil {
			ResponseError(w, e)
			return
		}
		addresses[i] = addr
	}
	if len(addresses) == 0 {
		ResponseErrorString(w, "param addresslist must give")
		return
	}
	if len(addresses) > 200 {
		ResponseErrorString(w, "address bumber cannot over 200")
		return
	}

	// kind = hsd
	kindStr := strings.ToLower(CheckParamString(r, "kind", ""))
	actAllKinds := false // 支持全部种类
	actKindHacash := false
	actKindSatoshi := false
	if len(kindStr) == 0 {
		actAllKinds = true
	} else {
		if strings.Contains(kindStr, "h") {
			actKindHacash = true
		}
		if strings.Contains(kindStr, "s") {
			actKindSatoshi = true
		}
	}

	// 是否以枚为单位
	isUnitMei := CheckParamBool(r, "unitmei", false)
	emptyAmt := fields.NewAmountSmall(0, 0)

	// read
	var lists = make([]interface{}, len(addresses))
	state := api.backend.BlockChain().State()
	for i, addr := range addresses {
		var item = make(map[string]interface{})
		if actAllKinds || actKindHacash {
			bls := state.Balance(*addr)
			if bls != nil {
				item["hacash"] = bls.Amount.ToMeiOrFinString(isUnitMei)
			} else {
				item["hacash"] = emptyAmt.ToMeiOrFinString(isUnitMei)
			}
		}
		if actAllKinds || actKindSatoshi {
			sat := state.Satoshi(*addr)
			if sat != nil {
				item["satoshi"] = uint64(sat.Amount)
			} else {
				item["satoshi"] = 0
			}
		}
		lists[i] = item
	}

	// return
	ResponseList(w, lists)
	return
}
