package rpc

import (
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
	prev288height := uint64(curheight) - mint_num288dj
	cost288sec := api.getMiao(lastest, prev288height, mint_num288dj)
	if cost288sec == 0 {
		cost288sec = 1
	}
	//fmt.Println(cost288sec)
	// cost time
	powbitshash := difficulty.DifficultyUint32ToHash( lastest.GetDifficulty() )
	powbitsbig := difficulty.CalculateHashWorth( powbitshash )
	powervalue := new(big.Int).Mul( powbitsbig, big.NewInt( int64(mint_num288dj) ) )
	powervalue =  new(big.Int).Div( powervalue, big.NewInt( int64(cost288sec) ) )
	//fmt.Println( mint_num288dj, cost288sec, powbitsbig.String(), powervalue.String() )
	// return
	result["power"] = powervalue.String()
	result["show"] = difficulty.ConvertPowPowerToShowFormat( powervalue )
	return result
}
