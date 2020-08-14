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
			break // 最多查询20个
		}
		addrhash, e := fields.CheckReadableAddress(addr)
		if e != nil {
			amtstrings += "[format error],"
			continue
		}
		finditem := state.Balance(*addrhash)
		if finditem == nil {
			amtstrings += "ㄜ0:0,"
			continue
		}
		amtstrings += strings.Trim(finditem.Amount.ToFinString(), ",") + ","
		totalamt, _ = totalamt.Add(&finditem.Amount)
		// satoshi
		sts := state.Satoshi(*addrhash)
		if sts != nil {
			satoshi += uint64(sts.Amount)
			satoshistrs += strconv.FormatUint(uint64(sts.Amount), 10) + ","
		} else {
			satoshistrs += "0,"
		}
	}

	// 0
	result["amounts"] = strings.TrimRight(amtstrings, ",")
	result["total"] = totalamt.ToFinString()
	result["satoshis"] = strings.TrimRight(satoshistrs, ",")
	result["satoshi"] = strconv.FormatUint(satoshi, 10)
	return result

}
