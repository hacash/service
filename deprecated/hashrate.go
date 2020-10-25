package rpc

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
	"math/big"
)

func (api *DeprecatedApiService) hashRate(params map[string]string) map[string]string {
	result := make(map[string]string)
	lastest, err1 := api.blockchain.State().ReadLastestBlockHeadAndMeta()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}
	curheight := lastest.GetHeight()
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
		result["current_hashrate"] = currentHashRate.String()
		result["current_show"] = difficulty.ConvertPowPowerToShowFormat(currentHashRate)
	} else {
		result["current_hashrate"] = "0"
		result["target_show"] = "0H/s"
	}

	// 本周期目标哈希率
	targetHashRate := new(big.Int).Div(targetHashWorth, new(big.Int).SetUint64(uint64(mint.EachBlockRequiredTargetTime)))
	result["target_hashrate"] = targetHashRate.String()
	result["target_show"] = difficulty.ConvertPowPowerToShowFormat(targetHashRate)

	// 返回
	return result

}
