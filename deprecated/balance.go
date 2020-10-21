package rpc

import (
	"github.com/hacash/core/fields"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) getBalance(params map[string]string) map[string]string {
	result := make(map[string]string)
	addrstr, ok1 := params["address"]
	if !ok1 {
		result["err"] = "address must"
		return result
	}

	state := api.blockchain.State()

	addrs := strings.Split(addrstr, ",")
	amtstrings := ""
	totalamt := fields.NewEmptyAmount()
	satoshi := uint64(0)
	satoshistrs := ""
	for k, addr := range addrs {
		if k > 20 {
			break // 一次最多查询20个
		}
		addrhash, e := fields.CheckReadableAddress(addr)
		if e != nil {
			amtstrings += "[format error]|"
			continue
		}
		finditem := state.Balance(*addrhash)
		if finditem == nil {
			amtstrings += "ㄜ0:0|"
			satoshistrs += "0|"
			continue
		}
		amtstrings += finditem.Hacash.ToFinString() + "|"
		totalamt, _ = totalamt.Add(&finditem.Hacash)
		// satoshi
		satoshi += uint64(finditem.Satoshi)
		satoshistrs += strconv.FormatUint(uint64(finditem.Satoshi), 10) + "|"
	}

	// 0
	result["amounts"] = strings.TrimRight(amtstrings, "|")
	result["total"] = totalamt.ToFinString()
	result["satoshis"] = strings.TrimRight(satoshistrs, "|")
	result["satoshi"] = strconv.FormatUint(satoshi, 10)
	return result

}
