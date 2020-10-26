package rpc

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"strings"
)

func hashRateList(blockstore interfaces.BlockStore, curheight uint64, allHistoryOr30Days bool) ([]string, error) {
	var stepNum = 30
	if curheight < mint.AdjustTargetDifficultyNumberOfBlocks*uint64(stepNum) {
		return []string{}, nil
	}
	blockHeadMetaSize := uint32(blocks.BlockHeadSize + blocks.BlockMetaSizeV1)
	// 历史三十天分布目标哈希率
	var allDivCut = big.NewInt(1)
	var allMaxRate = big.NewInt(1)
	var allHistoryRateStrs = make([]string, stepNum)
	var allHistoryRates = make([]*big.Int, stepNum)
	stepBlkHei := mint.AdjustTargetDifficultyNumberOfBlocks
	if allHistoryOr30Days {
		stepBlkHei = curheight/uint64(stepNum) - 1
	}
	for i := 0; i < stepNum; i++ {
		tarhei := curheight - (stepBlkHei * uint64(i))
		_, headbytes, err2 := blockstore.ReadBlockBytesByHeight(tarhei, blockHeadMetaSize)
		if err2 != nil {
			return nil, err2
		}
		if len(headbytes) < int(blockHeadMetaSize) {
			return nil, fmt.Errorf("block data read error")
		}
		blk, _, err3 := blocks.ParseExcludeTransactions(headbytes, 0)
		if err3 != nil {
			return nil, err3
		}
		//fmt.Println(tarhei, blk.GetDifficulty(), hex.EncodeToString(headbytes))
		targetHashWorth := difficulty.CalculateDifficultyWorth(tarhei, blk.GetDifficulty())
		if targetHashWorth.Cmp(allMaxRate) == 1 {
			allMaxRate = targetHashWorth
			allDivCut = new(big.Int).Div(targetHashWorth, big.NewInt(10000))
		}
		allHistoryRates[i] = targetHashWorth
	}
	// 截取计算，倒序
	var idx = 0
	for i := stepNum - 1; i >= 0; i-- {
		rlrt := allHistoryRates[i]
		showrate := new(big.Int).Div(rlrt, allDivCut)
		allHistoryRateStrs[idx] = showrate.String()
		idx++
	}
	return allHistoryRateStrs, nil
}

func (api *DeprecatedApiService) hashRateCharts(params map[string]string) map[string]string {
	result := make(map[string]string)
	lastest, err1 := api.blockchain.State().ReadLastestBlockHeadAndMeta()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}
	curheight := lastest.GetHeight()
	blockstore := api.blockchain.State().BlockStore()

	allHistory, e1 := hashRateList(blockstore, curheight, true)
	if e1 != nil {
		result["err"] = e1.Error()
		return result
	}
	days30, e2 := hashRateList(blockstore, curheight, false)
	if e2 != nil {
		result["err"] = e1.Error()
		return result
	}

	// ok
	result["jsondata"] = `{"historys":[` + strings.Join(allHistory, ",") + `],"days30":[` + strings.Join(days30, ",") + `]}`

	// 返回
	return result

}
