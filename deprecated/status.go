package rpc

import (
	"fmt"
	"github.com/hacash/core/transactions"
	"strings"
)

// GetRecentArrivedBlocks
func (api *DeprecatedApiService) getRecentArrivedBlockList(params map[string]string) map[string]string {
	result := make(map[string]string)

	var is_brief = false
	brief, ok1 := params["brief"]
	if ok1 && len(brief) > 0 {
		is_brief = true
	}

	blklist := api.blockchain.GetChainEngineKernel().GetRecentArrivedBlocks()
	var resjsonobjarys = make([]string, len(blklist))

	for i, v := range blklist {
		var trs = v.GetTrsList()
		if len(trs) <= 0 {
			continue
		}
		var cbobj, ok = trs[0].(*transactions.Transaction_0_Coinbase)
		if !ok {
			continue
		}
		hx1 := v.Hash().ToHex()
		hx2 := v.GetPrevHash().ToHex()
		addr := cbobj.Address.ToReadable()
		if is_brief {
			hx1 = hx1[64-12:]
			hx2 = hx2[64-12:]
			addr = addr[:10]
		}
		// ok
		resjsonobjarys[i] = fmt.Sprintf(
			`{"height":%d,"hx":"%s","prev":"%s","txs":%d,"miner":"%s","msg":"%s","arrive":%d}`,
			v.GetHeight(),
			hx1,
			hx2,
			len(trs)-1,
			addr,
			cbobj.Message.ValueShow(),
			v.ArrivedTime(),
		)
	}

	result["jsondata"] = fmt.Sprintf(`{"ret":0,"list":[%s]}`,
		strings.Join(resjsonobjarys, ","))

	return result

}
