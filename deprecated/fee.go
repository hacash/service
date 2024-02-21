package rpc

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"strconv"
	"strings"
)

func omitMoreDecimalForFloatString(number string) string {
	var pt = strings.IndexAny(number, ".")
	if pt == -1 {
		// not float with decimal
		return number
	}
	var n = len(number)
	var k = n
	for i := pt + 1; i < n; i++ {
		if number[i-1] != '0' && number[i] != '0' {
			k = i + 1
			break
		}
	}
	return number[0:k]
}

func (api *DeprecatedApiService) getLatestAverageFeePurity(params map[string]string) map[string]string {
	result := make(map[string]string)

	_, isunitmei := params["unitmei"]
	_, isomitdec := params["omitdec"]

	txszp, ok1 := params["txsize"]
	txsz := uint64(0)
	if ok1 {
		txsz, _ = strconv.ParseUint(txszp, 10, 64)
	}

	var kernel = api.blockchain.GetChainEngineKernel()
	//var feepb = uint64(kernel.GetLatestAverageFeePurity())
	//fmt.Println(feepb, "avgf feepb -------------")
	var feert = getLatestAverageFeePurityData(kernel, isunitmei, uint32(txsz))

	if txsz > 0 {
		var fpn string
		if isunitmei {
			fpn = fmt.Sprintf("%f", feert) // mei
		} else {
			fpn = strconv.FormatUint(uint64(feert), 10) // zhu
		}
		//fpn = "17"+fpn+"72656897"
		//fpn = "15.01"
		if isomitdec {
			fpn = omitMoreDecimalForFloatString(fpn)
		}
		result["feasible_fee"] = fpn

	} else {
		fpn := strconv.FormatUint(uint64(feert), 10)
		result["fee_purity"] = fpn
	}

	// ok
	return result

}

func getLatestAverageFeePurityData(kernel interfaces.ChainEngine, isunitmei bool, txsz uint32) float64 {

	var feepb = uint64(kernel.GetLatestAverageFeePurity())
	//fmt.Println(feepb, "avgf feepb -------------")

	if txsz > 0 {
		fp := feepb * (uint64(txsz)/32 + 1)
		if isunitmei {
			return float64(fp) / 10000_0000
			//fpn = fmt.Sprintf("%f", fpf)// mei
		} else {
			//fpn = strconv.FormatUint(fp, 10) // zhu
			return float64(fp)
		}
	} else {
		return float64(feepb)
	}
}
