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

	state := api.blockchain.GetChainEngineKernel().StateRead()

	addrs := strings.Split(addrstr, ",")
	amtstrings := ""
	totalamt := fields.NewEmptyAmount()
	satoshi := uint64(0)
	satoshistrs := ""
	diamond := uint64(0)
	diamondstrs := ""
	for k, addr := range addrs {
		if k > 20 {
			break // 一次最多查询20个
		}
		addrhash, e := fields.CheckReadableAddress(addr)
		if e != nil {
			amtstrings += "[format error]|"
			continue
		}
		finditem, e := state.Balance(*addrhash)
		if e != nil {
			amtstrings += e.Error()
			continue
		}

		if finditem == nil {
			amtstrings += "ㄜ0:0|"
			satoshistrs += "0|"
			diamondstrs += "0|"
			continue
		}
		// hacash
		amtstrings += finditem.Hacash.ToFinString() + "|"
		totalamt, _ = totalamt.Add(&finditem.Hacash)
		// satoshi
		satoshi += uint64(finditem.Satoshi)
		satoshistrs += strconv.FormatUint(uint64(finditem.Satoshi), 10) + "|"
		// diamond
		diamond += uint64(finditem.Diamond)
		diamondstrs += strconv.FormatUint(uint64(finditem.Diamond), 10) + "|"
	}

	// 0
	result["amounts"] = strings.TrimRight(amtstrings, "|")
	result["total"] = totalamt.ToFinString()
	result["satoshis"] = strings.TrimRight(satoshistrs, "|")
	result["satoshi"] = strconv.FormatUint(satoshi, 10)
	result["diamonds"] = strings.TrimRight(diamondstrs, "|")
	result["diamond"] = strconv.FormatUint(diamond, 10)
	return result

}
