package rpc

import (
	"github.com/hacash/core/fields"
	"net/http"
	"strings"
)

// View account balance
func (api *RpcService) balances(r *http.Request, w http.ResponseWriter, bodybytes []byte) {
	// Address list
	addresslistStr := strings.Trim(CheckParamString(r, "address_list", ""), " ")
	if len(addresslistStr) == 0 {
		ResponseErrorString(w, "param address_list must give")
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
		ResponseErrorString(w, "param address_list must give")
		return
	}

	if len(addresses) > 200 {
		ResponseErrorString(w, "address number cannot over 200")
		return
	}

	// kind = hsd
	kindStr := strings.ToLower(CheckParamString(r, "kind", ""))
	actAllKinds := false // 支持全部种类
	actKindHacash := false
	actKindSatoshi := false
	actKindDiamond := false
	if len(kindStr) == 0 {
		actAllKinds = true
	} else {
		if strings.Contains(kindStr, "h") {
			actKindHacash = true
		}
		if strings.Contains(kindStr, "s") {
			actKindSatoshi = true
		}
		if strings.Contains(kindStr, "d") {
			actKindDiamond = true
		}
	}

	// Is it in pieces
	isUnitMei := CheckParamBool(r, "unitmei", false)
	emptyAmt := fields.NewAmountSmall(0, 0)

	// read
	var lists = make([]interface{}, len(addresses))
	state := api.backend.BlockChain().GetChainEngineKernel().StateRead()
	for i, addr := range addresses {
		var item = make(map[string]interface{})
		bls, e := state.Balance(*addr)
		if e != nil {
			continue
		}

		if actAllKinds || actKindHacash {
			if bls != nil {
				item["hacash"] = bls.Hacash.ToMeiOrFinString(isUnitMei)
			} else {
				item["hacash"] = emptyAmt.ToMeiOrFinString(isUnitMei)
			}
		}

		if actAllKinds || actKindSatoshi {
			if bls != nil {
				item["satoshi"] = uint64(bls.Satoshi)
			} else {
				item["satoshi"] = 0
			}
		}

		if actAllKinds || actKindDiamond {
			if bls != nil {
				item["diamond"] = uint64(bls.Diamond)
			} else {
				item["diamond"] = 0
			}
		}
		lists[i] = item
	}

	// return
	ResponseList(w, lists)
	return
}
