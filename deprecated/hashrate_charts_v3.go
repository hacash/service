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

// 缓存
var hashRateChartsV3_lock sync.Mutex
var hashRateChartsV3_lastreqtime *time.Time = nil
var hashRateChartsV3_cachedata map[string]string = nil

func (api *DeprecatedApiService) hashRateChartsV3(params map[string]string) map[string]string {
	result := make(map[string]string)

	hashRateChartsV3_lock.Lock()
	var tn = time.Now()

	// 检查缓存
	if hashRateChartsV3_cachedata != nil && hashRateChartsV3_lastreqtime != nil {
		if tn.Sub(*hashRateChartsV3_lastreqtime) < time.Minute*time.Duration(5) {
			hashRateChartsV3_lock.Unlock() // 解锁
			//fmt.Println("hashRateChartsV3_cachedata // 返回缓存")
			return hashRateChartsV3_cachedata // 返回缓存
		}
	}
	//fmt.Println("////////////////////")
	hashRateChartsV3_lock.Unlock()

	// 锁定
	hashRateChartsV3_lock.Lock()
	defer hashRateChartsV3_lock.Unlock()

	// 正式开始计算
	lastest, _, err1 := api.blockchain.GetChainEngineKernel().LatestBlock()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}

	var jsondatastring = `{"ret":0`

	curheight := lastest.GetHeight()
	blockstore := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()

	// 当前
	var currentHashWorth *big.Int = nil
	targetHashWorth := difficulty.CalculateDifficultyWorth(lastest.GetDifficulty())
	// 当前实时哈希率： 300区块所耗费的时间
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
		// 实时哈希率
		//fmt.Println("realEachBlockCostTimeSec:  ", realEachBlockCostTimeSec)
		currentHashRate := new(big.Int).Div(targetHashWorth, new(big.Int).SetUint64(realEachBlockCostTimeSec))
		currentHashWorth = new(big.Int).Mul(currentHashRate, new(big.Int).SetUint64(mint.EachBlockRequiredTargetTime))
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

	// 缓存
	hashRateChartsV3_lastreqtime = &tn  // 缓存时间
	hashRateChartsV3_cachedata = result // 缓存数据

	// 返回
	return result

}
