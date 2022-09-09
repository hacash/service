package rpc

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
	"math/big"
)

func (api *DeprecatedApiService) powPower(params map[string]string) map[string]string {
	result := make(map[string]string)
	lastest, _, err1 := api.blockchain.GetChainEngineKernel().LatestBlock()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}

	curheight := lastest.GetHeight()
	mint_num288dj := uint64(mint.AdjustTargetDifficultyNumberOfBlocks / 4)
	prev288height := curheight - uint64(mint_num288dj)
	_, headbytes, err2 := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead().ReadBlockBytesByHeight(prev288height)
	if err2 != nil {
		result["err"] = err2.Error()
		return result
	}

	blk, _, err3 := blocks.ParseBlockHead(headbytes, 0)
	if err3 != nil {
		result["err"] = err3.Error()
		return result
	}

	cost288sec := lastest.GetTimestamp() - blk.GetTimestamp()
	cost288sec = cost288sec / mint_num288dj
	if cost288sec == 0 {
		cost288sec = 1
	}

	// cost time
	powbitsbig := difficulty.CalculateDifficultyWorthByHeight(lastest.GetHeight(), lastest.GetDifficulty())
	powervalue := new(big.Int).Div(powbitsbig, big.NewInt(int64(cost288sec)))

	// return
	result["power"] = powervalue.String()
	result["show"] = difficulty.ConvertPowPowerToShowFormat(powervalue)
	return result
}
