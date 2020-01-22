package rpc

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
	"math/big"
)


func (api *DeprecatedApiService) powPower(params map[string]string) map[string]string {
	result := make(map[string]string)
	lastest, err1 := api.blockchain.State().ReadLastestBlockHeadAndMeta()
	if err1 != nil {
		result["err"] = err1.Error()
		return result
	}
	curheight := lastest.GetHeight()
	mint_num288dj := uint64(mint.AdjustTargetDifficultyNumberOfBlocks / 4)
	prev288height := curheight - uint64(mint_num288dj)
	headbytes, err2 := api.blockchain.State().BlockStore().ReadBlockHeadBytesByHeight( prev288height )
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
	//fmt.Println(cost288sec)
	// cost time
	powbitshash := difficulty.DifficultyUint32ToHash( lastest.GetDifficulty() )
	powbitsbig := difficulty.CalculateHashWorth( powbitshash )
	powervalue := new(big.Int).Div( powbitsbig, big.NewInt( int64(cost288sec) ) )
	//fmt.Println( mint_num288dj, cost288sec, powbitsbig.String(), powervalue.String() )
	// return
	result["power"] = powervalue.String()
	result["show"] = difficulty.ConvertPowPowerToShowFormat( powervalue )
	return result
}
