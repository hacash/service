package rpc

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"strings"
)

func (api *DeprecatedApiService) hashRateChartsV3(params map[string]string) map[string]string {
	result := make(map[string]string)
	lastest, err1 := api.blockchain.State().ReadLastestBlockHeadAndMeta()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}

	var jsondatastring = `{"ret":0`

	curheight := lastest.GetHeight()
	blockstore := api.blockchain.State().BlockStore()

	// 当前
	var currentHashWorth *big.Int = nil
	targetHashWorth := difficulty.CalculateDifficultyWorth(curheight, lastest.GetDifficulty())
	// 当前实时哈希率： 4小时48区块所耗费的时间
	curCalcBlockNum := int64(48)
	prevHeight := int64(curheight) - curCalcBlockNum
	if prevHeight > 0 {
		headbytes, err2 := api.blockchain.State().BlockStore().ReadBlockHeadBytesByHeight(uint64(prevHeight))
		if err2 != nil {
			result["err"] = err2.Error()
			return result
		}
		blk, _, err3 := blocks.ParseBlockHead(headbytes, 0)
		if err3 != nil {
			result["err"] = err3.Error()
			return result
		}
		realEachBlockCostTimeSec := (lastest.GetTimestamp() - blk.GetTimestamp()) / uint64(curCalcBlockNum)
		// 实时哈希率
		//fmt.Println("realEachBlockCostTimeSec:  ", realEachBlockCostTimeSec)
		currentHashRate := new(big.Int).Div(targetHashWorth, new(big.Int).SetUint64(realEachBlockCostTimeSec))
		currentHashWorth = new(big.Int).Mul(currentHashRate, new(big.Int).SetUint64(300))
		jsondatastring += `,"current_hashrate":` + currentHashRate.String()
		jsondatastring += `,"current_show":"` + difficulty.ConvertPowPowerToShowFormat(currentHashRate) + `"`
	} else {
		jsondatastring += `,"current_hashrate":0`
		jsondatastring += `,"current_show":"0H/s"`
	}

	// 本周期目标哈希率
	targetHashRate := new(big.Int).Div(targetHashWorth, new(big.Int).SetUint64(uint64(mint.EachBlockRequiredTargetTime)))
	jsondatastring += `,"target_hashrate":` + targetHashRate.String()
	jsondatastring += `,"target_show":"` + difficulty.ConvertPowPowerToShowFormat(targetHashRate) + `"`

	days30, e2 := hashRateList(blockstore, curheight, false, currentHashWorth)
	if e2 != nil {
		result["err"] = e2.Error()
		return result
	}
	jsondatastring += `,"days30":[` + strings.Join(days30, ",") + `]`

	jsondatastring += "}"

	// ok
	result["jsondata"] = jsondatastring
	// `{"target":`+`,"current":`+currentHashRate.String()+`,"days30":[` + strings.Join(days30, ",") + `]}`

	// 返回
	return result

}
