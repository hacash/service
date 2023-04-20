package rpc

import (
	"fmt"
	"github.com/hacash/core/actions"
	rpc "github.com/hacash/service/server"
	"strconv"
	"strings"
)

// Scan the block to obtain all transfer information
func (api *DeprecatedApiService) getAllTransferLogByBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)

	is_include_btc_hacd := false // 是否返回 btc 和 hacd 转账
	_, is_include_btc_hacd = params["include_btc_hacd"]

	// Scanned block height
	block_height_str, ok1 := params["block_height"]
	if !ok1 {
		result["err"] = "param block_height must."
		return result
	}

	block_height, err2 := strconv.ParseUint(block_height_str, 10, 0)
	if err2 != nil {
		result["err"] = "param block_height format error."
		return result
	}

	lastest, _, e3 := api.blockchain.GetChainEngineKernel().LatestBlock()
	if e3 != nil {
		result["err"] = e3.Error()
		return result
	}

	// Judge block height
	if block_height <= 0 || block_height > lastest.GetHeight() {
		result["err"] = "block height not find."
		result["ret"] = "1" // Return error code
		return result
	}

	// Query block
	tarblock, e := rpc.LoadBlockWithCache(api.backend.BlockChain().GetChainEngineKernel(), block_height)
	if e != nil {
		result["err"] = "read block data error."
		return result
	}

	must_confirm, _ := params["must_confirm"]
	fmt.Println(must_confirm)
	if len(must_confirm) > 0 {
		var okey_block_hei = lastest.GetHeight()
		// fmt.Println("okey_block_hei: ", okey_block_hei)
		must_confirm_block_hei, _ := strconv.ParseUint(must_confirm, 10, 0)
		if block_height > 0 && must_confirm_block_hei > 0 && block_height+must_confirm_block_hei > okey_block_hei {
			// fmt.Println("must_confirm_block_hei: ", must_confirm_block_hei)
			result["err"] = fmt.Sprintf("block %d not be confirm", block_height)
			return result
		}
	}

	// Start scanning block
	allTransferLogs := make([]string, 0, 4)
	transactions := tarblock.GetTrsList()
	for _, v := range transactions {
		if 0 == v.Type() { // coinbase
			continue
		}
		for _, act := range v.GetActionList() {
			if 1 == act.Kind() { // HAC ordinary transfer
				fromAddr := v.GetAddress()
				act_k1 := act.(*actions.Action_1_SimpleToTransfer)
				allTransferLogs = append(allTransferLogs,
					fromAddr.ToReadable()+"|"+
						act_k1.ToAddress.ToReadable()+"|"+
						act_k1.Amount.ToFinString(),
				)
			} else if 13 == act.Kind() { // HAC ordinary transfer
				toAddr := v.GetAddress()
				act_k13 := act.(*actions.Action_13_FromTransfer)
				allTransferLogs = append(allTransferLogs,
					act_k13.FromAddress.ToReadable()+"|"+
						toAddr.ToReadable()+"|"+
						act_k13.Amount.ToFinString(),
				)
			} else if 14 == act.Kind() { // HAC ordinary transfer
				act_k14 := act.(*actions.Action_14_FromToTransfer)
				allTransferLogs = append(allTransferLogs,
					act_k14.FromAddress.ToReadable()+"|"+
						act_k14.ToAddress.ToReadable()+"|"+
						act_k14.Amount.ToFinString(),
				)
			}
			// Whether to return BTC and hacd transfer
			if !is_include_btc_hacd {
				continue // Not included
			}
			if 8 == act.Kind() { // BTC transfer
				fromAddr := v.GetAddress()
				act_k1 := act.(*actions.Action_8_SimpleSatoshiTransfer)
				allTransferLogs = append(allTransferLogs,
					fromAddr.ToReadable()+"|"+
						act_k1.ToAddress.ToReadable()+"|"+
						act_k1.Amount.ToString()+" SAT",
				)
			} else if 11 == act.Kind() { // BTC transfer
				act_k13 := act.(*actions.Action_11_FromToSatoshiTransfer)
				allTransferLogs = append(allTransferLogs,
					act_k13.FromAddress.ToReadable()+"|"+
						act_k13.ToAddress.ToReadable()+"|"+
						act_k13.Amount.ToString()+" SAT",
				)
			} else if 5 == act.Kind() { // Hacd transfer
				fromAddr := v.GetAddress()
				act_k14 := act.(*actions.Action_5_DiamondTransfer)
				allTransferLogs = append(allTransferLogs,
					fromAddr.ToReadable()+"|"+
						act_k14.ToAddress.ToReadable()+"|"+
						"1 HACD",
				)
			} else if 6 == act.Kind() { // Hacd batch transfer
				act_k14 := act.(*actions.Action_6_OutfeeQuantityDiamondTransfer)
				allTransferLogs = append(allTransferLogs,
					act_k14.FromAddress.ToReadable()+"|"+
						act_k14.ToAddress.ToReadable()+"|"+
						fmt.Sprintf("%d HACD", act_k14.DiamondList.Count),
				)
			}

		}
	}

	datasstr := strings.Join(allTransferLogs, "\",\"")
	if len(datasstr) > 0 {
		datasstr = "\"" + datasstr + "\""
	}

	// return
	result["jsondata"] = `{"timestamp":` + strconv.FormatUint(tarblock.GetTimestamp(), 10) + `,"datas":[` + datasstr + `]}`
	return result
}
