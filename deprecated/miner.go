package rpc

import (
	"fmt"
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
	mint_num288dj := uint64(mint.AdjustTargetDifficultyNumberOfBlocks)
	prev288height := uint64(curheight) / mint_num288dj * mint_num288dj
	num288 := uint64(curheight) - prev288height
	cost288sec := api.getMiao(lastest, prev288height, num288)
	fmt.Println(cost288sec)
	// cost time
	powbitsbig := difficulty.DifficultyUint32ToBig( lastest.GetDifficulty() )
	totalbig := new(big.Int).Mul( powbitsbig, big.NewInt( int64(num288) ) )
	powervalue := new(big.Int).Div( totalbig, big.NewInt( int64(cost288sec) ) )
	//fmt.Println( powbitsbig.String() )
	//fmt.Println( powervalue.String() )
	// return
	result["power"] = powervalue.String()
	result["show"] = difficulty.ConvertPowPowerToShowFormat( powervalue )
	return result
}
