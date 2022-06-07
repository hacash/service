package rpc

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"strings"
	"sync"
	"time"
)

// cache
var hashRateChartsV3_lock sync.Mutex
var hashRateChartsV3_lastreqtime *time.Time = nil
var hashRateChartsV3_cachedata map[string]string = nil

func (api *DeprecatedApiService) hashRateChartsV3(params map[string]string) map[string]string {
	result := make(map[string]string)

	hashRateChartsV3_lock.Lock()
	var tn = time.Now()

	// Check cache
	if hashRateChartsV3_cachedata != nil && hashRateChartsV3_lastreqtime != nil {
		if tn.Sub(*hashRateChartsV3_lastreqtime) < time.Minute*time.Duration(5) {
			hashRateChartsV3_lock.Unlock() // Unlock
			//fmt.Println("hashRateChartsV3_cachedata // 返回缓存")
			return hashRateChartsV3_cachedata // Return cache
		}
	}

	hashRateChartsV3_lock.Unlock()

	// locking
	hashRateChartsV3_lock.Lock()
	defer hashRateChartsV3_lock.Unlock()

	// Officially start calculation
	lastest, _, err1 := api.blockchain.GetChainEngineKernel().LatestBlock()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}

	var jsondatastring = `{"ret":0`

	curheight := lastest.GetHeight()
	blockstore := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()

	// current
	var currentHashWorth *big.Int = nil
	targetHashWorth := difficulty.CalculateDifficultyWorth(lastest.GetDifficulty())
	// Current real-time hash rate: time spent on 300 blocks
	curCalcBlockNum := int64(300)
	prevHeight := int64(curheight) - curCalcBlockNum
	if prevHeight > 0 {
		_, headbytes, err2 := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead().ReadBlockBytesByHeight(uint64(prevHeight))
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
		// Real time hash rate
		//fmt.Println("realEachBlockCostTimeSec:  ", realEachBlockCostTimeSec)
		currentHashRate := new(big.Int).Div(targetHashWorth, new(big.Int).SetUint64(realEachBlockCostTimeSec))
		currentHashWorth = new(big.Int).Mul(currentHashRate, new(big.Int).SetUint64(mint.EachBlockRequiredTargetTime))
		jsondatastring += `,"current_hashrate":` + currentHashRate.String()
		jsondatastring += `,"current_show":"` + difficulty.ConvertPowPowerToShowFormat(currentHashRate) + `"`
	} else {
		jsondatastring += `,"current_hashrate":0`
		jsondatastring += `,"current_show":"0H/s"`
	}

	// Target hash rate of this cycle
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

	// cache
	hashRateChartsV3_lastreqtime = &tn  // Cache time
	hashRateChartsV3_cachedata = result // Cache data

	// return
	return result
}
