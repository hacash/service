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

func hashRateList(blockstore interfaces.BlockStoreRead, curheight uint64, allHistoryOr300Days bool, appendItem *big.Int) ([]string, error) {
	var stepNum = 300
	// 第一个附加
	var hsti = 0
	if appendItem != nil {
		hsti = 1
	}
	if curheight < mint.AdjustTargetDifficultyNumberOfBlocks*uint64(stepNum) {
		return []string{}, nil
	}
	blockHeadMetaSize := uint32(blocks.BlockHeadSize + blocks.BlockMetaSizeV1)
	// 历史三十天分布目标哈希率
	var allDivCut = big.NewInt(1)
	var allMaxRate = big.NewInt(1)
	var allHistoryRateStrs = make([]string, stepNum+hsti)
	var allHistoryRates = make([]*big.Int, stepNum+hsti)
	stepBlkHei := mint.AdjustTargetDifficultyNumberOfBlocks
	if allHistoryOr300Days {
		stepBlkHei = curheight/uint64(stepNum) - 1
	}

	// 第一个附加
	if appendItem != nil {
		allMaxRate = appendItem
		allHistoryRates[0] = appendItem
	}

	// 读取的
	for i := 0; i < stepNum; i++ {
		tarhei := curheight - (stepBlkHei * uint64(i))
		//_, headbytes, err2 := blockstore.ReadBlockBytesLengthByHeight(tarhei, blockHeadMetaSize)
		_, headbytes, err2 := blockstore.ReadBlockBytesByHeight(tarhei)
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
		//fmt.Println(tarhei, blk.GetDifficulty(), hex.EncodeToString(difficulty.DifficultyUint32ToHash(blk.GetDifficulty())))
		targetHashWorth := difficulty.CalculateDifficultyWorth(blk.GetDifficulty())
		if targetHashWorth.Cmp(allMaxRate) == 1 {
			allMaxRate = targetHashWorth
			allDivCut = new(big.Int).Div(targetHashWorth, big.NewInt(10000))
		}
		allHistoryRates[i+hsti] = targetHashWorth
	}

	// 截取计算，倒序
	var idx = 0
	for i := stepNum + hsti - 1; i >= 0; i-- {
		rlrt := allHistoryRates[i]
		//fmt.Println(rlrt)
		showrate := new(big.Int).Div(rlrt, allDivCut)
		allHistoryRateStrs[idx] = showrate.String()
		idx++
	}
	return allHistoryRateStrs, nil
}

func (api *DeprecatedApiService) hashRateCharts(params map[string]string) map[string]string {
	result := make(map[string]string)
	lastest, err1 := api.blockchain.StateRead().ReadLastestBlockHeadMetaForRead()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}
	curheight := lastest.GetHeight()
	blockstore := api.blockchain.StateRead().BlockStoreRead()

	allHistory, e1 := hashRateList(blockstore, curheight, true, nil)
	if e1 != nil {
		result["err"] = e1.Error()
		return result
	}
	days30, e2 := hashRateList(blockstore, curheight, false, nil)
	if e2 != nil {
		result["err"] = e2.Error()
		return result
	}

	// ok
	result["jsondata"] = `{"historys":[` + strings.Join(allHistory, ",") + `],"days30":[` + strings.Join(days30, ",") + `]}`

	// 返回
	return result

}
