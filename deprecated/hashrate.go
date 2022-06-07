package rpc

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
	"math/big"
)

func (api *DeprecatedApiService) hashRate(params map[string]string) map[string]string {
	result := make(map[string]string)
	lastest, _, err1 := api.blockchain.GetChainEngineKernel().LatestBlock()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}

	curheight := lastest.GetHeight()
	targetHashWorth := difficulty.CalculateDifficultyWorth(lastest.GetDifficulty())

	// Current real-time hash rate: time spent on 48 blocks in 4 hours
	curCalcBlockNum := int64(48)
	prevHeight := int64(curheight) - curCalcBlockNum
	if prevHeight > 0 {
		_, blockbytes, err2 := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead().ReadBlockBytesByHeight(uint64(prevHeight))
		if err2 != nil {
			result["err"] = err2.Error()
			return result
		}
		blk, _, err3 := blocks.ParseBlockHead(blockbytes, 0)
		if err3 != nil {
			result["err"] = err3.Error()
			return result
		}
		realEachBlockCostTimeSec := (lastest.GetTimestamp() - blk.GetTimestamp()) / uint64(curCalcBlockNum)
		// Real time hash rate
		//fmt.Println("realEachBlockCostTimeSec:  ", realEachBlockCostTimeSec)
		currentHashRate := new(big.Int).Div(targetHashWorth, new(big.Int).SetUint64(realEachBlockCostTimeSec))
		result["current_hashrate"] = currentHashRate.String()
		result["current_show"] = difficulty.ConvertPowPowerToShowFormat(currentHashRate)
	} else {
		result["current_hashrate"] = "0"
		result["target_show"] = "0H/s"
	}

	// Target hash rate of this cycle
	targetHashRate := new(big.Int).Div(targetHashWorth, new(big.Int).SetUint64(uint64(mint.EachBlockRequiredTargetTime)))
	result["target_hashrate"] = targetHashRate.String()
	result["target_show"] = difficulty.ConvertPowPowerToShowFormat(targetHashRate)

	// return
	return result
}
