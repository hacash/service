package rpc

import "strconv"

func (api *DeprecatedApiService) getLatestAverageFeePurity(params map[string]string) map[string]string {
	result := make(map[string]string)

	txszp, ok1 := params["tx_size"]
	txsz := uint64(0)
	if ok1 {
		txsz, _ = strconv.ParseUint(txszp, 10, 64)
	}

	var kernel = api.blockchain.GetChainEngineKernel()
	var feepb = uint64(kernel.GetLatestAverageFeePurity())

	if txsz > 0 {
		fp := feepb * (txsz/32 + 1)
		fpn := strconv.FormatUint(fp, 10)
		result["feasible_fee"] = fpn

	} else {
		fpn := strconv.FormatUint(feepb, 10)
		result["fee_purity"] = fpn
	}

	// ok
	return result

}
