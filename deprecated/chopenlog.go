package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/actions"
	rpc "github.com/hacash/service/server"
	"strconv"
	"strings"
)

// 扫描区块 获取所有通道开启交易
func (api *DeprecatedApiService) getAllChannelOpenLogByBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)
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

	state := api.blockchain.State()

	lastest, e3 := state.ReadLastestBlockHeadAndMeta()
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
	tarblock, e := rpc.LoadBlockWithCache(api.backend.BlockChain().State(), block_height)
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
			if 2 == act.Kind() { // 类型为开启通带
				act_k2 := act.(*actions.Action_2_OpenPaymentChannel)
				idhex := hex.EncodeToString(act_k2.ChannelId)
				allTransferLogs = append(allTransferLogs,
					idhex+"|"+
						act_k2.LeftAddress.ToReadable()+"|"+
						act_k2.LeftAmount.ToFinString()+"|"+
						act_k2.RightAddress.ToReadable()+"|"+
						act_k2.RightAmount.ToFinString())
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
