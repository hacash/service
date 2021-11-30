package rpc

import (
	"fmt"
	"github.com/hacash/core/actions"
	rpc "github.com/hacash/service/server"
	"strconv"
	"strings"
)

// 扫描区块 获取所有转账信息
func (api *DeprecatedApiService) getAllTransferLogByBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)

	is_include_btc_hacd := false // 是否返回 btc 和 hacd 转账
	_, is_include_btc_hacd = params["include_btc_hacd"]

	// 扫描的区块高度
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

	state := api.blockchain.StateRead()

	lastest, e3 := state.ReadLastestBlockHeadMetaForRead()
	if e3 != nil {
		result["err"] = e3.Error()
		return result
	}

	// 判断区块高度
	if block_height <= 0 || block_height > lastest.GetHeight() {
		result["err"] = "block height not find."
		result["ret"] = "1" // 返回错误码
		return result
	}

	// 查询区块
	tarblock, e := rpc.LoadBlockWithCache(api.backend.BlockChain().StateRead(), block_height)
	if e != nil {
		result["err"] = "read block data error."
		return result
	}

	// 开始扫描区块
	allTransferLogs := make([]string, 0, 4)
	transactions := tarblock.GetTransactions()
	for _, v := range transactions {
		if 0 == v.Type() { // coinbase
			continue
		}
		for _, act := range v.GetActions() {
			if 1 == act.Kind() { // 类型为HAC普通转账
				fromAddr := v.GetAddress()
				act_k1 := act.(*actions.Action_1_SimpleToTransfer)
				allTransferLogs = append(allTransferLogs,
					fromAddr.ToReadable()+"|"+
						act_k1.ToAddress.ToReadable()+"|"+
						act_k1.Amount.ToFinString(),
				)
			} else if 13 == act.Kind() { // 类型为HAC普通转账
				toAddr := v.GetAddress()
				act_k13 := act.(*actions.Action_13_FromTransfer)
				allTransferLogs = append(allTransferLogs,
					act_k13.FromAddress.ToReadable()+"|"+
						toAddr.ToReadable()+"|"+
						act_k13.Amount.ToFinString(),
				)
			} else if 14 == act.Kind() { // 类型为HAC普通转账
				act_k14 := act.(*actions.Action_14_FromToTransfer)
				allTransferLogs = append(allTransferLogs,
					act_k14.FromAddress.ToReadable()+"|"+
						act_k14.ToAddress.ToReadable()+"|"+
						act_k14.Amount.ToFinString(),
				)
			}
			// 是否返回 btc 和 HACD 转账
			if !is_include_btc_hacd {
				continue // 不包含
			}
			if 8 == act.Kind() { // BTC 转账
				fromAddr := v.GetAddress()
				act_k1 := act.(*actions.Action_8_SimpleSatoshiTransfer)
				allTransferLogs = append(allTransferLogs,
					fromAddr.ToReadable()+"|"+
						act_k1.ToAddress.ToReadable()+"|"+
						act_k1.Amount.ToString()+" SAT",
				)
			} else if 11 == act.Kind() { // BTC 转账
				act_k13 := act.(*actions.Action_11_FromToSatoshiTransfer)
				allTransferLogs = append(allTransferLogs,
					act_k13.FromAddress.ToReadable()+"|"+
						act_k13.ToAddress.ToReadable()+"|"+
						act_k13.Amount.ToString()+" SAT",
				)
			} else if 5 == act.Kind() { // HACD 转账
				fromAddr := v.GetAddress()
				act_k14 := act.(*actions.Action_5_DiamondTransfer)
				allTransferLogs = append(allTransferLogs,
					fromAddr.ToReadable()+"|"+
						act_k14.ToAddress.ToReadable()+"|"+
						"1 HACD",
				)
			} else if 6 == act.Kind() { // HACD 批量转账
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

	// 返回
	result["jsondata"] = `{"timestamp":` + strconv.FormatUint(tarblock.GetTimestamp(), 10) + `,"datas":[` + datasstr + `]}`
	return result
}
